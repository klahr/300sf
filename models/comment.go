package models

import (
	"time"

	"gopkg.in/guregu/null.v4/zero"
)

type Comment struct {
	Model
	Timestamp time.Time `json:"timestamp"`
	UserId    int64     `json:"-"`
	User      User      `json:"user"`
	Text      string    `json:"text"`
	Edited    zero.Time `json:"edited"`
}
