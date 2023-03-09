package main

import (
	"io/ioutil"
	"net/http"
	"strings"
	"time"

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
		c.JSON(http.StatusUnauthorized, gin.H{"message": "missing_api_key", "code": 0})
		c.Abort()
		return
	}
	v := strings.Split(xApiKey, ".")
	if len(v) != 2 {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid_api_key", "code": 1})
		c.Abort()
		return
	}

	var apiKey models.ApiKey
	hashedOrignalKey := utils.ComputeHmac256(v[1], "0b8879c940c5af8076712b19cc614f732086ad822d41572e39f93be31ef7f550")
	if err := models.DB.Where("hash=?", hashedOrignalKey).Where("prefix=?", v[0]).First(&apiKey).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid_api_key", "code": 2})
		c.Abort()
		return
	}

	if apiKey.RevokedAt.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid_api_key", "code": 3})
		c.Abort()
		return
	}

	callCounter, _ := models.Cache.Get(v[0]).Result()
	iCallCounter := utils.Str2Int(callCounter)
	if iCallCounter >= apiKey.HardLimit {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "too_many_calls_on_api_key", "code": 4})
		c.Abort()
		return
	}

	models.Cache.Set(v[0], iCallCounter+1, time.Hour*24)

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

	r.GET("ping", Ping)

	// Lists
	r.GET("/v1/lists/list", apiKeyMiddleware, basicMiddleware, controllers.ListsList)
	r.POST("/v1/lists", apiKeyMiddleware, basicMiddleware, controllers.Lists)
	r.POST("/v1/lists/:uuid/archive", apiKeyMiddleware, basicMiddleware, controllers.ListsArchive)
	r.POST("/v1/lists/:uuid/rename", apiKeyMiddleware, basicMiddleware, controllers.ListsRename)
	r.POST("/v1/lists/:uuid/timespan", apiKeyMiddleware, basicMiddleware, controllers.ListsTimespan)

	// List templates
	r.GET("/v1/list_templates/list", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesList)
	r.POST("/v1/list_templates", apiKeyMiddleware, basicMiddleware, controllers.ListTemplates)
	r.POST("/v1/list_templates/:uuid/archive", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesArchive)
	r.POST("/v1/list_templates/:uuid/add_species", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesAddSpecies)
	r.POST("/v1/list_templates/:uuid/remove_species", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesRemoveSpecies)
	r.POST("/v1/list_templates/:uuid/rename", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesRename)

	// Api keys
	r.GET("/v1/api_keys/list", basicMiddleware, controllers.ApiKeysList)
	r.POST("/v1/api_keys", basicMiddleware, controllers.ApiKeys)
	r.POST("/v1/api_keys/:prefix/revoke", basicMiddleware, controllers.ApiKeysRevoke)

	// Species
	r.GET("/v1/species/list", apiKeyMiddleware, basicMiddleware, controllers.SpeciesList)

	// Reports
	r.GET("/v1/reports/list", apiKeyMiddleware, basicMiddleware, controllers.ReportsList)
	r.POST("/v1/reports", apiKeyMiddleware, basicMiddleware, controllers.Reports)
	r.POST("/v1/reports/:uuid/archive", apiKeyMiddleware, basicMiddleware, controllers.ReportsArchive)

	setupPrivateRoutes(r)
}

func setupPrivateRoutes(r *gin.Engine) {
	// Species
	r.POST("/v1/p/activities", adminMiddleware, controllers.CreateNamedRecord[models.Activity])
	r.GET("/v1/p/activities/list", adminMiddleware, controllers.NamedRecordList[models.Activity])
	r.POST("/v1/p/methods", adminMiddleware, controllers.CreateNamedRecord[models.Method])
	r.GET("/v1/p/methods/list", adminMiddleware, controllers.NamedRecordList[models.Method])
	r.POST("/v1/p/species", adminMiddleware, controllers.Species)
	r.GET("/v1/p/species/list", adminMiddleware, controllers.SpeciesList)
	r.POST("/v1/p/genders", adminMiddleware, controllers.CreateNamedRecord[models.Gender])
	r.GET("/v1/p/genders/list", adminMiddleware, controllers.NamedRecordList[models.Gender])
	r.POST("/v1/p/orders", adminMiddleware, controllers.CreateNamedRecord[models.Order])
	r.GET("/v1/p/orders/list", adminMiddleware, controllers.NamedRecordList[models.Order])
	r.GET("/v1/p/translations/list", adminMiddleware, controllers.TranslationList)
	r.POST("/v1/p/translations/update", adminMiddleware, controllers.TranslationUpdate)
	r.POST("/v1/p/ages", adminMiddleware, controllers.CreateNamedRecord[models.Age])
	r.GET("/v1/p/ages/list", adminMiddleware, controllers.NamedRecordList[models.Age])
}
