package models

type Zone struct {
	Model
	NameId  int64       `json:"-"`
	Name    Translation `json:"name"`
	Polygon string      `json:"polygon"`
	Type    int64       `json:"type"`
}
