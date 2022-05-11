package models

import (
	"time"

	"gopkg.in/guregu/null.v4/zero"
)

type ObservationList struct {
	Id        int64     `json:"-"`
	Timestamp time.Time `json:"-"`
	Uuid      string    `json:"uuid"`
	Name      string    `json:"name"`
	UserId    int64     `json:"-"`
	User      User      `json:"user"`
	From      zero.Time `json:"from"`
	Until     zero.Time `json:"until"`
	Hidden    bool      `json:"hidden"`
	Favorite  bool      `json:"favorite"`
}
