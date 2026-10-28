// SPDX-License-Identifier: AGPL-3.0-or-later
package models

type CompetitionUsers struct {
	CompetitionId int64 `json:"competition_id"`
	UserId        int64 `json:"user_id"`
}
