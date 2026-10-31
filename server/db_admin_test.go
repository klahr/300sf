// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"net/http"
	"testing"
)

// Statistics are for admins only, and count the signups and sightings just made.
func TestDBAdminStats(t *testing.T) {
	s := testDB(t)
	user, admin := s.signUp(t), s.signUpAdmin(t)
	species := someSpecies(t, 1)
	newReport(t, s, user, species[0])

	if r := s.call(t, user, "GET", "/v1/admin/stats", nil); r.Status == http.StatusOK {
		t.Fatalf("a user got the statistics: %s", r.Raw)
	}
	r := s.must(t, admin, http.StatusOK, "GET", "/v1/admin/stats", nil)
	weeks := items(r.Body["weeks"])
	if len(weeks) != 52 {
		t.Fatalf("weeks = %d", len(weeks))
	}
	last := func(key string) float64 {
		v := items(r.Body[key])
		return v[len(v)-1].(float64)
	}
	if last("signups") < 2 || last("weekly_active") < 1 || last("sightings") < 1 {
		t.Errorf("this week: signups %v, active %v, sightings %v", last("signups"), last("weekly_active"), last("sightings"))
	}
	if last("accounts") < last("signups") {
		t.Errorf("accounts %v below this week's signups %v", last("accounts"), last("signups"))
	}
	cohorts := items(r.Body["cohorts"])
	newest := cohorts[len(cohorts)-1]
	if get(newest, "size").(float64) < 2 || get(newest, "retained", 0).(float64) < 1 {
		t.Errorf("this month's cohort = %v", newest)
	}
}

// Anyone sees the donation progress; only admins set it.
func TestDBDonation(t *testing.T) {
	s := testDB(t)
	user, admin := s.signUp(t), s.signUpAdmin(t)

	if r := s.call(t, user, "PUT", "/v1/donation", map[string]any{"raised": 5, "goal": 10}); r.Status == http.StatusOK {
		t.Fatalf("a user set the donation progress: %s", r.Raw)
	}
	s.must(t, admin, http.StatusBadRequest, "PUT", "/v1/donation", map[string]any{"raised": 5, "goal": 0})
	s.must(t, admin, http.StatusOK, "PUT", "/v1/donation", map[string]any{"raised": 1500, "goal": 3000})
	r := s.do(t, "10.250.9.1", "", "GET", "/v1/donation", nil)
	if r.Status != http.StatusOK || r.Body["raised"] != float64(1500) || r.Body["goal"] != float64(3000) {
		t.Errorf("donation = %d %s", r.Status, r.Raw)
	}
}

// An admin sees a user's details, with their language.
func TestDBAdminUserDetails(t *testing.T) {
	s := testDB(t)
	user, admin := s.signUp(t), s.signUpAdmin(t)
	r := s.must(t, admin, http.StatusOK, "GET", "/v1/users/"+user.Uuid, nil)
	if code, _ := get(r.Body, "user", "language", "code").(string); code == "" {
		t.Errorf("user details lack the language: %s", r.Raw)
	}
}
