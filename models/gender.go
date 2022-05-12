package models

type Gender struct {
	Model
	NameId int64       `json:"-"`
	Name   Translation `json:"name"`
}
