package services

import (
	"time"

	"github.com/klahr/300sf/server/log"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/table"
	"gorm.io/gorm"
)

func DeleteAccounts(interval int64) {
	for {
		var users []models.User
		if err := models.DB.Select("users.*").Where("delete_at<=NOW()").Joins("JOIN delete_account_requests ON delete_account_request_id=delete_account_requests.id").Find(&users).Error; err != nil {
			log.Error(nil, 0, err.Error())
		}

		for _, u := range users {
			// One transaction per user so a failure never leaves an account half deleted.
			if err := models.DB.Transaction(func(tx *gorm.DB) error {
				return deleteUser(u, tx)
			}); err != nil {
				log.Error(nil, 0, err.Error())
			}
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func deleteUser(user models.User, tx *gorm.DB) error {
	// reports.user_id is not guaranteed to cascade (see db.sql), so remove them first.
	if err := tx.Table(table.Reports).Where("user_id=?", user.Id).Delete(&models.Report{}).Error; err != nil {
		return err
	}
	if err := tx.Table(table.Users).Where("id=?", user.Id).Delete(&models.User{}).Error; err != nil {
		return err
	}
	return nil
}
