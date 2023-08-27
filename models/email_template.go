// SPDX-License-Identifier: GPL-2.0-or-later
package models

type EmailTemplate struct {
	Model
	FromEmail  string      `json:"from_email"`
	FromName   string      `json:"from_name"`
	SubjectId  int64       `json:"-"`
	Subject    Translation `json:"subject"`
	TextPartId int64       `json:"-"`
	TextPart   Translation `json:"text_part"`
	HtmlPartId int64       `json:"-"`
	HtmlPart   Translation `json:"html_part"`
}
