// SPDX-License-Identifier: GPL-2.0-or-later
package models

type Species struct {
	Model
	NameId  int64       `json:"-"`
	Name    Translation `json:"name"`
	OrderId int64       `json:"-"`
	Order   Order       `json:"order"`
}
