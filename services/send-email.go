package services

import (
	"context"
	"time"

	"git.sr.ht/~klahr/300sf/config"
	"git.sr.ht/~klahr/300sf/log"
	"git.sr.ht/~klahr/300sf/models"

	sf_mailjet "git.sr.ht/~klahr/300sf/mailjet"
	"github.com/mailjet/mailjet-apiv3-go"
)

func SendAllEmails(ctx *context.Context, interval int64) {
	for {
		var emails []models.Email
		tx := models.DB.Begin()
		if err := tx.Where("send_at <= NOW()").Where("sent_at IS NULL").Find(&emails).Error; err != nil {
			tx.Rollback()
		} else {
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
							Name:  e.ToEmail,
						},
					},
					Subject:  e.Subject,
					TextPart: e.TextPart,
					HTMLPart: e.HtmlPart,
				}
				messagesInfo = append(messagesInfo, mi)
			}
			if len(messagesInfo) > 0 {
				if err := sf_mailjet.SendEmails(messagesInfo, config.AppConfig.MailjetPublicKey, config.AppConfig.MailjetSecretKey); err != nil {
					log.Error(ctx, 0, err.Error())
					tx.Rollback()
				} else {
					now := time.Now()
					for _, e := range emails {
						if err := tx.Model(e).Update("sent_at", now).Error; err != nil {
							log.Error(ctx, 0, err.Error())
						}
					}
					tx.Commit()
				}
			} else {
				tx.Commit()
			}
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}
