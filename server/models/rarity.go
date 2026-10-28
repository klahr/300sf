// SPDX-License-Identifier: AGPL-3.0-or-later
package models

type Rarity struct {
	Id        int64  `json:"-"`
	Category  string `json:"category"`
	SpeciesId int64  `json:"-"`
	ZoneId    int64  `json:"-"`
}
