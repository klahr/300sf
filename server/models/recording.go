// SPDX-License-Identifier: AGPL-3.0-or-later
package models

type Recording struct {
	Url       string  `json:"url"`
	Filename  string  `json:"filename"`
	Type      string  `json:"type"`
	Sex       string  `json:"sex"`
	Stage     string  `json:"stage"`
	Method    string  `json:"method"`
	Sono      string  `json:"sono"`
	Recorder  string  `json:"recorder"`
	Country   string  `json:"country"`
	Location  string  `json:"location"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
