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
		Name     string             `json:"name"`
		Content  models.Translation `json:"content"`
		From     zero.Time          `json:"from"`
		Until    zero.Time          `json:"until"`
		Priority int                `json:"priority"`
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
		Name:     r.Name,
		Content:  content,
		From:     r.From,
		Until:    r.Until,
		Priority: r.Priority,
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
