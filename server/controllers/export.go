package controllers

import (
	"bytes"
	"encoding/csv"
	"strings"
	"time"

	"github.com/klahr/300sf/server/models"
)

// Shared formatting for the data exports: the per-list CSV (ReportsExport) and the account export
// (UsersExportData).

// The byte order mark that makes Excel read a CSV file as UTF-8, so "Sånglärka" stays intact.
const utf8BOM = string(rune(0xFEFF))

// csvDelimiter returns the separator spreadsheet programs expect for the language: Swedish and
// German locales use the comma as decimal separator and open comma-separated files as one column.
func csvDelimiter(language string) rune {
	if language == "sv" || language == "de" {
		return ';'
	}
	return ','
}

// csvText keeps spreadsheet programs from treating text typed by the user as a formula.
func csvText(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return "'" + s
	}
	return s
}

// writeCSV renders the rows with the delimiter for the language, behind a byte order mark.
func writeCSV(rows [][]string, language string) (string, error) {
	var b bytes.Buffer
	b.WriteString(utf8BOM)
	w := csv.NewWriter(&b)
	w.Comma = csvDelimiter(language)
	if err := w.WriteAll(rows); err != nil {
		return "", err
	}
	return b.String(), nil
}

// isDateOnly reports whether a sighting time was recorded as a date without a time of day. Those are
// stored as midnight UTC, mostly by imports from the old app.
func isDateOnly(t time.Time) bool {
	u := t.UTC()
	return u.Hour() == 0 && u.Minute() == 0 && u.Second() == 0
}

// csvTime formats a sighting time for the CSV in the user's time zone, or just the date when no time
// of day was recorded.
func csvTime(t time.Time, local func(time.Time) time.Time) string {
	if isDateOnly(t) {
		return t.UTC().Format("2006-01-02")
	}
	return local(t).Format("2006-01-02 15:04")
}

// jsonTime formats a time as ISO 8601 for the account export: a date alone when no time of day was
// recorded, otherwise the moment in UTC.
func jsonTime(t time.Time) string {
	if isDateOnly(t) {
		return t.UTC().Format("2006-01-02")
	}
	return t.UTC().Format(time.RFC3339)
}

// translated returns the name in the language, or in English when there is none.
func translated(t models.Translation, language string) string {
	var s string
	switch language {
	case "sv":
		s = t.Sv.String
	case "de":
		s = t.De.String
	case "la":
		s = t.La.String
	default:
		s = t.En.String
	}
	if s == "" {
		s = t.En.String
	}
	return s
}

// exportName is a name in the languages it has, without the empty ones and the internal id.
type exportName struct {
	Sv string `json:"sv,omitempty"`
	En string `json:"en,omitempty"`
	De string `json:"de,omitempty"`
	La string `json:"la,omitempty"`
}

func newExportName(t models.Translation) exportName {
	return exportName{Sv: t.Sv.String, En: t.En.String, De: t.De.String, La: t.La.String}
}
