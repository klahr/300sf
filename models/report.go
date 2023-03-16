package models

import "gopkg.in/guregu/null.v4/zero"

type Report struct {
	Model
	UserId       int64         `json:"-"`
	User         User          `json:"user"`
	SpeciesId    int64         `json:"-"`
	Species      Species       `json:"species"`
	From         zero.Time     `json:"from"`
	Until        zero.Time     `json:"until"`
	Latitude     zero.Float    `json:"latitude"`
	Longitude    zero.Float    `json:"longitude"`
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
	Lists        []List        `json:"lists" gorm:"many2many:list_reports"`
	Tags         []Tag         `json:"tags" gorm:"many2many:report_tags"`
	Notification *Notification `json:"notification"`
}
