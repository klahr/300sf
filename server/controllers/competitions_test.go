package controllers

import (
	"testing"
	"time"

	"github.com/klahr/300sf/server/models"
)

func TestRankCompetition(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, 5, d, 8, 0, 0, 0, time.UTC) }
	user := func(id int64, name string) models.User {
		u := models.User{DisplayName: name, Username: name}
		u.Id = id
		return u
	}
	people := []models.User{user(1, "anna"), user(2, "erik"), user(3, "maja"), user(4, "nils")}
	sightings := []competitionSighting{
		{UserId: 1, SpeciesId: 10, At: day(1)}, {UserId: 1, SpeciesId: 11, At: day(5)}, {UserId: 1, SpeciesId: 12, At: day(9)},
		{UserId: 2, SpeciesId: 10, At: day(2)}, {UserId: 2, SpeciesId: 11, At: day(3)},
		{UserId: 3, SpeciesId: 10, At: day(4)}, {UserId: 3, SpeciesId: 12, At: day(6)},
	}

	standings, wonAt := rankCompetition(people, sightings, 0)
	if wonAt.Valid {
		t.Fatalf("no target, but won at %v", wonAt.Time)
	}
	want := []struct {
		name  string
		rank  int
		count int
	}{{"anna", 1, 3}, {"erik", 2, 2}, {"maja", 3, 2}, {"nils", 4, 0}}
	for i, w := range want {
		s := standings[i]
		if s.User.Username != w.name || s.Rank != w.rank || s.Count != w.count {
			t.Errorf("place %d = %s rank %d count %d, want %s rank %d count %d", i, s.User.Username, s.Rank, s.Count, w.name, w.rank, w.count)
		}
	}
	if standings[3].Latest.Valid {
		t.Error("someone without sightings has a latest sighting")
	}

	// First to two species: Erik gets there on day 3, so later sightings stop counting.
	standings, wonAt = rankCompetition(people, sightings, 2)
	if !wonAt.Valid || !wonAt.Time.Equal(day(3)) {
		t.Fatalf("won at %v, want day 3", wonAt)
	}
	if standings[0].User.Username != "erik" || standings[0].Count != 2 {
		t.Errorf("winner = %s with %d", standings[0].User.Username, standings[0].Count)
	}
	if standings[1].User.Username != "anna" || standings[1].Count != 1 {
		t.Errorf("anna after the win = %s with %d, want 1 species", standings[1].User.Username, standings[1].Count)
	}
	if standings[2].User.Username != "maja" || standings[2].Count != 0 {
		t.Errorf("maja's sighting after the win counted: %+v", standings[2])
	}
	if standings[2].Rank != standings[3].Rank {
		t.Errorf("equal participants got different ranks: %d and %d", standings[2].Rank, standings[3].Rank)
	}
}
