// SPDX-License-Identifier: AGPL-3.0-or-later
package models

type Age struct {
	Model
	NameId int64       `json:"-"`
	Name   Translation `json:"name"`
}
