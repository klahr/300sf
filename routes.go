package main

import (
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"git.sr.ht/~klahr/300sf/auth"
	"git.sr.ht/~klahr/300sf/config"
	"git.sr.ht/~klahr/300sf/controllers"
	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/utils"
)

func requirePermission(c *gin.Context, requiredPermission int64) {
	// TODO: Cache this value and also use in ws.
	publicKey, _ := ioutil.ReadFile(config.AppConfig.PublicKeyFilename)

	xAccessToken := c.Request.Header.Get("X-Access-Token")
	accessToken := auth.Token{}
	if xAccessToken != "" {
		if err := auth.ParseToken(xAccessToken, &accessToken, publicKey); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "logged out", "code": 0})
			c.Abort()
			return
		}
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "logged out", "code": 0})
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
		c.JSON(http.StatusUnauthorized, gin.H{"message": "missing permission", "code": 0})
		c.Abort()
	}
}

func noneMiddleware(c *gin.Context) {
	requirePermission(c, auth.AuthPermissionNone)
}

func basicMiddleware(c *gin.Context) {
	requirePermission(c, auth.AuthPermissionBasic)
}

func adminMiddleware(c *gin.Context) {
	requirePermission(c, auth.AuthPermissionAdmin)
}

func languageMiddleware(c *gin.Context) {
	lang := c.Request.Header.Get("X-Language")
	if lang != "" {
		c.Set("language", lang)
	} else {
		c.Set("language", "en")
	}

	c.Next()
}

func apiKeyMiddleware(c *gin.Context) {
	xApiKey := c.Request.Header.Get("X-Api-Key")
	if xApiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "missing api key", "code": 0})
		c.Abort()
		return
	}
	v := strings.Split(xApiKey, ".")
	if len(v) != 2 {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid api key", "code": 1})
		c.Abort()
		return
	}

	var apiKey models.ApiKey
	hashedOrignalKey := utils.ComputeHmac256(v[1], "0b8879c940c5af8076712b19cc614f732086ad822d41572e39f93be31ef7f550")
	if err := models.DB.Where("hash=?", hashedOrignalKey).Where("prefix=?", v[0]).First(&apiKey).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid api key", "code": 2})
		c.Abort()
		return
	}

	if apiKey.RevokedAt.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid api key", "code": 3})
		c.Abort()
		return
	}

	callCounter, _ := models.Cache.Get(v[0]).Result()
	iCallCounter := utils.Str2Int(callCounter)
	if iCallCounter >= apiKey.HardLimit {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "too many calls on api key", "code": 4})
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
	r.GET("/v1/users/list", adminMiddleware, controllers.UsersGetUsers)
	r.GET("/v1/users/:uuid", adminMiddleware, controllers.UsersGetUser)
	r.GET("/v1/users/whoami", apiKeyMiddleware, basicMiddleware, controllers.UsersWhoAmI)
	r.POST("/v1/users/migrate", apiKeyMiddleware, basicMiddleware, controllers.UsersMigrate)
	r.GET("/v1/users/export", apiKeyMiddleware, basicMiddleware, controllers.UsersExportData)

	r.GET("ping", Ping)

	// Lists
	r.GET("/v1/lists/list", apiKeyMiddleware, basicMiddleware, controllers.ListsList)
	r.POST("/v1/lists", apiKeyMiddleware, basicMiddleware, controllers.Lists)
	r.POST("/v1/lists/:uuid/archive", apiKeyMiddleware, basicMiddleware, controllers.ListsArchive)
	r.POST("/v1/lists/:uuid/rename", apiKeyMiddleware, basicMiddleware, controllers.ListsRename)
	r.POST("/v1/lists/:uuid/timespan", apiKeyMiddleware, basicMiddleware, controllers.ListsTimespan)
	r.GET("/v1/lists/:uuid", apiKeyMiddleware, controllers.ListsDetails)
	r.GET("/v1/lists/:uuid/reports", apiKeyMiddleware, basicMiddleware, controllers.ListsReports)

	// List templates
	r.GET("/v1/list_templates/list", apiKeyMiddleware, noneMiddleware, controllers.ListTemplatesList)
	r.POST("/v1/list_templates", apiKeyMiddleware, basicMiddleware, controllers.ListTemplates)
	r.POST("/v1/list_templates/:uuid/archive", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesArchive)
	r.POST("/v1/list_templates/:uuid/add_species", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesAddSpecies)
	r.POST("/v1/list_templates/:uuid/remove_species", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesRemoveSpecies)
	r.POST("/v1/list_templates/:uuid/name", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesName)
	r.POST("/v1/list_templates/:uuid/description", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesDescription)
	r.GET("/v1/list_templates/:uuid", apiKeyMiddleware, noneMiddleware, controllers.ListTemplatesDetails)

	// Tags
	r.GET("/v1/tags/list", apiKeyMiddleware, basicMiddleware, controllers.TagsList)
	r.POST("/v1/tags", apiKeyMiddleware, basicMiddleware, controllers.Tags)
	r.POST("/v1/tags/:uuid/archive", apiKeyMiddleware, basicMiddleware, controllers.TagsArchive)
	r.POST("/v1/tags/:uuid/rename", apiKeyMiddleware, basicMiddleware, controllers.TagsRename)
	r.GET("/v1/tags/:uuid", apiKeyMiddleware, basicMiddleware, controllers.TagsDetails)

	// Api keys
	r.GET("/v1/api_keys/list", basicMiddleware, controllers.ApiKeysList)
	r.POST("/v1/api_keys", basicMiddleware, controllers.ApiKeys)
	r.POST("/v1/api_keys/:prefix/revoke", basicMiddleware, controllers.ApiKeysRevoke)

	// Species
	r.GET("/v1/species/list", apiKeyMiddleware, controllers.SpeciesList)
	r.GET("/v1/species/:uuid", languageMiddleware, apiKeyMiddleware, controllers.SpeciesDetails)

	// Reports
	r.GET("/v1/reports/list", apiKeyMiddleware, basicMiddleware, controllers.ReportsList)
	r.POST("/v1/reports", apiKeyMiddleware, basicMiddleware, controllers.Reports)
	r.POST("/v1/reports/:uuid/archive", apiKeyMiddleware, basicMiddleware, controllers.ReportsArchive)
	r.GET("/v1/reports/:uuid", apiKeyMiddleware, basicMiddleware, controllers.ReportsDetails)
	r.POST("/v1/reports/:uuid/remove_tag", apiKeyMiddleware, basicMiddleware, controllers.ReportsRemoveTag)
	r.POST("/v1/reports/:uuid/add_tag", apiKeyMiddleware, basicMiddleware, controllers.ReportsAddTag)
	r.GET("/v1/reports/basis", apiKeyMiddleware, basicMiddleware, controllers.ReportsBasis)

	// Activities
	r.GET("/v1/activities/list", apiKeyMiddleware, controllers.NamedRecordList[models.Activity])

	// Methods
	r.GET("/v1/methods/list", apiKeyMiddleware, controllers.NamedRecordList[models.Method])

	// Genders
	r.GET("/v1/genders/:uuid", apiKeyMiddleware, controllers.NamedRecordDetails[models.Gender])
	r.GET("/v1/genders/list", apiKeyMiddleware, controllers.NamedRecordList[models.Gender])

	// Orders
	r.GET("/v1/orders/list", apiKeyMiddleware, controllers.NamedRecordList[models.Order])
	r.GET("/v1/orders/:uuid", apiKeyMiddleware, controllers.NamedRecordDetails[models.Order])

	// Ages
	r.GET("/v1/ages/list", apiKeyMiddleware, controllers.NamedRecordList[models.Age])

	// Zones
	r.GET("/v1/zones/list", adminMiddleware, controllers.ZonesList)
	r.GET("/v1/zones/:uuid", adminMiddleware, controllers.NamedRecordDetails[models.Zone])

	// Notifications
	r.POST("/v1/notifications/:uuid/comment", apiKeyMiddleware, basicMiddleware, controllers.NotificationsComment)

	setupPrivateRoutes(r)
}

func setupPrivateRoutes(r *gin.Engine) {
	// Species
	r.POST("/v1/activities", adminMiddleware, controllers.CreateNamedRecord[models.Activity])
	r.POST("/v1/methods", adminMiddleware, controllers.CreateNamedRecord[models.Method])
	r.POST("/v1/species", adminMiddleware, controllers.Species)
	r.POST("/v1/genders", adminMiddleware, controllers.CreateNamedRecord[models.Gender])
	r.POST("/v1/orders", adminMiddleware, controllers.CreateNamedRecord[models.Order])
	r.GET("/v1/translations/list", adminMiddleware, controllers.TranslationList)
	r.POST("/v1/translations/update", adminMiddleware, controllers.TranslationUpdate)
	r.POST("/v1/ages", adminMiddleware, controllers.CreateNamedRecord[models.Age])

	// Zones
	r.POST("/v1/zones", adminMiddleware, controllers.Zones)
}
