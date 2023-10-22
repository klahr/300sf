package services

import (
	"context"
	"time"

	"git.sr.ht/~klahr/300sf/log"
	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/table"
	"gorm.io/gorm"
)

func DeleteAccounts(ctx *context.Context, interval int64) {
	for {
		var users []models.User
		tx := models.DB.Begin()
		if err := tx.Select("users.*").Where("delete_at<=NOW()").Joins("JOIN delete_account_requests ON delete_account_request_id=delete_account_requests.id").Find(&users).Error; err != nil {
			log.Error(ctx, 0, err.Error())
			tx.Rollback()
		} else {
			for _, u := range users {
				if err := deleteUser(u, tx); err != nil {
					log.Error(ctx, 0, err.Error())
				}
			}

			tx.Commit()
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func deleteUser(user models.User, tx *gorm.DB) error {
	if err := tx.Table(table.Users).Where("id=?", user.Id).Delete(&models.User{}).Error; err != nil {
		return err
	}
	return nil
}
