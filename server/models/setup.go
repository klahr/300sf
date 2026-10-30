// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/klahr/300sf/server/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var Cache *redis.Client

func SetupDB(dataSource string, logLevel int) error {
	database, err := gorm.Open(mysql.Open(dataSource), &gorm.Config{
		Logger: logger.Default.LogMode(logger.LogLevel(logLevel)),
	})
	if err != nil {
		log.Error(nil, 0, err.Error())
		return err
	}

	sqlDB, err := database.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	DB = database
	return nil
}

func SetupCache(address, password string) error {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       0,
	})

	Cache = client
	return nil
}
