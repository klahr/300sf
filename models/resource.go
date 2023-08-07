// SPDX-License-Identifier: GPL-2.0-or-later
package models

type Resource struct {
	Model
	Type     string `json:"type"`
	Mimetype string `json:"mimetype"`
	Filename string `json:"filename"`
	Url      string `json:"url"`
	UserId   int64  `json:"-"`
}
