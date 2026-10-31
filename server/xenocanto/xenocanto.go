// SPDX-License-Identifier: AGPL-3.0-or-later
package xenocanto

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Sono struct {
	Small string `json:"small"`
}

type Recording struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Sex      string `json:"sex"`
	Stage    string `json:"stage"`
	Method   string `json:"method"`
	Url      string `json:"url"`
	File     string `json:"file"`
	Filename string `json:"file-name"`
	Sono     Sono   `json:"sono"`
	Rec      string `json:"rec"`
	Cnt      string `json:"cnt"`
	Loc      string `json:"loc"`
	Lat      string `json:"lat"`
	// API v3 calls it lon; v2 called it lng.
	Lon string `json:"lon"`
	Lng string `json:"lng"`
	Alt string `json:"alt"`
}

type RecordingsResponse struct {
	NumRecordings string      `json:"numRecordings"`
	NumSpecies    string      `json:"numSpecies"`
	Recordings    []Recording `json:"recordings"`
}

var (
	client   = &http.Client{Timeout: 10 * time.Second}
	endpoint = "https://xeno-canto.org/api/3/recordings"
)

// Longitude is the recording's longitude, whichever name it came under.
func (r Recording) Longitude() string {
	if r.Lon != "" {
		return r.Lon
	}
	return r.Lng
}

// Recordings are the best (quality A) recordings of a species, by its
// scientific name, from API v3; key is required (v2 is gone).
func Recordings(key, scientificName string) (RecordingsResponse, error) {
	var response RecordingsResponse
	parts := strings.Fields(scientificName)
	if key == "" || len(parts) < 2 {
		return response, nil
	}
	// v3 only takes tagged queries: the genus and the species.
	query := fmt.Sprintf("gen:%s sp:%s q:A", parts[0], strings.ToLower(parts[1]))
	q := endpoint + "?query=" + url.QueryEscape(query) + "&key=" + url.QueryEscape(key)

	r, err := client.Get(q)
	if err != nil {
		return response, err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return response, fmt.Errorf("xeno-canto: %s", r.Status)
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return response, err
	}
	return response, json.Unmarshal(body, &response)
}
