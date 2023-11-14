package controllers

import (
	"net/http"
	"time"

	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/table"
	"git.sr.ht/~klahr/300sf/utils"
	"github.com/gin-gonic/gin"
	"gopkg.in/guregu/null.v4/zero"
)

func Announcements(c *gin.Context) {
	type request struct {
		Name    string             `json:"name"`
		Content models.Translation `json:"content"`
		From    zero.Time          `json:"from" time_format:"2006-01-02T15:04Z"`
		Until   zero.Time          `json:"until" time_format:"2006-01-02T15:04Z"`
		Icon    string             `json:"icon"`
		Color   string             `json:"color"`
		OnHold  bool               `json:"on_hold"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{
			Message: err.Error(),
			Code:    ErrorParse,
		})
		return
	}

	tx := models.DB.Begin()

	content := r.Content
	content.Timestamp = time.Now()
	content.Uuid = utils.GenerateShortUuid()
	if err := tx.Create(&content).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{
			Message: err.Error(),
			Code:    ErrorInsert,
		})
		return
	}

	announcement := models.Announcement{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		Name:    r.Name,
		Content: content,
		From:    r.From,
		Until:   r.Until,
		Icon:    r.Icon,
		Color:   r.Color,
	}

	if r.OnHold {
		announcement.OnHold = zero.TimeFrom(time.Now())
	}

	if err := tx.Create(&announcement).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{
			Message: err.Error(),
			Code:    ErrorInsert,
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, GeneralSuccessResponse{
		Message: "success",
		Data:    announcement,
	})
}

func AnnouncementsPause(c *gin.Context) {
	if err := models.DB.Table(table.Announcements).Where("uuid=?", c.Param("uuid")).UpdateColumn("on_hold", time.Now()).Error; err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{
			Message: err.Error(),
			Code:    ErrorRecordNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, GeneralSuccessResponse{
		Message: "success",
	})
}

func AnnouncementsResume(c *gin.Context) {
	if err := models.DB.Table(table.Announcements).Where("uuid=?", c.Param("uuid")).UpdateColumn("on_hold", nil).Error; err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{
			Message: err.Error(),
			Code:    ErrorRecordNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, GeneralSuccessResponse{
		Message: "success",
	})
}

func AnnouncementsDelete(c *gin.Context) {
	if err := models.DB.Table(table.Announcements).Where("uuid=?", c.Param("uuid")).Delete(&models.Announcement{}).Error; err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{
			Message: err.Error(),
			Code:    ErrorInsert,
		})
		return
	}

	c.JSON(http.StatusOK, GeneralSuccessResponse{
		Message: "success",
	})
}

func AnnouncementsRead(c *gin.Context) {
	var announcement models.Announcement
	if err := models.DB.Where("uuid=?", c.Param("uuid")).First(&announcement).Error; err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{
			Message: err.Error(),
			Code:    ErrorRecordNotFound,
		})
		return
	}

	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{
			Message: err.Error(),
			Code:    ErrorRecordNotFound,
		})
		return
	}

	if err := models.DB.Table(table.Announcements).Where("uuid=?", c.Param("uuid")).UpdateColumn("on_hold", nil).Error; err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{
			Message: err.Error(),
			Code:    ErrorRecordNotFound,
		})
		return
	}

	if err := models.DB.Create(&models.AnnouncementUser{
		AnnouncementId: announcement.Id,
		UserId:         user.Id,
	}).Error; err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{
			Message: err.Error(),
			Code:    ErrorInsert,
		})
		return
	}

	c.JSON(http.StatusOK, GeneralSuccessResponse{
		Message: "success",
	})
}

func AnnouncementsList(c *gin.Context) {
	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var announcements []models.Announcement
	if err := models.DB.Table(table.Announcements).Preload("Content").Where("on_hold IS NULL").Where("`from` IS NULL OR `from`<=NOW()").Where("until IS NULL OR until>=NOW()").Where("announcement_users.announcement_id IS NULL").Joins("LEFT JOIN announcement_users ON announcement_users.announcement_id=announcements.id").Find(&announcements).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GeneralSuccessResponse{
		Message: "success",
		Data:    announcements,
	})
}

func AnnouncementsAll(c *gin.Context) {
	var announcements []models.Announcement
	if err := models.DB.Table(table.Announcements).Preload("Content").Find(&announcements).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GeneralSuccessResponse{
		Message: "success",
		Data:    announcements,
	})
}
