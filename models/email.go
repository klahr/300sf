// SPDX-License-Identifier: GPL-2.0-or-later
package models

import (
	"time"

	"gopkg.in/guregu/null.v4/zero"
)

type Email struct {
	Model
	FromEmail string    `json:"from_email"`
	FromName  string    `json:"from_name"`
	ToEmail   string    `json:"to_email"`
	ToName    string    `json:"to_name"`
	Subject   string    `json:"subject"`
	TextPart  string    `json:"text_part"`
	HtmlPart  string    `json:"html_part"`
	SendAt    time.Time `json:"send_at"`
	SentAt    zero.Time `json:"sent_at"`
	Checksum  string    `json:"checksum"`
}
