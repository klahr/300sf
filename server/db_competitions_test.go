// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/klahr/300sf/server/models"
)

func newCompetition(t *testing.T, s *dbServer, u *testUser, extra map[string]any) response {
	t.Helper()
	now := time.Now().UTC()
	body := map[string]any{
		"name":  map[string]any{"en": "Autumn race", "sv": "Höstracet"},
		"from":  now.Add(-time.Hour).Format(time.RFC3339),
		"until": now.Add(30 * 24 * time.Hour).Format(time.RFC3339),
		"color": "#1273a3", "visibility": models.VisibilityPrivate, "creator_join": true,
	}
	for k, v := range extra {
		body[k] = v
	}
	return s.call(t, u, "POST", "/v1/competitions", body)
}

func TestDBCompetitionJoiningAndStandings(t *testing.T) {
	s := testDB(t)
	creator, friend, guesser := s.signUp(t), s.signUp(t), s.signUp(t)
	species := someSpecies(t, 3)

	r := newCompetition(t, s, creator, map[string]any{"password": "owl-night", "max_competitors": 2})
	if r.Status != http.StatusOK {
		t.Fatalf("create = %d: %s", r.Status, r.Raw)
	}
	id := r.Body["uuid"].(string)
	if len(id) < 24 {
		t.Errorf("competition id %q is short enough to guess", id)
	}

	// Wrong passwords are refused, and only so many may be tried.
	statuses := map[int]int{}
	for i := 0; i < 11; i++ {
		statuses[s.call(t, guesser, "PUT", "/v1/competitions/"+id+"/join", map[string]any{"password": "guess"}).Status]++
	}
	if statuses[http.StatusForbidden] != 10 || statuses[http.StatusTooManyRequests] != 1 {
		t.Errorf("11 wrong passwords answered %v, want 10 forbidden then too many", statuses)
	}

	s.must(t, friend, http.StatusOK, "PUT", "/v1/competitions/"+id+"/join", map[string]any{"password": "owl-night"})

	// Full at two: the creator and the friend.
	third := s.signUp(t)
	if r := s.call(t, third, "PUT", "/v1/competitions/"+id+"/join", map[string]any{"password": "owl-night"}); r.Status != http.StatusConflict {
		t.Errorf("joining a full competition = %d: %s", r.Status, r.Raw)
	}

	newReport(t, s, friend, species[0])
	newReport(t, s, friend, species[1])
	newReport(t, s, creator, species[2])

	info := s.must(t, friend, http.StatusOK, "GET", "/v1/competitions/"+id, nil)
	standings := items(info.Body["standings"])
	if len(standings) != 2 {
		t.Fatalf("standings = %v", info.Raw)
	}
	if get(standings[0], "count") != float64(2) || get(standings[0], "me") != true {
		t.Errorf("leader = %v, want the friend with 2", standings[0])
	}
	// Others see the day of a sighting, not its time.
	latest, err := time.Parse(time.RFC3339, get(standings[0], "latest").(string))
	if err != nil || !latest.Equal(latest.UTC().Truncate(24*time.Hour)) {
		t.Errorf("latest = %v (%v), want a whole day", get(standings[0], "latest"), err)
	}
}

func TestDBCompetitionLimits(t *testing.T) {
	s := testDB(t)
	creator := s.signUp(t)

	if r := newCompetition(t, s, creator, map[string]any{"from": time.Now().AddDate(-3, 0, 0).Format(time.RFC3339)}); r.Status != http.StatusBadRequest {
		t.Errorf("starting three years back = %d", r.Status)
	}

	// Many joining at once still stop at the limit.
	r := newCompetition(t, s, creator, map[string]any{"max_competitors": 3, "visibility": models.VisibilityPublic})
	if r.Status != http.StatusOK {
		t.Fatalf("create = %d: %s", r.Status, r.Raw)
	}
	id := r.Body["uuid"].(string)
	joiners := make([]*testUser, 6)
	for i := range joiners {
		joiners[i] = s.signUp(t)
	}
	var wg sync.WaitGroup
	for _, u := range joiners {
		wg.Add(1)
		go func(u *testUser) {
			defer wg.Done()
			s.call(t, u, "PUT", "/v1/competitions/"+id+"/join", nil)
		}(u)
	}
	wg.Wait()
	var n int64
	models.DB.Table("competition_users").Joins("JOIN competitions ON competitions.id=competition_users.competition_id").Where("competitions.uuid=?", id).Count(&n)
	if n != 3 {
		t.Errorf("%d took part, limit 3", n)
	}

	// Anyone may read a public competition's results.
	if r := s.do(t, "10.250.2.1", "", "GET", "/v1/competitions/"+id+"/result", nil); r.Status != http.StatusOK {
		t.Errorf("public result = %d", r.Status)
	}
}

func TestDBOldPrivateCompetitionsAreHidden(t *testing.T) {
	s := testDB(t)
	creator, stranger := s.signUp(t), s.signUp(t)
	r := newCompetition(t, s, creator, nil)
	if r.Status != http.StatusOK {
		t.Fatalf("create = %d: %s", r.Status, r.Raw)
	}
	old := "c" + r.Body["uuid"].(string)[:5]
	if err := models.DB.Model(&models.Competition{}).Where("uuid=?", r.Body["uuid"]).Update("uuid", old).Error; err != nil {
		t.Fatal(err)
	}
	s.must(t, creator, http.StatusOK, "GET", "/v1/competitions/"+old, nil)
	s.must(t, stranger, http.StatusNotFound, "GET", "/v1/competitions/"+old, nil)
	s.must(t, stranger, http.StatusNotFound, "PUT", "/v1/competitions/"+old+"/join", nil)
}
