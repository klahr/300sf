package models

type Method struct {
	Model
	NameId int64       `json:"-"`
	Name   Translation `json:"name"`
}
