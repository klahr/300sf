package models

import "time"

type Method struct {
	Id        int64       `json:"-"`
	Timestamp time.Time   `json:"-"`
	Uuid      string      `json:"uuid"`
	Name      Translation `json:"name"`
}
