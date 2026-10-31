// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import (
	"gopkg.in/guregu/null.v4/zero"
)

type List struct {
	Model
	NameId         int64        `json:"-"`
	Name           Translation  `json:"name"`
	UserId         int64        `json:"-"`
	From           zero.Time    `json:"from"`
	Until          zero.Time    `json:"until"`
	Visibility     int64        `json:"visibility"`
	Archived       zero.Time    `json:"-"`
	Species        []Species    `json:"species" gorm:"many2many:list_species"`
	ListTemplateId zero.Int     `json:"-"`
	ListTemplate   ListTemplate `json:"-"`
	Color          zero.String  `json:"color"`
	// PreviousUuid is the old id of a list that got a new one when shared, for
	// its owner's sightings saved offline with the old id.
	PreviousUuid zero.String `json:"-"`
	// Template is the uuid of the species list the list was made from, when the list is shown to its owner.
	Template string `json:"template,omitempty" gorm:"-"`
}
