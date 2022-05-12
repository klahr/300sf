package models

type Habitat struct {
	Model
	NameId int64       `json:"-"`
	Name   Translation `json:"name"`
}
