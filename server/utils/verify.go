// SPDX-License-Identifier: AGPL-3.0-or-later
package utils

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/table"
)

func VerifyDisplayName(displayName string) bool {
	n := utf8.RuneCountInString(displayName)
	return n >= 4 && n <= 64 && !strings.ContainsAny(displayName, "<>")
}

func VerifyUsername(username string) bool {
	regex, _ := regexp.Compile(`^\w(?:[\w.#-]*\w){3,31}$`)
	return regex.MatchString(username)
}

func VerifyEmail(email string) bool {
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	if match, _ := regexp.MatchString(regex, email); !match {
		return false
	}

	var count int64
	models.DB.Table(table.Users).Where("email=?", email).Count(&count)

	return count <= 0
}

// VerifyPassword applies to new passwords; existing ones still sign in.
func VerifyPassword(password string) bool {
	return len(password) >= 8 && len(password) <= 1024
}

// VerifyPortrait accepts the name of one of the portraits in /img/portraits.
func VerifyPortrait(portrait string) bool {
	regex := regexp.MustCompile(`^[0-9]{1,3}\.(webp|png)$`)
	return regex.MatchString(portrait)
}
