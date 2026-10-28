// SPDX-License-Identifier: GPL-2.0-or-later
package utils

import (
	"regexp"

	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/table"
)

func VerifyDisplayName(displayName string) bool {
	return len(displayName) >= 4
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

func VerifyPassword(password string) bool {
	return len(password) >= 4
}
