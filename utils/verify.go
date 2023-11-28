// SPDX-License-Identifier: GPL-2.0-or-later
package utils

import (
	"fmt"
	"regexp"

	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/table"
	"github.com/badoux/checkmail"
)

func VerifyDisplayName(displayName string) bool {
	return len(displayName) >= 4
}

func VerifyUsername(username string) bool {
	regex, err := regexp.Compile(`^\w(?:[\w.#-]*\w){3,31}$`)
	if err != nil {
		fmt.Println("Error compiling regex:", err)
		return false
	}
	return regex.MatchString(username)
}

func VerifyEmail(email string) bool {
	if checkmail.ValidateFormat(email) != nil {
		return false
	}

	var count int64
	models.DB.Table(table.Users).Where("email=?", email).Count(&count)

	return count <= 0
}

func VerifyPassword(password string) bool {
	return len(password) >= 4
}
