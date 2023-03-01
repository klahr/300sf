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

	if configFilename := flag.String("config", "", "config filename"); configFilename != nil {
		flag.Parse()
		if err := config.SetupAppConfig(*configFilename); err != nil {
			log.Error(ctx, 0, err.Error())
			panic(err)
		}
	} else {
		log.Error(ctx, 0, "no config file provided")
		panic("config file is required")
	}

	r := gin.New()
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

	if err := models.SetupDB(ctx, config.AppConfig.DataSource, config.GetLogLevelAsInt()); err != nil {
		log.Error(ctx, 0, err.Error())
		panic(err)
	}

	if err := models.SetupCache(ctx, config.AppConfig.RedisHost); err != nil {
		log.Error(ctx, 0, err.Error())
		panic(err)
	}

	SetupRoutes(r)

	r.Run(config.AppConfig.Address)
}
