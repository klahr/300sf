// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import "time"

type Model struct {
	Id        int64     `json:"-"`
	Timestamp time.Time `json:"-"`
	Uuid      string    `json:"uuid"`
}
