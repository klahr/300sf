// SPDX-License-Identifier: GPL-2.0-or-later
package models

type Gender struct {
	Model
	NameId int64       `json:"-"`
	Name   Translation `json:"name"`
}
