package models

type Age struct {
	Model
	NameId int64       `json:"-"`
	Name   Translation `json:"name"`
}
