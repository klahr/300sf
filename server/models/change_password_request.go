package models

import "time"

type ChangePasswordRequest struct {
	Model
	ExpireAt time.Time `json:"expire_at"`
	Token    string    `json:"token"`
	UserId   int64     `json:"-"`
	User     User      `json:"user"`
}
