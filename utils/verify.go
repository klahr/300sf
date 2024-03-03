// SPDX-License-Identifier: GPL-2.0-or-later
package utils

import (
	"regexp"

	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/table"
)

func VerifyDisplayName(displayName string) bool {
	return len(displayName) >= 4
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
