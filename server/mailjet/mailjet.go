// SPDX-License-Identifier: AGPL-3.0-or-later
package mailjet

import (
	"github.com/mailjet/mailjet-apiv3-go"
)

func SendEmails(messagesInfo []mailjet.InfoMessagesV31, publicKey, secretKey string) error {
	c := mailjet.NewMailjetClient(publicKey, secretKey)
	messages := mailjet.MessagesV31{
		Info: messagesInfo,
	}
	_, err := c.SendMailV31(&messages)
	return err
}
