// SPDX-License-Identifier: GPL-2.0-or-later
package mailjet

import (
  "github.com/mailjet/mailjet-apiv3-go"
)

func SendEmails(messagesInfo []mailjet.InfoMessagesV31) error {
  c := mailjet.NewMailjetClient("", "")
  messages := mailjet.MessagesV31 {
    Info: messagesInfo,
  }
  _, err := c.SendMailV31(&messages)
  return err
}
