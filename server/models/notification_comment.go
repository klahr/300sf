// SPDX-License-Identifier: AGPL-3.0-or-later
package models

type NotificationComment struct {
	NotificationId int64 `json:"-"`
	CommentId      int64 `json:"-"`
}
