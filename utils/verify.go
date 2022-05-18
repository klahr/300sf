package utils

import (
	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/table"
	"github.com/badoux/checkmail"
)

func VerifyDisplayName(displayName string) bool {
	if len(displayName) < 4 {
		return false
	}
	return true
}

func VerifyEmail(email string) bool {
	if checkmail.ValidateFormat(email) != nil {
		return false
	}

	var count int64
	models.DB.Table(table.Users).Where("email=?", email).Count(&count)
	if count > 0 {
		return false
	}

	return true
}
