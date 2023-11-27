package models

import "gopkg.in/guregu/null.v4/zero"

type CompetitionInvite struct {
	Model
	UserId        int64     `json:"-"`
	User          User      `json:"user"`
	CompetitionId int64     `json:"-"`
	AcceptedAt    zero.Time `json:"accepted_at"`
	RejectedAt    zero.Time `json:"rejected_at"`
}
