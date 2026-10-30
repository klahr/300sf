// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import "gopkg.in/guregu/null.v4/zero"

// LocalizedText is a text in each of the app's languages, stored in the row
// itself (columns <prefix>sv, <prefix>en, <prefix>de) rather than in
// translations, whose 512 characters are too few for a news post.
type LocalizedText struct {
	Sv zero.String `json:"sv"`
	En zero.String `json:"en"`
	De zero.String `json:"de"`
}

// Announcement is a post in the news feed. It is published from From (or
// when it was made) until Until, unless it is on hold (a draft, or taken
// down). Readers who have read it have a row in announcement_users.
type Announcement struct {
	Model
	// Name is the title in English, for the admin's list.
	Name   string        `json:"-"`
	Title  LocalizedText `json:"title" gorm:"embedded;embeddedPrefix:title_"`
	Body   LocalizedText `json:"body" gorm:"embedded;embeddedPrefix:body_"`
	From   zero.Time     `json:"from"`
	Until  zero.Time     `json:"until"`
	UserId zero.Int      `json:"-"`
	OnHold zero.Time     `json:"on_hold"`
	Icon   string        `json:"icon"`
	Color  string        `json:"color"`
}
