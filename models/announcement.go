package models

import "gopkg.in/guregu/null.v4/zero"

type Announcement struct {
	Model
	Name      string      `json:"name"`
	ContentId int         `json:"-"`
	Content   Translation `json:"content"`
	From      zero.Time   `json:"from"`
	Until     zero.Time   `json:"until"`
	Priority  int         `json:"priority"`
	UserId    zero.Int    `json:"-"`
	OnHold    zero.Time   `json:"on_hold"`
}
