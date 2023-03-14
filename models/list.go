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
	// Reports    []Report    `json:"reports" gorm:"many2many:list_reports"`
}
