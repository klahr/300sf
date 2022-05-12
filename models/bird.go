package models

type Bird struct {
	Model
	NameId int64       `json:"-"`
	Name   Translation `json:"name"`
}
