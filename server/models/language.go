// SPDX-License-Identifier: AGPL-3.0-or-later
package models

type Language struct {
	Id         int64  `json:"-"`
	IsoName    string `json:"iso_name"`
	NativeName string `json:"native_name"`
	Code       string `json:"code"`
}
