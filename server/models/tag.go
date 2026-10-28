// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import "gopkg.in/guregu/null.v4/zero"

type Tag struct {
	Model
	Name     string      `json:"name"`
	UserId   int64       `json:"-"`
	Color    zero.String `json:"color"`
	Archived zero.Time   `json:"-"`
}
