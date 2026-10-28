package services

import (
	"time"

	"github.com/klahr/300sf/api/config"
	"github.com/klahr/300sf/api/log"
	"github.com/klahr/300sf/api/models"
	"github.com/klahr/300sf/api/table"

	sf_mailjet "github.com/klahr/300sf/api/mailjet"
	"github.com/mailjet/mailjet-apiv3-go"
)

func SendAllEmails(interval int64) {
	for {
		var emails []models.Email
		if err := models.DB.Where("send_at <= NOW()").Where("sent_at IS NULL").Find(&emails).Error; err != nil {
			log.Error(nil, 0, err.Error())
		} else {
			// Sent one at a time and marked individually, so a single message
			// mailjet refuses cannot hold back the rest of the queue.
			for _, e := range emails {
				messagesInfo := []mailjet.InfoMessagesV31{
					{
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
						ReplyTo: &mailjet.RecipientV31{
							Email: e.ReplyEmail,
							Name:  e.ReplyName,
						},
						Subject:  e.Subject,
						TextPart: e.TextPart,
						HTMLPart: e.HtmlPart,
					},
				}

				if err := sf_mailjet.SendEmails(messagesInfo, config.AppConfig.MailjetPublicKey, config.AppConfig.MailjetSecretKey); err != nil {
					log.Error(nil, 0, "could not send email "+e.Uuid+": "+err.Error())
					continue
				}

				if err := models.DB.Table(table.Emails).Where("id=?", e.Id).UpdateColumn("sent_at", time.Now()).Error; err != nil {
					// Leaving sent_at unset would resend this message on every
					// pass, so make the failure loud.
					log.Error(nil, 0, "sent email "+e.Uuid+" but could not mark it sent: "+err.Error())
				}
			}
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}
