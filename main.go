// SPDX-License-Identifier: GPL-2.0-or-later
package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"git.sr.ht/~klahr/300sf/config"
	"git.sr.ht/~klahr/300sf/log"
	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/services"
	"github.com/gin-gonic/gin"
)

func main() {
	ctx := log.CreateContext()
	log.Initialize()
	log.Info(ctx, 0, "starting server")

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		s := <-sigc
		print(s.String())
		log.Info(ctx, 0, s.String())
		log.Info(ctx, 0, "bye")
		os.Exit(0)
	}()

	configFilename := flag.String("config", "", "config filename")
	flag.Parse()
	if err := config.SetupAppConfig(*configFilename); err != nil {
		log.Error(ctx, 0, err.Error())
		os.Exit(1)
	}

	if err := models.SetupDB(ctx, config.AppConfig.DataSource, config.GetLogLevelAsInt()); err != nil {
		log.Error(ctx, 0, err.Error())
		os.Exit(1)
	}

	if err := models.SetupCache(ctx, config.AppConfig.RedisHost); err != nil {
		log.Error(ctx, 0, err.Error())
		os.Exit(1)
	}

	gin.SetMode(config.AppConfig.Mode)
	r := gin.New()
	r.SetTrustedProxies([]string{})
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Access-Token,X-Refresh-Token,X-App-Version,X-App-Language,X-Api-Key")
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	SetupRoutes(r)

	if config.AppConfig.SendEmailInterval > 0 {
		go services.SendAllEmails(ctx, config.AppConfig.SendEmailInterval)
	}

	r.Run(config.AppConfig.Address)
}
