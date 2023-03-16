package models

import "gopkg.in/guregu/null.v4/zero"

type Notification struct {
	Model
	ReportId  int64     `json:"-"`
	Report    Report    `json:"report"`
	CommentId int64     `json:"-"`
	Comment   Comment   `json:"comment"`
	Processed zero.Time `json:"processed"`
}
