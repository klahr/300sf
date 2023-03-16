package models

import "gopkg.in/guregu/null.v4/zero"

type Comment struct {
	Model
	UserId int64     `json:"-"`
	User   User      `json:"user"`
	Text   string    `json:"text"`
	Edited zero.Time `json:"edited"`
}
