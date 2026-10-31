// SPDX-License-Identifier: AGPL-3.0-or-later
package xenocanto

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// The query names the genus and species with tags and carries the key; the
// longitude is read under v3's name.
func TestRecordings(t *testing.T) {
	var got string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query().Get("query") + "|" + r.URL.Query().Get("key")
		w.Write([]byte(`{"numRecordings":"1","recordings":[{"id":"1","file":"https://xeno-canto.org/1/download","lat":"59.3","lon":"18.1"}]}`))
	}))
	defer server.Close()
	endpoint = server.URL

	res, err := Recordings("secret", "Grus grus")
	if err != nil {
		t.Fatal(err)
	}
	if got != "gen:Grus sp:grus q:A|secret" {
		t.Errorf("query = %q", got)
	}
	if len(res.Recordings) != 1 || res.Recordings[0].Longitude() != "18.1" {
		t.Errorf("recordings = %+v", res.Recordings)
	}

	// Without a key or a scientific name, nothing is asked.
	got = ""
	if _, err := Recordings("", "Grus grus"); err != nil || got != "" {
		t.Errorf("asked without a key: %q %v", got, err)
	}
	if _, err := Recordings("secret", "Grus"); err != nil || got != "" {
		t.Errorf("asked without a species name: %q %v", got, err)
	}
}

// An error status, such as v3's 401 for a wrong key, is an error.
func TestRecordingsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"client_error"}`, http.StatusUnauthorized)
	}))
	defer server.Close()
	endpoint = server.URL
	if _, err := Recordings("wrong", "Grus grus"); err == nil {
		t.Error("no error for 401")
	}
}
