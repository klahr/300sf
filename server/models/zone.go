// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import "gopkg.in/guregu/null.v4/zero"

type Zone struct {
	Model
	NameId       int64       `json:"-"`
	Name         Translation `json:"name"`
	Polygon      []byte      `json:"polygon"`
	AdminLevel   int64       `json:"admin_level"`
	ParentZoneId zero.Int    `json:"-"`
	ParentZone   *Zone       `json:"parent_zone"`
	// The OpenStreetMap relation ("R54220") of a zone made by the geocoder.
	OsmId zero.String `json:"-"`
}
