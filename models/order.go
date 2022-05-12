package models

type Order struct {
	Model
	NameId int64       `json:"-"`
	Name   Translation `json:"name"`
}
