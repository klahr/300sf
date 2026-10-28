// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import "gopkg.in/guregu/null.v4/zero"

type Report struct {
	Model
	UserId       int64         `json:"-"`
	SpeciesId    int64         `json:"-"`
	Species      Species       `json:"species"`
	From         zero.Time     `json:"from"`
	Until        zero.Time     `json:"until"`
	Latitude     zero.Float    `json:"latitude"`
	Longitude    zero.Float    `json:"longitude"`
	ZoneId       zero.Int      `json:"-"`
	Zone         *Zone         `json:"zone"`
	Quantity     zero.Int      `json:"quantity"`
	AgeId        zero.Int      `json:"-"`
	Age          *Age          `json:"age"`
	ActivityId   zero.Int      `json:"-"`
	Activity     *Activity     `json:"activity"`
	GenderId     zero.Int      `json:"-"`
	Gender       *Gender       `json:"gender"`
	MethodId     zero.Int      `json:"-"`
	Method       *Method       `json:"method"`
	CommentId    zero.Int      `json:"-"`
	Comment      *Comment      `json:"comment"`
	Archived     zero.Time     `json:"-"`
	Site         zero.String   `json:"site"`
	Lists        []List        `json:"lists" gorm:"many2many:list_reports"`
	Tags         []Tag         `json:"tags" gorm:"many2many:report_tags"`
	Notification *Notification `json:"notification"`
}
