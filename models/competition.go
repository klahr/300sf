package models

import (
	"time"

	"gopkg.in/guregu/null.v4/zero"
)

type Competition struct {
	Model
	Name             Translation  `json:"name"`
	UserId           zero.Int     `json:"-"`
	User             *User        `json:"user"`
	From             time.Time    `json:"from"`
	Until            time.Time    `json:"until"`
	FinishedAt       zero.Time    `json:"finished_at"`
	ListTemplateId   zero.Int     `json:"-"`
	ListTemplate     ListTemplate `json:"list_template"`
	Description      Translation  `json:"description"`
	EndOnReports     zero.Int     `json:"end_on_reports"`
	MaxCompetitors   zero.Int     `json:"max_competitors"`
	Visibility       int64        `json:"visibility"`
	JoinFrom         time.Time    `json:"join_from"`
	JoinUntil        time.Time    `json:"join_until"`
	InviteOnly       bool         `json:"invite_only"`
	Icon             string       `json:"icon"`
	Color            string       `json:"color"`
	CompetitionZones []Zone       `json:"zones" gorm:"many2many:competition_zones"`
	Archived         zero.Time    `json:"-"`
	Participants     []User       `json:"participants" gorm:"many2many:competition_participants"`
}
