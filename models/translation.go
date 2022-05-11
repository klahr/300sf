package models

import "time"

type Translation struct {
	Id        int64     `json:"-"`
	Timestamp time.Time `json:"-"`
	Uuid      string    `json:"uuid"`
	Sv        string    `json:"sv"`
	En        string    `json:"en"`
	La        string    `json:"la"`
}
