package models

import "time"

type DeleteAccountRequest struct {
	Model
	ExpireAt time.Time `json:"expire_at"`
	Token    string    `json:"token"`
	UserId   int64     `json:"-"`
	User     User      `json:"user"`
	DeleteAt time.Time `json:"delete_at"`
}
