package models

import "gopkg.in/guregu/null.v4/zero"

type Report struct {
	Model
	ObservationListId int64           `json:"-"`
	ObservationList   ObservationList `json:"observation_list"`
	SpeciesId         int64           `json:"-"`
	Species           Species         `json:"species"`
	From              zero.Time       `json:"from"`
	Until             zero.Time       `json:"until"`
	Latitude          zero.Float      `json:"latitude"`
	Longitude         zero.Float      `json:"longitude"`
	Quantity          zero.Int        `json:"quantity"`
	AgeId             int64           `json:"-"`
	Age               Age             `json:"age"`
	ActivityId        int64           `json:"-"`
	Activity          Activity        `json:"activity"`
	GenderId          int64           `json:"-"`
	Gender            Gender          `json:"gender"`
	MethodId          int64           `json:"-"`
	Method            Method          `json:"method"`
	Comment           zero.String     `json:"comment"`
}
