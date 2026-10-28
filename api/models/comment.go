// SPDX-License-Identifier: GPL-2.0-or-later
package models

import (
	"gopkg.in/guregu/null.v4/zero"
)

type Comment struct {
	Model
	UserId int64     `json:"-"`
	Text   string    `json:"text"`
	Edited zero.Time `json:"edited"`
}
