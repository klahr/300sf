// SPDX-License-Identifier: GPL-2.0-or-later
package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/api/config"
	"github.com/klahr/300sf/api/log"
	"github.com/klahr/300sf/api/models"
	"github.com/klahr/300sf/api/services"
)

func main() {
	log.Initialize()
	log.Info(nil, 0, "starting server")

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		s := <-sigc
		print(s.String())
		log.Info(nil, 0, s.String())
		log.Info(nil, 0, "bye")
		os.Exit(0)
	}()

	configFilename := flag.String("config", "", "config filename")
	flag.Parse()
	if err := config.SetupAppConfig(*configFilename); err != nil {
		log.Error(nil, 0, err.Error())
		os.Exit(1)
	}

	if err := models.SetupDB(config.AppConfig.DataSource, config.GetLogLevelAsInt()); err != nil {
		log.Error(nil, 0, err.Error())
		os.Exit(1)
	}

	if err := models.SetupCache(config.AppConfig.RedisHost); err != nil {
		log.Error(nil, 0, err.Error())
		os.Exit(1)
	}

	gin.SetMode(config.AppConfig.Mode)
	r := gin.New()
	r.SetTrustedProxies([]string{})
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-App-Version,X-App-Language")
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
		go services.SendAllEmails(config.AppConfig.SendEmailInterval)
	}

	if config.AppConfig.DeleteAccountsInterval > 0 {
		go services.DeleteAccounts(config.AppConfig.DeleteAccountsInterval)
	}

	r.Run(config.AppConfig.Address)
}
