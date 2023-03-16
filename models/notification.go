package models

import "gopkg.in/guregu/null.v4/zero"

type Notification struct {
	Model
	ReportId int64       `json:"-"`
	Report   Report      `json:"report"`
	Comment  zero.String `json:"comment"`
}
