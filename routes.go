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
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error(), "code": 0})
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
		c.JSON(http.StatusUnauthorized, gin.H{"message": "missing_permission", "code": 0})
		c.Abort()
	}
}

func basicMiddleware(c *gin.Context) {
	requirePermission(c, auth.AuthPermissionBasic)
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
	if err := models.DB.Where("hash=?", hashedOrignalKey).Where("prefix=?", v[0]).First(&apiKey).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_api_key", "code": 2})
		c.Abort()
		return
	}

	if apiKey.RevokedAt.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_api_key", "code": 3})
		c.Abort()
		return
	}

	c.Next()
}

func Ping(c *gin.Context) {
	if param, found := c.GetQuery("param"); found {
		c.Writer.Write([]byte(param))
	} else {
		c.Writer.Write([]byte("pong"))
	}
	c.Writer.WriteHeader(http.StatusOK)
}

func SetupRoutes(r *gin.Engine) {
	// Authenticate
	r.POST("/v1/auth", controllers.Auth)

	// Users
	r.POST("/v1/users", controllers.Users)

	r.GET("/v1/test", apiKeyMiddleware)
	r.GET("ping", Ping)

	// Lists
	r.POST("/v1/lists", basicMiddleware, controllers.Lists)

	// List templates
	//	r.GET("/v1/list_templates/list", basicMiddleware, controllers.ListTemplatesList)
	r.POST("/v1/list_templates", basicMiddleware, controllers.ListTemplates)

	// Api keys
	r.GET("/v1/api_keys/list", basicMiddleware, controllers.ApiKeysList)
	r.POST("/v1/api_keys", basicMiddleware, controllers.ApiKeys)
	r.POST("/v1/api_keys/:prefix/revoke", basicMiddleware, controllers.ApiKeysRevoke)

	setupPrivateRoutes(r)
}

func setupPrivateRoutes(r *gin.Engine) {
	// Species
	r.POST("/v1/p/activities", adminMiddleware, controllers.CreateNamedRecord[models.Activity])
	r.GET("/v1/p/activities/list", adminMiddleware, controllers.NamedRecordList[models.Activity])
	r.POST("/v1/p/methods", adminMiddleware, controllers.CreateNamedRecord[models.Method])
	r.GET("/v1/p/methods/list", adminMiddleware, controllers.NamedRecordList[models.Method])
	r.POST("/v1/p/species", adminMiddleware, controllers.Species)
	r.POST("/v1/p/genders", adminMiddleware, controllers.CreateNamedRecord[models.Gender])
	r.GET("/v1/p/genders/list", adminMiddleware, controllers.NamedRecordList[models.Gender])
	r.POST("/v1/p/orders", adminMiddleware, controllers.CreateNamedRecord[models.Order])
	r.GET("/v1/p/orders/list", adminMiddleware, controllers.NamedRecordList[models.Order])
	r.GET("/v1/p/translations/list", adminMiddleware, controllers.TranslationList)
	r.POST("/v1/p/translations/update", adminMiddleware, controllers.TranslationUpdate)
}
