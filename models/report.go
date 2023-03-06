package models

import "gopkg.in/guregu/null.v4/zero"

type Report struct {
	Model
	ListId     int64       `json:"-"`
	List       List        `json:"list"`
	SpeciesId  int64       `json:"-"`
	Species    Species     `json:"species"`
	From       zero.Time   `json:"from"`
	Until      zero.Time   `json:"until"`
	Latitude   zero.Float  `json:"latitude"`
	Longitude  zero.Float  `json:"longitude"`
	Quantity   zero.Int    `json:"quantity"`
	AgeId      zero.Int    `json:"-"`
	Age        Age         `json:"age"`
	ActivityId zero.Int    `json:"-"`
	Activity   Activity    `json:"activity"`
	GenderId   zero.Int    `json:"-"`
	Gender     Gender      `json:"gender"`
	MethodId   zero.Int    `json:"-"`
	Method     Method      `json:"method"`
	Comment    zero.String `json:"comment"`
}
