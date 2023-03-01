package main

import (
	"io/ioutil"
	"net/http"
	"strings"

	"git.sr.ht/~klahr/300sf/auth"
	"git.sr.ht/~klahr/300sf/config"
	"git.sr.ht/~klahr/300sf/controllers"
	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/utils"
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

func editMiddleware(c *gin.Context) {
	requirePermission(c, auth.AuthPermissionEdit)
}

func adminMiddleware(c *gin.Context) {
	requirePermission(c, auth.AuthPermissionAdmin)
}

func apiKeyMiddleware(c *gin.Context) {
	xApiKey := c.Request.Header.Get("X-Api-Key")
	if xApiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_api_key", "code": 0})
		c.Abort()
		return
	}
	v := strings.Split(xApiKey, ".")
	if len(v) != 2 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_api_key", "code": 1})
		c.Abort()
		return
	}

	var apiKey models.ApiKey
	hashedOrignalKey := utils.ComputeHmac256(v[1], "0b8879c940c5af8076712b19cc614f732086ad822d41572e39f93be31ef7f550")
	if err := models.DB.Where("hash=?", hashedOrignalKey).First(&apiKey).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_api_key", "code": 2})
		c.Abort()
		return
	}

	if apiKey.DeactivatedAt.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_api_key", "code": 3})
		c.Abort()
		return
	}

	c.Next()
}

func SetupRoutes(r *gin.Engine) {
	// Authenticate
	r.POST("/v1/auth", controllers.Auth)

	// Users
	r.POST("/v1/users", controllers.User)

	r.GET("/v1/test", apiKeyMiddleware)

	r.POST("/v1/observation_lists", basicMiddleware, controllers.ObservationList)

	setupPrivateRoutes(r)
}

func setupPrivateRoutes(r *gin.Engine) {
	// API Key
	r.POST("/v1/p/api_key_tiers", adminMiddleware, controllers.ApiKeyTier)
	r.POST("/v1/p/api_keys", adminMiddleware, controllers.ApiKey)
	r.PUT("/v1/p/api_keys/:uuid/deactivate", editMiddleware, controllers.ApiKeyDeactivate)

	// Species
	r.POST("/v1/p/species", adminMiddleware, controllers.Species)
	r.POST("/v1/p/genders", adminMiddleware, controllers.Gender)
	r.POST("/v1/p/orders", adminMiddleware, controllers.Order)
}
