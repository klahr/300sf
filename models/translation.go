package models

type Translation struct {
	Model
	Sv string `json:"sv"`
	En string `json:"en"`
	La string `json:"la"`
}
