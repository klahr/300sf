package models

import "gopkg.in/guregu/null.v4/zero"

type Zone struct {
	Model
	NameId       int64       `json:"-"`
	Name         Translation `json:"name"`
	Polygon      []byte      `json:"polygon"`
	AdminLevel   int64       `json:"admin_level"`
	ParentZoneId zero.Int    `json:"-"`
	ParentZone   *Zone       `json:"parent_zone"`
}
