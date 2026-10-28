// SPDX-License-Identifier: AGPL-3.0-or-later
package models

type Activity struct {
	Model
	NameId int64       `json:"-"`
	Name   Translation `json:"name"`
}
