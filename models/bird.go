package models

import "time"

type Bird struct {
	Id        int64       `json:"-"`
	Timestamp time.Time   `json:"-"`
	Uuid      string      `json:"uuid"`
	NameId    int64       `json:"-"`
	Name      Translation `json:"name"`
}
