// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import "gopkg.in/guregu/null.v4/zero"

type Notification struct {
	Model
	ReportId  int64     `json:"-"`
	Report    Report    `json:"report"`
	CommentId int64     `json:"-"`
	Comment   Comment   `json:"comment"`
	Processed zero.Time `json:"processed"`
	Comments  []Comment `json:"comments" gorm:"many2many:notification_comments"`
}
