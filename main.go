package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"git.sr.ht/~klahr/300sf/config"
	"github.com/gin-gonic/gin"
)

func main() {
	ctx := log.CreateContext()
	log.Initialize()
	log.Info(ctx, 0, "starting server")

	var taskUpdater1s chan bool
	var taskUpdater1m chan bool

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		taskUpdater1s <- true
		taskUpdater1m <- true
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

	SetupRoutes(r)

	r.Run("localhost:8080")
}
