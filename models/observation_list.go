package models

import (
	"gopkg.in/guregu/null.v4/zero"
)

type ObservationList struct {
	Model
	Name     string    `json:"name"`
	UserId   int64     `json:"-"`
	User     User      `json:"user"`
	From     zero.Time `json:"from"`
	Until    zero.Time `json:"until"`
	Hidden   bool      `json:"hidden"`
	Favorite bool      `json:"favorite"`
}
