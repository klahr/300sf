package models

import (
	"gopkg.in/guregu/null.v4/zero"
)

type ObservationData struct {
	BirdId    int64       `json:"-"`
	Bird      Bird        `json:"bird"`
	Latitude  zero.Float  `json:"latitude"`
	Longitude zero.Float  `json:"longitude"`
	Count     zero.Int    `json:"count"`
	Age       *Age        `json:"age"`
	Activity  *Activity   `json:"activity"`
	Gender    *Gender     `json:"gender"`
	Method    *Method     `json:"method"`
	Comment   zero.String `json:"comment"`
}

type Observation struct {
	Model
	ObservationListId int64     `json:"-"`
	ListId            int64     `json:"-"`
	From              zero.Time `json:"from"`
	Until             zero.Time `json:"until"`
	ObservationData
}
