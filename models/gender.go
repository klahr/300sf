package models

import "time"

type Gender struct {
	Id        int64       `json:"-"`
	Timestamp time.Time   `json:"-"`
	Uuid      string      `json:"uuid"`
	Name      Translation `json:"name"`
}
