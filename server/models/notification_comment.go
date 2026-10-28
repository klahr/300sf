// SPDX-License-Identifier: GPL-2.0-or-later
package models

type NotificationComment struct {
	NotificationId int64 `json:"-"`
	CommentId      int64 `json:"-"`
}
