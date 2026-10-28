// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"github.com/klahr/300sf/server/config"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mailjet/mailjet-apiv3-go"

	sf_mailjet "github.com/klahr/300sf/server/mailjet"
	"github.com/klahr/300sf/server/models"
)

func EmailsSendAll(c *gin.Context) {
	tx := models.DB.Begin()
	defer tx.Rollback()

	var emails []models.Email
	if err := tx.Where("send_at <= NOW()").Where("sent_at IS NULL").Find(&emails).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	var messagesInfo []mailjet.InfoMessagesV31
	for _, e := range emails {
		mi := mailjet.InfoMessagesV31{
			From: &mailjet.RecipientV31{
				Email: e.FromEmail,
				Name:  e.FromName,
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: e.ToEmail,
					Name:  e.ToName,
				},
			},
			ReplyTo: &mailjet.RecipientV31{
				Email: e.ReplyEmail,
				Name:  e.ReplyName,
			},
			Subject:  e.Subject,
			TextPart: e.TextPart,
			HTMLPart: e.HtmlPart,
		}
		messagesInfo = append(messagesInfo, mi)
	}
	if len(messagesInfo) > 0 {
		if err := sf_mailjet.SendEmails(messagesInfo, config.AppConfig.MailjetPublicKey, config.AppConfig.MailjetSecretKey); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		} else {
			now := time.Now()
			for _, e := range emails {
				if err := tx.Model(e).Update("sent_at", now).Error; err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"message": err.Error(),
					})
					return
				}
			}
			tx.Commit()
		}
	}

	type response struct {
		EmailsSent int `json:"emails_sent"`
	}

	c.JSON(http.StatusOK, response{
		EmailsSent: len(emails),
	})
}
