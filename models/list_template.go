package models

import "gopkg.in/guregu/null.v4/zero"

type ListTemplate struct {
	Model
	NameId int64       `json:"-"`
	Name   Translation `json:"name"`
	UserId int64       `json:"-"`
	User   User        `json:"user"`
	//Species    []Species   `json:"species"`
	Visibility int64     `json:"visibility"`
	Archived   zero.Time `json:"archived"`
}
