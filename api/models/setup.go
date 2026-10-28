// SPDX-License-Identifier: GPL-2.0-or-later
package models

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/klahr/300sf/api/log"
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

func SetupCache(address string) error {
	client := redis.NewClient(&redis.Options{
		Addr: address,
		DB:   0,
	})

	Cache = client
	return nil
}
