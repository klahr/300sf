package models

import "gopkg.in/guregu/null.v4/zero"

type Tag struct {
	Model
	Name     string      `json:"name"`
	UserId   int64       `json:"-"`
	Color    zero.String `json:"color"`
	Archived zero.String `json:"-"`
}
