// SPDX-License-Identifier: AGPL-3.0-or-later
package xenocanto

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	Lng      string `json:"lng"`
	Alt      string `json:"alt"`
}

type RecordingsResponse struct {
	NumRecordings string      `json:"numRecordings"`
	NumSpecies    string      `json:"numSpecies"`
	Recordings    []Recording `json:"recordings"`
}

var client = &http.Client{Timeout: 10 * time.Second}

func Recordings(query string) RecordingsResponse {
	var response RecordingsResponse

	q := fmt.Sprintf("https://xeno-canto.org/api/2/recordings?query=%s%%20q:A", query)

	r, err := client.Get(q)
	if err == nil {
		defer r.Body.Close()
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &response)
	}

	return response
}
