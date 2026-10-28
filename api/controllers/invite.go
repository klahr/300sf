package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/api/config"
	"github.com/klahr/300sf/api/models"
	"github.com/klahr/300sf/api/utils"
)

func Invite(c *gin.Context) {
	type request struct {
		Email        string `json:"email"`
		LanguageCode string `json:"language_code"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	tx := models.DB.Begin()
	defer tx.Rollback()

	et, err := utils.GetEmailTemplate("invite", r.LanguageCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	email := models.Email{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		FromEmail:  et.Email,
		ToEmail:    r.Email,
		FromName:   et.Name,
		Subject:    et.Subject,
		TextPart:   et.Txt,
		HtmlPart:   et.Html,
		SendAt:     time.Now(),
		ReplyEmail: et.Email,
		ReplyName:  et.Name,
	}

	email.Checksum = utils.ComputeHmac256(email.Subject+email.HtmlPart+email.SendAt.String(), config.AppConfig.HashingSecret)
	if err := tx.Create(&email).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	tx.Commit()
}
