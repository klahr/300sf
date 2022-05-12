package models

import (
	"context"

	"git.sr.ht/~klahr/300sf/log"
	"github.com/go-redis/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var Cache *redis.Client

func SetupDB(ctx *context.Context, dataSource string, logLevel int) error {
	database, err := gorm.Open(mysql.Open(dataSource), &gorm.Config{
		Logger: logger.Default.LogMode(logger.LogLevel(logLevel)),
	})
	if err != nil {
		log.Error(ctx, 0, err.Error())
		return err
	}

	DB = database
	return nil
}

func SetupCache(ctx *context.Context, address string) error {
	client := redis.NewClient(&redis.Options{
		Addr: address,
		DB:   0,
	})

	Cache = client
	return nil
}
