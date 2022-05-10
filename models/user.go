package models

import (
	"time"

	"gopkg.in/guregu/null.v4/zero"
)

type User struct {
	Id            int64     `json:"-"`
	Timestamp     time.Time `json:"-"`
	Uuid          string    `json:"uuid"`
	LastSeen      zero.Time `json:"last_seen"`
	DeactivatedAt zero.Time `json:"-"`
	DeletedAt     zero.Time `json:"-"`
	Email         string    `json:"-"`
	Password      string    `json:"-"`
	PasswordSalt  string    `json:"password_salt"`
	Permissions   int64     `json:"-"`
}
