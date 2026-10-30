package controllers

import (
	"html"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/config"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/utils"
)

func Contact(c *gin.Context) {
	type request struct {
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	if config.AppConfig.ContactEmail == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "the contact form is off",
		})
		return
	}

	if throttle(c, "contact/"+c.GetString("uuid"), 5, time.Hour) {
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

	email := models.Email{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		FromEmail:  config.AppConfig.ContactEmail,
		ToEmail:    config.AppConfig.ContactEmail,
		FromName:   user.DisplayName,
		ToName:     "300sf",
		Subject:    r.Subject,
		TextPart:   r.Message,
		HtmlPart:   html.EscapeString(r.Message),
		SendAt:     time.Now(),
		ReplyEmail: user.Email,
		ReplyName:  user.DisplayName,
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

	c.JSON(http.StatusCreated, nil)
}
