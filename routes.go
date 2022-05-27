package main

import (
	"io/ioutil"
	"net/http"

	"git.sr.ht/~klahr/300sf/auth"
	"git.sr.ht/~klahr/300sf/config"
	"git.sr.ht/~klahr/300sf/controllers"
	"github.com/gin-gonic/gin"
)

func requirePermission(c *gin.Context, requiredPermission int64) {
	// TODO: Cache this value and also use in ws.
	publicKey, _ := ioutil.ReadFile(config.AppConfig.PublicKeyFilename)

	xAccessToken := c.Request.Header.Get("X-Access-Token")
	accessToken := auth.Token{}
	if err := auth.ParseToken(xAccessToken, &accessToken, publicKey); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "code": 0})
		c.Abort()
		return
	}

	// xAppLanguage := c.Request.Header.Get("X-App-Language")

	if accessToken.Permissions&requiredPermission == requiredPermission {
		c.Set("uuid", accessToken.Uuid)
		// c.Set("app_language", xAppLanguage)
		c.Set("permissions", accessToken.Permissions)
		c.Next()
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_permission", "code": 0})
		c.Abort()
	}
}

func basicMiddleware(c *gin.Context) {
	requirePermission(c, auth.AuthPermissionBasic)
}

func adminMiddleware(c *gin.Context) {
	requirePermission(c, auth.AuthPermissionAdmin)
}

func SetupRoutes(r *gin.Engine) {
	// Authenticate
	r.POST("/v1/auth", controllers.Auth)

	// Users
	r.POST("/v1/users", controllers.User)

	// API Key
	r.POST("/v1/api_key", basicMiddleware, controllers.ApiKey)

	setupPrivateRoutes(r)
}

func setupPrivateRoutes(r *gin.Engine) {
	r.POST("/v1/p/genders", adminMiddleware, controllers.Gender)

	r.POST("/v1/p/api_key_tiers", adminMiddleware, controllers.ApiKeyTier)
}
