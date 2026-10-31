// SPDX-License-Identifier: AGPL-3.0-or-later
package main

// What a client of the public API needs beyond the app: its own API keys,
// signing in with them, editing sightings and the account.

import (
	"mime"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/utils"
)

func TestDBOwnApiKey(t *testing.T) {
	s := testDB(t)
	u := s.signUp(t)

	r := s.must(t, u, http.StatusOK, "POST", "/v1/api_keys", map[string]any{"name": "My client"})
	clientId, secret := r.Body["client_id"].(string), r.Body["api_key"].(string)
	if clientId == "" || r.Body["scopes"] != "user" || r.Body["prefix"] != secret[:7] {
		t.Fatalf("new key = %s", r.Raw)
	}

	// The key signs in by itself, and without asking for scopes gets the user scope.
	signIn := s.do(t, u.address, "", "POST", "/v1/auth", map[string]any{
		"grant_type": "password", "client_id": clientId, "client_secret": secret, "email": u.Email, "password": u.Password,
	})
	if signIn.Status != http.StatusOK || signIn.Body["scope"] != "user" {
		t.Fatalf("sign in with own key = %d: %s", signIn.Status, signIn.Raw)
	}
	s.must(t, &testUser{address: u.address, Token: signIn.Body["access_token"].(string)}, http.StatusOK, "GET", "/v1/users/whoami", nil)

	// A key can't have scopes its owner lacks.
	if r := s.call(t, u, "POST", "/v1/api_keys", map[string]any{"name": "x", "scopes": []string{"user", "admin"}}); r.Status != http.StatusBadRequest || get(r.Body, "code") != float64(3005) {
		t.Errorf("non-admin made an admin key: %d %s", r.Status, r.Raw)
	}
	admin := s.signUpAdmin(t)
	r = s.must(t, admin, http.StatusOK, "POST", "/v1/api_keys", map[string]any{"name": "Admin client", "scopes": []string{"admin", "user", "admin"}})
	if r.Body["scopes"] != "admin user" {
		t.Errorf("admin key scopes = %v", r.Body["scopes"])
	}
}

// A read-only key's sessions can read but not change anything.
func TestDBReadOnlyApiKey(t *testing.T) {
	s := testDB(t)
	u := s.signUp(t)

	if r := s.call(t, u, "POST", "/v1/api_keys", map[string]any{"name": "x", "scopes": []string{"read", "user"}}); r.Status != http.StatusBadRequest || get(r.Body, "code") != float64(3005) {
		t.Errorf("read-only key with the user scope: %d %s", r.Status, r.Raw)
	}
	r := s.must(t, u, http.StatusOK, "POST", "/v1/api_keys", map[string]any{"name": "Reader", "scopes": []string{"read"}})
	if r.Body["scopes"] != "read" {
		t.Fatalf("read-only key = %s", r.Raw)
	}
	clientId, secret := r.Body["client_id"].(string), r.Body["api_key"].(string)

	// Without asking for scopes it gets read; it can't get user.
	signIn := func(scopes ...string) response {
		body := map[string]any{"grant_type": "password", "client_id": clientId, "client_secret": secret, "email": u.Email, "password": u.Password}
		if scopes != nil {
			body["scopes"] = scopes
		}
		return s.do(t, u.address, "", "POST", "/v1/auth", body)
	}
	if r := signIn("user"); r.Status != http.StatusBadRequest {
		t.Errorf("read-only key signed in with user: %d %s", r.Status, r.Raw)
	}
	r = signIn()
	if r.Status != http.StatusOK || r.Body["scope"] != "read" {
		t.Fatalf("sign in with read-only key = %d: %s", r.Status, r.Raw)
	}
	reader := &testUser{address: u.address, Token: r.Body["access_token"].(string)}
	s.must(t, reader, http.StatusOK, "GET", "/v1/users/whoami", nil)
	s.must(t, reader, http.StatusOK, "GET", "/v1/lists/list", nil)
	s.must(t, reader, http.StatusUnauthorized, "POST", "/v1/users/display_name", map[string]any{"display_name": "Changed"})
	s.must(t, reader, http.StatusUnauthorized, "POST", "/v1/api_keys", map[string]any{"name": "more"})
}

// A new account gets its first key with its password, without the website.
func TestDBApiKeyWithPassword(t *testing.T) {
	s := testDB(t)
	u := s.signUpAdmin(t)

	r := s.do(t, u.address, "", "POST", "/v1/api_keys/password", map[string]any{"email": u.Email, "password": u.Password, "name": "Fågelbok"})
	if r.Status != http.StatusOK || r.Body["scopes"] != "user" || r.Body["client_id"] == "" {
		t.Fatalf("key with password = %d: %s", r.Status, r.Raw)
	}
	clientId, secret := r.Body["client_id"].(string), r.Body["api_key"].(string)

	// It signs in, and only ever with the user scope, even for an admin.
	signIn := s.do(t, u.address, "", "POST", "/v1/auth", map[string]any{
		"grant_type": "password", "client_id": clientId, "client_secret": secret,
		"email": u.Email, "password": u.Password, "scopes": []string{"user", "admin"},
	})
	if signIn.Status != http.StatusOK || signIn.Body["scope"] != "user" {
		t.Fatalf("sign in with the new key = %d: %s", signIn.Status, signIn.Raw)
	}
	keys := items(s.must(t, u, http.StatusOK, "GET", "/v1/api_keys/list", nil).Body["api_keys"])
	if len(keys) != 1 || get(keys[0], "name") != "Fågelbok" {
		t.Errorf("keys = %v", keys)
	}

	// A wrong password makes no key and counts towards the guessing limit.
	wrong := s.do(t, u.address, "", "POST", "/v1/api_keys/password", map[string]any{"email": u.Email, "password": "guess"})
	if wrong.Status != http.StatusBadRequest || get(wrong.Body, "code") != float64(7001) {
		t.Errorf("wrong password = %d: %s", wrong.Status, wrong.Raw)
	}
	unknown := s.do(t, u.address, "", "POST", "/v1/api_keys/password", map[string]any{"email": "nobody@example.com", "password": "guess"})
	if unknown.Status != http.StatusBadRequest || get(unknown.Body, "code") != float64(7001) {
		t.Errorf("unknown email = %d: %s", unknown.Status, unknown.Raw)
	}

	// Throttled per network.
	other := s.signUp(t)
	for i := 0; i < 10; i++ {
		s.do(t, other.address, "", "POST", "/v1/api_keys/password", map[string]any{"email": "nobody@example.com", "password": "guess"})
	}
	if r := s.do(t, other.address, "", "POST", "/v1/api_keys/password", map[string]any{"email": other.Email, "password": other.Password}); r.Status != http.StatusTooManyRequests {
		t.Errorf("11th call from one network = %d: %s", r.Status, r.Raw)
	}
}

func TestDBEditSighting(t *testing.T) {
	s := testDB(t)
	owner, stranger, admin := s.signUp(t), s.signUp(t), s.signUpAdmin(t)
	species := someSpecies(t, 2)
	gender := s.must(t, admin, http.StatusOK, "POST", "/v1/genders", map[string]any{"name": map[string]any{"en": "Female"}}).Body["uuid"].(string)
	tag := s.must(t, owner, http.StatusOK, "POST", "/v1/tags", map[string]any{"name": "Garden"}).Body["uuid"].(string)
	first, second := newList(t, s, owner, "First", models.VisibilityPrivate), newList(t, s, owner, "Second", models.VisibilityPrivate)

	// Reporting answers with the sighting as it is shown.
	created := s.must(t, owner, http.StatusOK, "POST", "/v1/reports", map[string]any{
		"uuid": utils.GenerateUuidWithFixedLength(20), "species_uuid": species[0], "from": "2026-05-01T06:30:00Z",
		"gender_uuid": gender, "comment": "Singing", "list_uuids": []string{first}, "quantity": 2,
	})
	uuid := created.Body["uuid"].(string)
	if get(created.Body, "species", "uuid") != species[0] || get(created.Body, "species", "name", "en") == "" || get(created.Body, "gender", "uuid") != gender ||
		get(created.Body, "comment", "text") != "Singing" || get(created.Body, "lists", 0, "name", "en") != "First" {
		t.Errorf("created report = %s", created.Raw)
	}
	details := s.must(t, owner, http.StatusOK, "GET", "/v1/reports/"+uuid, nil)
	if get(details.Body, "result", "gender", "uuid") != gender || get(details.Body, "result", "lists", 0, "name", "en") != "First" {
		t.Errorf("details = %s", details.Raw)
	}

	// Editing replaces every field.
	edit := map[string]any{
		"species_uuid": species[1], "from": "2026-05-02T07:00:00Z", "quantity": 5, "site": "Pond",
		"comment": "Two pairs", "list_uuids": []string{second}, "tag_uuids": []string{tag},
	}
	edited := s.must(t, owner, http.StatusOK, "PUT", "/v1/reports/"+uuid, edit)
	if get(edited.Body, "uuid") != uuid || get(edited.Body, "species", "uuid") != species[1] || get(edited.Body, "quantity") != float64(5) ||
		get(edited.Body, "site") != "Pond" || get(edited.Body, "gender") != nil || get(edited.Body, "comment", "text") != "Two pairs" ||
		len(items(edited.Body["lists"])) != 1 || get(edited.Body, "lists", 0, "uuid") != second || get(edited.Body, "tags", 0, "uuid") != tag {
		t.Errorf("edited report = %s", edited.Raw)
	}
	if n := len(items(s.must(t, owner, http.StatusOK, "GET", "/v1/lists/"+first+"/reports", nil).Body["result"])); n != 0 {
		t.Errorf("the first list still has %d sightings", n)
	}
	if n := len(items(s.must(t, owner, http.StatusOK, "GET", "/v1/lists/"+second+"/reports", nil).Body["result"])); n != 1 {
		t.Errorf("the second list has %d sightings, want 1", n)
	}

	// Only the owner edits, and only sightings that are not archived.
	s.must(t, stranger, http.StatusBadRequest, "PUT", "/v1/reports/"+uuid, edit)
	s.must(t, owner, http.StatusBadRequest, "PUT", "/v1/reports/"+uuid, map[string]any{"species_uuid": "nope"})
	s.must(t, owner, http.StatusNoContent, "POST", "/v1/reports/"+uuid+"/archive", nil)
	s.must(t, owner, http.StatusBadRequest, "PUT", "/v1/reports/"+uuid, edit)
}

func TestDBListDatesAndVisibility(t *testing.T) {
	s := testDB(t)
	owner, friend := s.signUp(t), s.signUp(t)

	r := s.must(t, owner, http.StatusOK, "POST", "/v1/lists", map[string]any{
		"name": map[string]any{"en": "Year"}, "from": "2026-01-01", "until": "2026-12-31T23:00:00Z",
	})
	list := r.Body["uuid"].(string)
	if from, _ := time.Parse(time.RFC3339, r.Body["from"].(string)); from.Year() != 2026 {
		t.Errorf("from = %v", r.Body["from"])
	}
	s.must(t, owner, http.StatusBadRequest, "POST", "/v1/lists", map[string]any{"name": map[string]any{"en": "Bad"}, "from": "01/01/2026"})

	s.must(t, owner, http.StatusNoContent, "POST", "/v1/lists/"+list+"/timespan", map[string]any{"from": "*", "until": "2027-06-30"})
	s.must(t, owner, http.StatusBadRequest, "POST", "/v1/lists/"+list+"/timespan", map[string]any{"from": "tomorrow"})
	var stored models.List
	models.DB.Where("uuid=?", list).First(&stored)
	if stored.From.Valid || stored.Until.Time.Year() != 2027 {
		t.Errorf("timespan = %v – %v", stored.From, stored.Until)
	}

	// A private list becomes visible to others with the link, and back.
	s.must(t, friend, http.StatusBadRequest, "GET", "/v1/lists/"+list, nil)
	s.must(t, owner, http.StatusOK, "POST", "/v1/lists/"+list+"/visibility", map[string]any{"visibility": models.VisibilityHidden})
	s.must(t, friend, http.StatusOK, "GET", "/v1/lists/"+list, nil)
	s.must(t, owner, http.StatusOK, "POST", "/v1/lists/"+list+"/visibility", map[string]any{"visibility": models.VisibilityPrivate})
	s.must(t, friend, http.StatusBadRequest, "GET", "/v1/lists/"+list, nil)
	s.must(t, owner, http.StatusBadRequest, "POST", "/v1/lists/"+list+"/visibility", map[string]any{"visibility": 3})
	s.must(t, friend, http.StatusBadRequest, "POST", "/v1/lists/"+list+"/visibility", map[string]any{"visibility": 2})
}

func TestDBChangeAccount(t *testing.T) {
	s := testDB(t)
	u, other := s.signUp(t), s.signUp(t)

	s.must(t, u, http.StatusNoContent, "POST", "/v1/users/display_name", map[string]any{"display_name": "  Bird Person "})
	s.must(t, u, http.StatusBadRequest, "POST", "/v1/users/display_name", map[string]any{"display_name": "<b>"})
	if r := s.must(t, u, http.StatusOK, "GET", "/v1/users/whoami", nil); get(r.Body, "data", "display_name") != "Bird Person" {
		t.Errorf("display name = %v", get(r.Body, "data", "display_name"))
	}

	newName := "renamed" + utils.GenerateUuidWithFixedLength(8)
	s.must(t, u, http.StatusNoContent, "POST", "/v1/users/username", map[string]any{"username": newName})
	s.must(t, u, http.StatusNoContent, "POST", "/v1/users/username", map[string]any{"username": newName})
	if r := s.must(t, other, http.StatusConflict, "POST", "/v1/users/username", map[string]any{"username": newName}); get(r.Body, "code") != float64(4006) {
		t.Errorf("taken username = %s", r.Raw)
	}
	s.must(t, u, http.StatusBadRequest, "POST", "/v1/users/username", map[string]any{"username": "a b"})
	signUp := s.do(t, "10.251.0.1", "", "POST", "/v1/users", map[string]any{
		"email": "x-" + newName + "@example.com", "display_name": "Someone", "username": newName, "password": "long-enough-1", "accepted_terms": true,
	})
	if signUp.Status != http.StatusConflict || get(signUp.Body, "code") != float64(4006) {
		t.Errorf("sign up with a taken username = %d %s", signUp.Status, signUp.Raw)
	}

	// The email takes the password, can't be someone else's, and the old address is told.
	newEmail := "moved-" + utils.GenerateUuidWithFixedLength(8) + "@example.com"
	if r := s.call(t, u, "POST", "/v1/users/email", map[string]any{"email": newEmail, "password": "wrong"}); r.Status != http.StatusBadRequest || get(r.Body, "code") != float64(-30002) {
		t.Errorf("wrong password = %d %s", r.Status, r.Raw)
	}
	if r := s.call(t, u, "POST", "/v1/users/email", map[string]any{"email": other.Email, "password": u.Password}); r.Status != http.StatusBadRequest || get(r.Body, "code") != float64(4002) {
		t.Errorf("taken email = %d %s", r.Status, r.Raw)
	}
	oldEmail := u.Email
	s.must(t, u, http.StatusNoContent, "POST", "/v1/users/email", map[string]any{"email": newEmail, "password": u.Password})
	var notices int64
	models.DB.Model(&models.Email{}).Where("to_email=?", oldEmail).Where("text_part LIKE ?", "%"+newEmail+"%").Count(&notices)
	if notices != 1 {
		t.Errorf("%d notices to the old address, want 1", notices)
	}
	if r := s.signIn(t, u, "user"); r.Status == http.StatusOK {
		t.Errorf("signed in with the old email")
	}
	u.Email = newEmail
	if r := s.signIn(t, u, "user"); r.Status != http.StatusOK {
		t.Errorf("sign in with the new email = %d %s", r.Status, r.Raw)
	}
}

// A sighting goes on and off checklists without resending the whole of it.
func TestDBSightingOnAndOffLists(t *testing.T) {
	s := testDB(t)
	owner, stranger := s.signUp(t), s.signUp(t)
	species := someSpecies(t, 1)
	first, second := newList(t, s, owner, "First", models.VisibilityPrivate), newList(t, s, owner, "Second", models.VisibilityPrivate)
	theirs := newList(t, s, stranger, "Theirs", models.VisibilityPrivate)
	uuid := newReport(t, s, owner, species[0], first).Body["uuid"].(string)
	lists := func() []string {
		var out []string
		for _, l := range items(get(s.must(t, owner, http.StatusOK, "GET", "/v1/reports/"+uuid, nil).Body, "result", "lists")) {
			out = append(out, get(l, "uuid").(string))
		}
		return out
	}

	s.must(t, owner, http.StatusNoContent, "POST", "/v1/reports/"+uuid+"/add_list", map[string]any{"list_uuid": second})
	if got := lists(); len(got) != 2 {
		t.Errorf("after add: %v", got)
	}
	if r := s.call(t, owner, "POST", "/v1/reports/"+uuid+"/add_list", map[string]any{"list_uuid": second}); r.Status != http.StatusBadRequest {
		t.Errorf("adding twice = %d: %s", r.Status, r.Raw)
	}
	s.must(t, owner, http.StatusNoContent, "POST", "/v1/reports/"+uuid+"/remove_list", map[string]any{"list_uuid": first})
	if got := lists(); len(got) != 1 || got[0] != second {
		t.Errorf("after moving: %v", got)
	}
	// Nothing else about the sighting changed.
	if got := get(s.must(t, owner, http.StatusOK, "GET", "/v1/reports/"+uuid, nil).Body, "result", "from"); got == nil || strings.HasPrefix(got.(string), "0001") {
		t.Errorf("from = %v", got)
	}

	// Only your own sightings and lists.
	for _, c := range []struct {
		u    *testUser
		list string
	}{{stranger, theirs}, {owner, theirs}, {owner, "nope"}} {
		if r := s.call(t, c.u, "POST", "/v1/reports/"+uuid+"/add_list", map[string]any{"list_uuid": c.list}); r.Status != http.StatusBadRequest || get(r.Body, "code") != float64(1001) {
			t.Errorf("add_list %s = %d: %s", c.list, r.Status, r.Raw)
		}
	}
}

// A bare date is a date-only sighting, stored as midnight UTC.
func TestDBSightingWithBareDate(t *testing.T) {
	s := testDB(t)
	u := s.signUp(t)
	species := someSpecies(t, 1)
	r := s.must(t, u, http.StatusOK, "POST", "/v1/reports", map[string]any{
		"uuid": utils.GenerateUuidWithFixedLength(20), "species_uuid": species[0], "from": "2026-03-01", "until": "2026-03-02",
	})
	for _, field := range []string{"from", "until"} {
		got, err := time.Parse(time.RFC3339, r.Body[field].(string))
		want := map[string]string{"from": "2026-03-01", "until": "2026-03-02"}[field]
		if err != nil || got.UTC().Format(time.RFC3339) != want+"T00:00:00Z" {
			t.Errorf("%s = %v", field, r.Body[field])
		}
	}
	if r := s.call(t, u, "POST", "/v1/reports", map[string]any{"species_uuid": species[0], "from": "2026-13-01"}); r.Status != http.StatusBadRequest {
		t.Errorf("bad date = %d", r.Status)
	}
	if r := s.call(t, u, "POST", "/v1/reports", map[string]any{"species_uuid": species[0], "from": "yesterday"}); r.Status != http.StatusBadRequest {
		t.Errorf("not a date = %d", r.Status)
	}
}

// The CSV export is named after its list.
func TestDBExportCsvIsNamed(t *testing.T) {
	s := testDB(t)
	u := s.signUp(t)
	list := newList(t, s, u, "Årslista 2026", models.VisibilityPrivate)
	req := httptest.NewRequest("PUT", "/v1/reports/export", strings.NewReader(`{"list_uuid":"`+list+`"}`))
	req.RemoteAddr = u.address + ":40000"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+u.Token)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("export = %d: %s", w.Code, w.Body.String())
	}
	_, params, err := mime.ParseMediaType(w.Header().Get("Content-Disposition"))
	if err != nil || params["filename"] != "Årslista 2026.csv" {
		t.Errorf("Content-Disposition = %q (%v)", w.Header().Get("Content-Disposition"), err)
	}
	if got := w.Header().Get("Access-Control-Expose-Headers"); !strings.Contains(got, "Content-Disposition") {
		t.Errorf("Access-Control-Expose-Headers = %q", got)
	}
}

// Labels have a name and a #rrggbb color, and an archived one can't be added.
func TestDBTagsAreChecked(t *testing.T) {
	s := testDB(t)
	u := s.signUp(t)
	for _, body := range []map[string]any{{"name": " "}, {"name": "x", "color": "red"}} {
		if r := s.call(t, u, "POST", "/v1/tags", body); r.Status != http.StatusBadRequest || get(r.Body, "code") != float64(1005) {
			t.Errorf("tag %v = %d: %s", body, r.Status, r.Raw)
		}
	}
	tag := s.must(t, u, http.StatusOK, "POST", "/v1/tags", map[string]any{"name": " Garden ", "color": "#1273a3"}).Body
	if tag["name"] != "Garden" {
		t.Errorf("name = %q", tag["name"])
	}
	s.must(t, u, http.StatusBadRequest, "POST", "/v1/tags/"+tag["uuid"].(string)+"/rename", map[string]any{"name": ""})

	species := someSpecies(t, 1)
	report := s.must(t, u, http.StatusOK, "POST", "/v1/reports", map[string]any{"species_uuid": species[0]}).Body["uuid"].(string)
	add := func(tagUuid string) response {
		return s.call(t, u, "POST", "/v1/reports/"+report+"/add_tag", map[string]any{"tag_uuid": tagUuid})
	}
	if r := add(tag["uuid"].(string)); r.Status != http.StatusNoContent {
		t.Fatalf("add_tag = %d: %s", r.Status, r.Raw)
	}
	if r := add(tag["uuid"].(string)); r.Status != http.StatusBadRequest || get(r.Body, "code") != float64(1002) {
		t.Errorf("adding it again = %d: %s", r.Status, r.Raw)
	}
	old := s.must(t, u, http.StatusOK, "POST", "/v1/tags", map[string]any{"name": "Old"}).Body["uuid"].(string)
	s.must(t, u, http.StatusNoContent, "POST", "/v1/tags/"+old+"/archive", nil)
	if r := add(old); r.Status != http.StatusBadRequest || get(r.Body, "code") != float64(1001) {
		t.Errorf("adding an archived tag = %d: %s", r.Status, r.Raw)
	}
}

// Impossible sightings and text too long for its column are the client's mistake: 400, not 500.
func TestDBSightingIsChecked(t *testing.T) {
	s := testDB(t)
	u := s.signUp(t)
	species := someSpecies(t, 1)
	for _, extra := range []map[string]any{
		{"quantity": -1},
		{"latitude": 999, "longitude": 18},
		{"from": "2026-03-02", "until": "2026-03-01"},
	} {
		body := map[string]any{"species_uuid": species[0]}
		for k, v := range extra {
			body[k] = v
		}
		if r := s.call(t, u, "POST", "/v1/reports", body); r.Status != http.StatusBadRequest {
			t.Errorf("sighting with %v = %d: %s", extra, r.Status, r.Raw)
		}
	}
	r := s.call(t, u, "POST", "/v1/reports", map[string]any{"species_uuid": species[0], "site": strings.Repeat("x", 300)})
	if r.Status != http.StatusBadRequest || get(r.Body, "code") != float64(2001) || r.Body["message"] != "site is too long" {
		t.Errorf("long site = %d: %s", r.Status, r.Raw)
	}
	r = s.call(t, u, "POST", "/v1/tags", map[string]any{"name": strings.Repeat("x", 65)})
	if r.Status != http.StatusBadRequest || r.Body["message"] != "name is too long" {
		t.Errorf("long tag name = %d: %s", r.Status, r.Raw)
	}
}
