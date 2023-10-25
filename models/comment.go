// SPDX-License-Identifier: GPL-2.0-or-later
package models

import (
	"time"

	"gopkg.in/guregu/null.v4/zero"
)

type Comment struct {
	Model
	Timestamp time.Time `json:"timestamp"`
	UserId    int64     `json:"-"`
	Text      string    `json:"text"`
	Edited    zero.Time `json:"edited"`
}
