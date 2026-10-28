// SPDX-License-Identifier: GPL-2.0-or-later
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	_ "time/tzdata" // the alpine image has no zoneinfo; reports export uses IANA zones

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/config"
	"github.com/klahr/300sf/server/controllers"
	"github.com/klahr/300sf/server/log"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/services"
	"github.com/klahr/300sf/server/site"
)

// apiHeaders marks API responses as JSON. The public API (/v1) allows calls
// from any origin; the app's API (/api) is same-origin only, since it is
// authenticated with cookies. Website pages are left alone.
func apiHeaders(c *gin.Context) {
	path := c.Request.URL.Path
	if !site.IsAPIPath(path) {
		c.Next()
		return
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Cache-Control", "no-store")
	if strings.HasPrefix(path, "/api/") {
		c.Next()
		return
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-App-Version,X-App-Language")
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
}

// newRouter serves the API (at the root and under /api) and the websites.
func newRouter(web *site.Site) *gin.Engine {
	r := gin.New()
	r.SetTrustedProxies([]string{})
	r.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		log.Error(c, 0, fmt.Sprintf("panic: %v", err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "internal server error", "code": controllers.ErrorInternalServerError})
	}))
	r.Use(web.Redirect)
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(apiHeaders)

	SetupRoutes(r, controllers.Auth, controllers.AuthLogout)
	SetupRoutes(r.Group("/api", web.AppSession), web.AppSignIn, web.AppSignOut)
	web.Register(r)
	return r
}

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
	// In debug mode, serve site/ from disk so CSS and app.js changes show up
	// without rebuilding the server.
	var siteFiles fs.FS
	if config.AppConfig.Mode == gin.DebugMode {
		if _, err := os.Stat("site/templates"); err == nil {
			siteFiles = os.DirFS("site")
		}
	}
	web, err := site.New(config.AppConfig.Site, siteFiles)
	if err != nil {
		log.Error(nil, 0, err.Error())
		os.Exit(1)
	}
	r := newRouter(web)

	if config.AppConfig.SendEmailInterval > 0 {
		go services.SendAllEmails(config.AppConfig.SendEmailInterval)
	}

	if config.AppConfig.DeleteAccountsInterval > 0 {
		go services.DeleteAccounts(config.AppConfig.DeleteAccountsInterval)
	}

	r.Run(config.AppConfig.Address)
}
