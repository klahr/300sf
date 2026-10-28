package models

import (
	"time"

	"gopkg.in/guregu/null.v4/zero"
)

type DeleteAccountRequest struct {
	Model
	ExpireAt time.Time `json:"expire_at"`
	Token    string    `json:"token"`
	UserId   int64     `json:"-"`
	User     User      `json:"user"`
	DeleteAt zero.Time `json:"delete_at"`
}
