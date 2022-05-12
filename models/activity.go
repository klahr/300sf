package models

type Activity struct {
	Model
	NameId int64       `json:"-"`
	Name   Translation `json:"name"`
}
