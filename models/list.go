package models

import (
	"gopkg.in/guregu/null.v4/zero"
)

type List struct {
	Model
	NameId     int64       `json:"-"`
	Name       Translation `json:"name"`
	UserId     int64       `json:"-"`
	User       User        `json:"user"`
	From       zero.Time   `json:"from"`
	Until      zero.Time   `json:"until"`
	Visibility int64       `json:"visibility"`
	Archived   zero.Time   `json:"archived"`
}
