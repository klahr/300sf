package controllers

import (
	"strings"
	"testing"
	"time"

	"github.com/klahr/300sf/server/models"
	"gopkg.in/guregu/null.v4/zero"
)

func TestCSVDelimiter(t *testing.T) {
	for lang, want := range map[string]rune{"sv": ';', "de": ';', "en": ',', "": ','} {
		if got := csvDelimiter(lang); got != want {
			t.Errorf("csvDelimiter(%q) = %q, want %q", lang, got, want)
		}
	}
}

func TestCSVTextNeutralisesFormulas(t *testing.T) {
	cases := map[string]string{
		"=HYPERLINK(\"x\")": "'=HYPERLINK(\"x\")",
		"+46 70":            "'+46 70",
		"-1":                "'-1",
		"@SUM(A1)":          "'@SUM(A1)",
		"Spatserade":        "Spatserade",
		"":                  "",
	}
	for in, want := range cases {
		if got := csvText(in); got != want {
			t.Errorf("csvText(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWriteCSV(t *testing.T) {
	out, err := writeCSV([][]string{{"species", "comment"}, {"Sånglärka", "a; b"}}, "sv")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, utf8BOM) {
		t.Errorf("missing byte order mark: %q", out)
	}
	want := utf8BOM + "species;comment\nSånglärka;\"a; b\"\n"
	if out != want {
		t.Errorf("writeCSV = %q, want %q", out, want)
	}
}

func TestSightingTimes(t *testing.T) {
	stockholm, err := time.LoadLocation("Europe/Stockholm")
	if err != nil {
		t.Skip("no time zone data")
	}
	local := func(t time.Time) time.Time { return t.In(stockholm) }

	dateOnly := time.Date(2023, 5, 25, 0, 0, 0, 0, time.UTC)
	if got := csvTime(dateOnly, local); got != "2023-05-25" {
		t.Errorf("csvTime(date only) = %q", got)
	}
	if got := jsonTime(dateOnly); got != "2023-05-25" {
		t.Errorf("jsonTime(date only) = %q", got)
	}

	withTime := time.Date(2025, 4, 3, 11, 24, 0, 0, time.UTC)
	if got := csvTime(withTime, local); got != "2025-04-03 13:24" {
		t.Errorf("csvTime = %q, want local summer time", got)
	}
	if got := jsonTime(withTime); got != "2025-04-03T11:24:00Z" {
		t.Errorf("jsonTime = %q", got)
	}
}

func TestTranslatedFallsBackToEnglish(t *testing.T) {
	name := models.Translation{Sv: zero.StringFrom("Skata"), En: zero.StringFrom("Eurasian magpie")}
	if got := translated(name, "sv"); got != "Skata" {
		t.Errorf("sv = %q", got)
	}
	if got := translated(name, "de"); got != "Eurasian magpie" {
		t.Errorf("de without German name = %q, want English", got)
	}
}
