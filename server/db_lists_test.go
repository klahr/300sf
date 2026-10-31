// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/klahr/300sf/server/controllers"
	"github.com/klahr/300sf/server/live"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/utils"
	"gopkg.in/guregu/null.v4/zero"
)

// someSpecies is a species from the starter data.
func someSpecies(t *testing.T, n int) []string {
	t.Helper()
	var out []string
	if err := models.DB.Table("species").Order("id").Limit(n).Pluck("uuid", &out).Error; err != nil || len(out) < n {
		t.Fatalf("species: %v (%d)", err, len(out))
	}
	return out
}

func newList(t *testing.T, s *dbServer, u *testUser, name string, visibility int64) string {
	t.Helper()
	r := s.must(t, u, http.StatusOK, "POST", "/v1/lists", map[string]any{
		"name": map[string]any{"en": name, "sv": name}, "visibility": visibility, "color": "#1273a3",
	})
	return r.Body["uuid"].(string)
}

func newReport(t *testing.T, s *dbServer, u *testUser, species string, lists ...string) response {
	t.Helper()
	return s.call(t, u, "POST", "/v1/reports", map[string]any{
		"uuid": utils.GenerateUuidWithFixedLength(20), "species_uuid": species,
		"from": time.Now().UTC().Format(time.RFC3339), "list_uuids": lists,
	})
}

func TestDBListsAndSightingsStayPrivate(t *testing.T) {
	s := testDB(t)
	owner, stranger := s.signUp(t), s.signUp(t)
	species := someSpecies(t, 2)

	private := newList(t, s, owner, "Garden", models.VisibilityPrivate)
	if r := newReport(t, s, owner, species[0], private); r.Status != http.StatusOK {
		t.Fatalf("report = %d: %s", r.Status, r.Raw)
	}

	if got := uuids(s.must(t, owner, http.StatusOK, "GET", "/v1/lists/list", nil).Body["result"]); !contains(got, private) {
		t.Errorf("owner's lists %v lack %s", got, private)
	}
	reports := s.must(t, owner, http.StatusOK, "GET", "/v1/lists/"+private+"/reports", nil)
	if n := len(items(reports.Body["result"])); n != 1 {
		t.Errorf("list has %d sightings, want 1", n)
	}

	// Nobody else sees the list, its sightings, or can add to it.
	if r := s.call(t, stranger, "GET", "/v1/lists/"+private, nil); r.Status == http.StatusOK {
		t.Errorf("stranger read private list: %s", r.Raw)
	}
	if r := s.call(t, stranger, "GET", "/v1/lists/"+private+"/reports", nil); r.Status == http.StatusOK && len(items(r.Body["result"])) > 0 {
		t.Errorf("stranger read private sightings: %s", r.Raw)
	}
	newReport(t, s, stranger, species[1], private)
	reports = s.must(t, owner, http.StatusOK, "GET", "/v1/lists/"+private+"/reports", nil)
	if n := len(items(reports.Body["result"])); n != 1 {
		t.Errorf("a stranger's sighting landed on the owner's list (%d sightings)", n)
	}
	if got := len(items(s.must(t, stranger, http.StatusOK, "GET", "/v1/reports/list", nil).Body["result"])); got != 1 {
		t.Errorf("stranger has %d sightings, want only their own one", got)
	}
	if r := s.do(t, "10.250.1.1", "", "POST", "/v1/lists/compare", map[string]any{"uuids": []string{private}}); r.Status == http.StatusOK {
		t.Errorf("compare showed a private list: %s", r.Raw)
	}
	s.must(t, stranger, http.StatusBadRequest, "POST", "/v1/lists/"+private+"/rename", map[string]any{"name": map[string]any{"en": "Mine"}})
}

func TestDBSharedLists(t *testing.T) {
	s := testDB(t)
	owner, friend := s.signUp(t), s.signUp(t)
	species := someSpecies(t, 1)

	shared := newList(t, s, owner, "Shared", models.VisibilityHidden)
	if len(shared) < utils.ShortUuidLength {
		t.Fatalf("new list id %q is short enough to guess", shared)
	}
	newReport(t, s, owner, species[0], shared)
	s.must(t, friend, http.StatusOK, "GET", "/v1/lists/"+shared, nil)

	r := s.do(t, "10.250.1.2", "", "POST", "/v1/lists/compare", map[string]any{"uuids": []string{shared}})
	if r.Status != http.StatusOK {
		t.Fatalf("compare = %d: %s", r.Status, r.Raw)
	}
	entry := get(r.Body, "result", 0)
	if get(entry, "user", "display_name") == nil || len(items(get(entry, "reports"))) != 1 {
		t.Errorf("compare entry = %v", entry)
	}
	// Only names and species: never where or when.
	for _, field := range []string{"latitude", "longitude", "from", "comment"} {
		if get(entry, "reports", 0, field) != nil {
			t.Errorf("compare leaks %s: %v", field, entry)
		}
	}

	// A list with an old 6-character id opens for its owner only.
	old := "o" + utils.GenerateUuidWithFixedLength(5)
	if err := models.DB.Model(&models.List{}).Where("uuid=?", shared).Update("uuid", old).Error; err != nil {
		t.Fatal(err)
	}
	s.must(t, owner, http.StatusOK, "GET", "/v1/lists/"+old, nil)
	if r := s.call(t, friend, "GET", "/v1/lists/"+old, nil); r.Status == http.StatusOK {
		t.Errorf("old short id opened for someone else: %s", r.Raw)
	}
	if r := s.do(t, "10.250.1.3", "", "POST", "/v1/lists/compare", map[string]any{"uuids": []string{old}}); r.Status == http.StatusOK {
		t.Errorf("compare opened an old short id: %s", r.Raw)
	}

	// Sharing it again gives it a long id that can be shared; the old one is gone.
	r = s.must(t, owner, http.StatusOK, "POST", "/v1/lists/"+old+"/visibility", map[string]any{"visibility": models.VisibilityHidden})
	renewed, _ := r.Body["uuid"].(string)
	if len(renewed) != utils.ShortUuidLength {
		t.Fatalf("shared old list got id %q", renewed)
	}
	s.must(t, friend, http.StatusOK, "GET", "/v1/lists/"+renewed, nil)
	if r := s.call(t, owner, "GET", "/v1/lists/"+old, nil); r.Status == http.StatusOK {
		t.Errorf("old id still opens: %s", r.Raw)
	}
	r = s.must(t, owner, http.StatusOK, "POST", "/v1/lists/"+renewed+"/visibility", map[string]any{"visibility": models.VisibilityPrivate})
	if r.Body["uuid"] != renewed {
		t.Errorf("making it private changed its id to %v", r.Body["uuid"])
	}
}

func TestDBExportHasEverything(t *testing.T) {
	s := testDB(t)
	u := s.signUp(t)
	species := someSpecies(t, 1)
	list := newList(t, s, u, "Exported", models.VisibilityPrivate)
	newReport(t, s, u, species[0], list)

	r := s.must(t, u, http.StatusOK, "GET", "/v1/users/export", nil)
	if len(r.Raw) < 100 || !strings.Contains(r.Raw, "Exported") {
		t.Errorf("export lacks the list: %.300s", r.Raw)
	}
}

// Compare names each entry's list, in the order asked for: one user can share several.
func TestDBCompareNamesLists(t *testing.T) {
	s := testDB(t)
	owner, other := s.signUp(t), s.signUp(t)
	species := someSpecies(t, 2)
	a := newList(t, s, owner, "Year", models.VisibilityHidden)
	b := newList(t, s, other, "Garden", models.VisibilityPublic)
	c := newList(t, s, owner, "Holiday", models.VisibilityHidden)
	newReport(t, s, owner, species[0], a)
	newReport(t, s, owner, species[1], c)

	for _, order := range [][]string{{a, b, c}, {c, a, b}} {
		r := s.do(t, "10.250.1.4", "", "POST", "/v1/lists/compare", map[string]any{"uuids": append(order, "unknown123")})
		if r.Status != http.StatusOK {
			t.Fatalf("compare = %d: %s", r.Status, r.Raw)
		}
		entries := items(r.Body["result"])
		if len(entries) != 3 {
			t.Fatalf("entries = %v", entries)
		}
		for i, uuid := range order {
			if get(entries[i], "list", "uuid") != uuid {
				t.Errorf("entry %d is list %v, want %s", i, get(entries[i], "list", "uuid"), uuid)
			}
		}
		if get(r.Body, "reference_list", "uuid") != order[0] {
			t.Errorf("reference = %v", get(r.Body, "reference_list", "uuid"))
		}
	}
	r := s.do(t, "10.250.1.4", "", "POST", "/v1/lists/compare", map[string]any{"uuids": []string{a, c}})
	if get(r.Body, "result", 0, "list", "name", "en") != "Year" || get(r.Body, "result", 1, "list", "name", "en") != "Holiday" {
		t.Errorf("names = %s", r.Raw)
	}
}

// A template that can't be used is the caller's mistake, and leaves nothing behind.
func TestDBListFromUnusableTemplate(t *testing.T) {
	s := testDB(t)
	owner, other := s.signUp(t), s.signUp(t)
	species := someSpecies(t, 1)
	private := s.must(t, other, http.StatusOK, "POST", "/v1/list_templates", map[string]any{
		"name": map[string]any{"en": "Mine"}, "species_uuids": species,
	}).Body["uuid"].(string)

	var before, after int64
	models.DB.Table("translations").Count(&before)
	for _, template := range []string{private, "nope"} {
		r := s.call(t, owner, "POST", "/v1/lists", map[string]any{"name": map[string]any{"en": "X"}, "template": template})
		if r.Status != http.StatusBadRequest || get(r.Body, "code") != float64(1001) {
			t.Errorf("template %s = %d: %s", template, r.Status, r.Raw)
		}
	}
	models.DB.Table("translations").Count(&after)
	if after != before {
		t.Errorf("failed creates left %d names behind", after-before)
	}
	// Your own private template works.
	s.must(t, other, http.StatusOK, "POST", "/v1/lists", map[string]any{"name": map[string]any{"en": "Y"}, "template": private})
}

// A sighting's zone comes without its outline, which can be megabytes.
func TestDBSightingZoneHasNoPolygon(t *testing.T) {
	s := testDB(t)
	u := s.signUp(t)
	species := someSpecies(t, 1)
	poly := `{"type":"Polygon","coordinates":[[[10,50],[11,50],[11,51],[10,51],[10,50]]]}`
	zone := models.Zone{
		Model:      models.Model{Timestamp: time.Now(), Uuid: utils.GenerateShortUuid()},
		Name:       models.Translation{Model: models.Model{Timestamp: time.Now(), Uuid: utils.GenerateShortUuid()}, En: zero.StringFrom("Testland")},
		Polygon:    []byte(poly),
		AdminLevel: models.ZoneTypeCounty,
	}
	if err := models.DB.Create(&zone).Error; err != nil {
		t.Fatal(err)
	}
	list := newList(t, s, u, "Zoned", models.VisibilityPrivate)
	created := s.must(t, u, http.StatusOK, "POST", "/v1/reports", map[string]any{
		"uuid": utils.GenerateUuidWithFixedLength(20), "species_uuid": species[0], "from": "2026-05-01T06:30:00Z",
		"latitude": 50.5, "longitude": 10.5, "list_uuids": []string{list},
	})
	uuid := created.Body["uuid"].(string)
	// Matching positions to zones is the geocoder's business; set it here.
	if err := models.DB.Table("reports").Where("uuid=?", uuid).Update("zone_id", zone.Id).Error; err != nil {
		t.Fatal(err)
	}
	answers := map[string]any{
		"get":          get(s.must(t, u, http.StatusOK, "GET", "/v1/reports/"+uuid, nil).Body, "result"),
		"list":         get(s.must(t, u, http.StatusOK, "GET", "/v1/reports/list", nil).Body, "result", 0),
		"list reports": get(s.must(t, u, http.StatusOK, "GET", "/v1/lists/"+list+"/reports", nil).Body, "result", 0),
		"me":           get(s.must(t, u, http.StatusOK, "GET", "/v1/users/me", nil).Body, "last_reports", 0),
	}
	for name, report := range answers {
		if get(report, "zone", "name", "en") != "Testland" {
			t.Errorf("%s: zone = %v", name, get(report, "zone"))
		}
		if get(report, "zone", "polygon") != nil {
			t.Errorf("%s: zone has its polygon", name)
		}
	}
}

// Visibility must be given and be 0, 1 or 2; a template is private or public.
func TestDBVisibilityIsChecked(t *testing.T) {
	s := testDB(t)
	u := s.signUp(t)
	// A public checklist is one shared with a link: nothing lists them.
	list := newList(t, s, u, "Shared", models.VisibilityPublic)
	if r := s.must(t, u, http.StatusOK, "GET", "/v1/lists/"+list, nil); get(r.Body, "result", "visibility") != float64(models.VisibilityHidden) {
		t.Errorf("public list stored as %v", get(r.Body, "result", "visibility"))
	}
	if r := s.call(t, u, "POST", "/v1/lists/"+list+"/visibility", map[string]any{}); r.Status != http.StatusBadRequest {
		t.Errorf("no visibility = %d", r.Status)
	}
	if r := s.must(t, u, http.StatusOK, "GET", "/v1/lists/"+list, nil); get(r.Body, "result", "visibility") != float64(models.VisibilityHidden) {
		t.Errorf("visibility after a refused change = %v", get(r.Body, "result", "visibility"))
	}
	s.must(t, u, http.StatusOK, "POST", "/v1/lists/"+list+"/visibility", map[string]any{"visibility": models.VisibilityPrivate})
	s.must(t, u, http.StatusOK, "POST", "/v1/lists/"+list+"/visibility", map[string]any{"visibility": models.VisibilityPublic})
	if r := s.must(t, u, http.StatusOK, "GET", "/v1/lists/"+list, nil); get(r.Body, "result", "visibility") != float64(models.VisibilityHidden) {
		t.Errorf("list made public stored as %v", get(r.Body, "result", "visibility"))
	}
	if r := s.call(t, u, "POST", "/v1/lists", map[string]any{"name": map[string]any{"en": "x"}, "visibility": 9}); r.Status != http.StatusBadRequest || get(r.Body, "code") != float64(1005) {
		t.Errorf("list with visibility 9 = %d: %s", r.Status, r.Raw)
	}
	if r := s.call(t, u, "POST", "/v1/list_templates", map[string]any{"name": map[string]any{"en": "x"}, "visibility": 9}); r.Status != http.StatusBadRequest {
		t.Errorf("template with visibility 9 = %d", r.Status)
	}
	r := s.must(t, u, http.StatusOK, "POST", "/v1/list_templates", map[string]any{"name": map[string]any{"en": "x"}, "visibility": 1})
	if r.Body["visibility"] != float64(models.VisibilityPrivate) {
		t.Errorf("hidden template stored as %v", r.Body["visibility"])
	}
}

// The share page shows each species on a shared list and when it was first seen there, newest first.
func TestDBShareList(t *testing.T) {
	s := testDB(t)
	u := s.signUp(t)
	species := someSpecies(t, 3)
	template := s.must(t, u, http.StatusOK, "POST", "/v1/list_templates", map[string]any{
		"name": map[string]any{"en": "Three"}, "species_uuids": species,
	}).Body["uuid"].(string)
	list := s.must(t, u, http.StatusOK, "POST", "/v1/lists", map[string]any{
		"name": map[string]any{"en": "Shared"}, "template": template, "visibility": models.VisibilityHidden,
	}).Body["uuid"].(string)
	s.must(t, u, http.StatusOK, "POST", "/v1/reports", map[string]any{"uuid": utils.GenerateUuidWithFixedLength(20), "species_uuid": species[0], "from": "2026-05-01T08:00:00Z", "list_uuids": []string{list}})
	s.must(t, u, http.StatusOK, "POST", "/v1/reports", map[string]any{"uuid": utils.GenerateUuidWithFixedLength(20), "species_uuid": species[0], "from": "2026-04-01T08:00:00Z", "list_uuids": []string{list}})
	s.must(t, u, http.StatusOK, "POST", "/v1/reports", map[string]any{"uuid": utils.GenerateUuidWithFixedLength(20), "species_uuid": species[2], "from": "2026-06-01T08:00:00Z", "list_uuids": []string{list}})

	shared, err := controllers.ShareList(list)
	if err != nil {
		t.Fatal(err)
	}
	if len(shared.Species) != 3 || shared.Seen != 2 || len(shared.Latest) != 2 {
		t.Fatalf("species %d, seen %d, latest %d", len(shared.Species), shared.Seen, len(shared.Latest))
	}
	if first := shared.Species[0].Seen; first.Month() != time.April {
		t.Errorf("first seen %v, want the earliest sighting", first)
	}
	if !shared.Species[1].Seen.IsZero() {
		t.Errorf("unseen species seen %v", shared.Species[1].Seen)
	}
	if shared.Latest[0].Seen.Month() != time.June {
		t.Errorf("latest starts with %v", shared.Latest[0].Seen)
	}

	s.must(t, u, http.StatusOK, "POST", "/v1/lists/"+list+"/visibility", map[string]any{"visibility": models.VisibilityPrivate})
	if _, err := controllers.ShareList(list); err == nil {
		t.Error("private list shared")
	}
}

// Anyone can search species and read about one, without signing in.
func TestDBSpeciesArePublic(t *testing.T) {
	s := testDB(t)
	var species models.Species
	if err := models.DB.Preload("Name").Where("id=(SELECT MIN(id) FROM species)").First(&species).Error; err != nil {
		t.Fatal(err)
	}
	q := strings.ToLower(species.Name.En.String[:3])
	r := s.do(t, "10.250.2.1", "", "GET", "/v1/species/search?lang=en&q="+q, nil)
	if r.Status != http.StatusOK || !strings.Contains(r.Raw, species.Uuid) {
		t.Fatalf("search %q = %d: %.300s", q, r.Status, r.Raw)
	}
	if n := len(items(r.Body["result"])); n > 8 {
		t.Errorf("search gave %d results", n)
	}
	if r := s.do(t, "10.250.2.1", "", "GET", "/v1/species/search?q=a", nil); r.Status != http.StatusOK || len(items(r.Body["result"])) != 0 {
		t.Errorf("one letter = %d: %s", r.Status, r.Raw)
	}
}

// Saving a sighting tells the share pages of that user's lists; reading changes nothing.
func TestDBChangesReachLiveSharePages(t *testing.T) {
	s := testDB(t)
	u := s.signUp(t)
	species := someSpecies(t, 1)
	changes, stop := live.Watch([]string{"user:" + u.Uuid}, nil)
	defer stop()
	heard := func() bool {
		select {
		case <-changes:
			return true
		case <-time.After(time.Second):
			return false
		}
	}

	list := newList(t, s, u, "Read", models.VisibilityPrivate)
	if !heard() {
		t.Error("a new list wasn't passed on")
	}
	s.must(t, u, http.StatusOK, "POST", "/v1/reports/list", map[string]any{"list_uuid": list})
	if heard() {
		t.Error("reading sightings was taken for a change")
	}
	if r := newReport(t, s, u, species[0]); r.Status != http.StatusOK {
		t.Fatalf("report = %d", r.Status)
	}
	if !heard() {
		t.Error("a new sighting wasn't passed on")
	}
}
