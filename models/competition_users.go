// SPDX-License-Identifier: GPL-2.0-or-later
package models

type CompetitionUsers struct {
	CompetitionId int64 `json:"competition_id"`
	UserId        int64 `json:"user_id"`
}
