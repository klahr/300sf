// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/log"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/utils"
)

func NotificationsComment(c *gin.Context) {
	type request struct {
		Text string `json:"text"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		log.Error(c, 0, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not parse request",
			"code":    ErrorParse,
		})
		return
	}

	tx := models.DB.Begin()
	defer tx.Rollback()

	var user models.User
	if err := tx.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	// Only on notifications about the user's own reports.
	var notification models.Notification
	if err := tx.Joins("JOIN reports ON reports.id=notifications.report_id").Where("notifications.uuid=?", c.Param("uuid")).Where("reports.user_id=?", user.Id).First(&notification).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	comment := models.Comment{
		Model: models.Model{
			Uuid:      utils.GenerateShortUuid(),
			Timestamp: time.Now(),
		},
		UserId: user.Id,
		Text:   r.Text,
	}

	if err := tx.Create(&comment).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	noticationComment := models.NotificationComment{
		NotificationId: notification.Id,
		CommentId:      comment.Id,
	}

	if err := tx.Create(&noticationComment).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, comment)
}
