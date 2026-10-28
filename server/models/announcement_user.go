// SPDX-License-Identifier: AGPL-3.0-or-later
package models

type AnnouncementUser struct {
	AnnouncementId int64 `json:"-"`
	UserId         int64 `json:"-"`
}
