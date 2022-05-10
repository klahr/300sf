package models

import "time"

type Language struct {
	Id         int64     `json:"-"`
	Timestamp  time.Time `json:"-"`
	Uuid       string    `json:"uuid"`
	IsoName    string    `json:"iso_name"`
	NativeName string    `json:"native_name"`
	Code       string    `json:"code"`
}
