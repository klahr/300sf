package models

import (
	"gopkg.in/guregu/null.v4/zero"
)

type User struct {
	Model
	LastSeen      zero.Time `json:"last_seen"`
	DeactivatedAt zero.Time `json:"-"`
	DeletedAt     zero.Time `json:"-"`
	Username      string    `json:"username"`
	Email         string    `json:"-"`
	Password      string    `json:"-"`
	PasswordSalt  string    `json:"password_salt"`
	Permissions   int64     `json:"-"`
}
