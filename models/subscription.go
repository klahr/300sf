// SPDX-License-Identifier: GPL-2.0-or-later
package models

import "gopkg.in/guregu/null.v4/zero"

type Subscription struct {
	Model
	UserId   int64     `json:"-"`
	User     User      `json:"user"`
	ZoneId   int64     `json:"-"`
	Zone     Zone      `json:"zone"`
	Paused   zero.Time `json:"paused"`
	Archived zero.Time `json:"-"`
	Species  []Species `json:"species" gorm:"many2many:subscription_species"`
}
