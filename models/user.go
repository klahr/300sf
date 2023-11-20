// SPDX-License-Identifier: GPL-2.0-or-later
package models

import (
	"strings"

	"gopkg.in/guregu/null.v4/zero"
)

type User struct {
	Model
	DisplayName            string      `json:"display_name"`
	Email                  string      `json:"-"`
	Password               string      `json:"-"`
	PasswordSalt           string      `json:"-"`
	Notifications          int64       `json:"notifications"`
	LastSeenAt             zero.Time   `json:"last_seen_at"`
	Confirmed_at           zero.Time   `json:"confirmed_at"`
	DeactivatedAt          zero.Time   `json:"deactivated_at"`
	DeletedAt              zero.Time   `json:"deleted_at"`
	AcceptedTermsAt        zero.Time   `json:"accepted_terms_at"`
	ImportedAt             zero.Time   `json:"imported_at"`
	LanguageId             int64       `json:"-"`
	Language               Language    `json:"language"`
	Portrait               zero.String `json:"portrait"`
	DeleteAccountRequestId zero.Int    `json:"-"`
	Scopes                 string      `json:"scopes"`
}

func (u *User) HasScope(scope string) bool {
	s := strings.Split(u.Scopes, " ")
	for i := range s {
		if s[i] == scope {
			return true
		}
	}
	return false
}
