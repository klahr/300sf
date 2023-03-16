package models

type NotificationComment struct {
	NotificationId int64 `json:"-"`
	CommentId      int64 `json:"-"`
}
