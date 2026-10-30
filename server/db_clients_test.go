// SPDX-License-Identifier: AGPL-3.0-or-later
package main

// What a client of the public API needs beyond the app: its own API keys,
// signing in with them, editing sightings and the account.

import (
	"net/http"
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
	s.must(t, owner, http.StatusNoContent, "POST", "/v1/lists/"+list+"/visibility", map[string]any{"visibility": models.VisibilityHidden})
	s.must(t, friend, http.StatusOK, "GET", "/v1/lists/"+list, nil)
	s.must(t, owner, http.StatusNoContent, "POST", "/v1/lists/"+list+"/visibility", map[string]any{"visibility": models.VisibilityPrivate})
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
