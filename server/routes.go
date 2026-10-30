// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"

	"github.com/klahr/300sf/server/auth"
	"github.com/klahr/300sf/server/config"
	"github.com/klahr/300sf/server/controllers"
	"github.com/klahr/300sf/server/log"
	"github.com/klahr/300sf/server/models"
)

// parseToken returns the claims of a valid access token in an
// "Authorization: Bearer" header.
func parseToken(header string) jwt.MapClaims {
	v := strings.Split(header, " ")
	if len(v) != 2 || v[0] != "Bearer" {
		return nil
	}
	claims, ok := controllers.ValidateAccessToken(v[1])
	if !ok {
		return nil
	}
	return claims
}

func requireScope(c *gin.Context, requiredScope string) {
	xAccessToken := c.Request.Header.Get("Authorization")

	xAppLanguage := c.Request.Header.Get("X-App-Language")
	c.Set("app_language", xAppLanguage)
	c.Set("request", uuid.New().String())

	ok := requiredScope == auth.AuthScopeNone
	claims := parseToken(xAccessToken)
	if claims == nil {
		if requiredScope != auth.AuthScopeNone {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "missing scope", "code": controllers.ErrorAuthInvalid})
			c.Abort()
			return
		}
	} else {
		c.Set("uuid", claims["sub"])

		if requiredScope == "" {
			ok = true
		} else {
			scopes := strings.Split(claims["scp"].(string), " ")
			for _, s := range scopes {
				if s == requiredScope {
					ok = true
					break
				}
			}
		}
	}

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "missing scope", "code": controllers.ErrorMissingScope})
		c.Abort()
	}

	c.Next()
}

func noneMiddleware(c *gin.Context) {
	requireScope(c, auth.AuthScopeNone)
}

func basicMiddleware(c *gin.Context) {
	requireScope(c, auth.AuthScopeUser)
}

func adminMiddleware(c *gin.Context) {
	requireScope(c, auth.AuthScopeAdmin)
}

func apiKeyMiddleware(c *gin.Context) {
	claims := parseToken(c.Request.Header.Get("Authorization"))
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized", "code": controllers.ErrorAuthInvalid})
		c.Abort()
		return
	}

	var apiKey models.ApiKey
	if err := models.DB.Where("client_id=?", claims["aud"]).First(&apiKey).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid api key", "code": controllers.ErrorInvalidApiKey})
		c.Abort()
		return
	}

	if apiKey.RevokedAt.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid api key", "code": controllers.ErrorInvalidApiKey})
		c.Abort()
		return
	}

	// INCR is atomic, so concurrent requests can't slip past the limit. If
	// Redis is unavailable the request is let through rather than failing the
	// whole API.
	key := "/api-key/" + apiKey.Uuid
	calls, err := models.Cache.Incr(key).Result()
	if err != nil {
		log.Warning(c, 0, "rate limit counter unavailable: "+err.Error())
	} else {
		// Also repair a counter left without expiry, or the key would be locked forever.
		if calls == 1 || models.Cache.TTL(key).Val() < 0 {
			models.Cache.Expire(key, time.Second*time.Duration(config.AppConfig.ApiCallCounterTTL))
		}
		if calls > apiKey.HardLimit {
			log.Warning(c, controllers.ErrorTooManyApiCalls, "too many calls on api key")
			c.JSON(http.StatusTooManyRequests, gin.H{"message": "too many calls on api key", "code": controllers.ErrorTooManyApiCalls})
			c.Abort()
			return
		}
	}

	c.Next()
}

func Health(c *gin.Context) {
	type response struct {
		Version string `json:"version"`
	}

	res := response{
		Version: version,
	}

	c.JSON(http.StatusOK, controllers.GeneralSuccessResponse{
		Message: "healthy",
		Data:    res,
	})
}

// SetupRoutes registers the API on r. It is mounted twice: at /v1 for API
// clients, which sign in with controllers.Auth and get tokens back, and under
// /api for the app, which signs in with cookie-based handlers instead.
func SetupRoutes(r gin.IRouter, signIn, signOut gin.HandlerFunc) {
	// Authenticate
	r.POST("/v1/auth", signIn)
	r.POST("/v1/auth/logout", signOut)

	// Users
	r.POST("/v1/users", noneMiddleware, controllers.Users)
	r.GET("/v1/users/list", adminMiddleware, controllers.UsersGetUsers)
	r.GET("/v1/users/me", apiKeyMiddleware, basicMiddleware, controllers.UsersMe)
	r.GET("/v1/users/:uuid", adminMiddleware, controllers.UsersGetUser)
	r.GET("/v1/users/whoami", apiKeyMiddleware, basicMiddleware, controllers.UsersWhoAmI)
	r.POST("/v1/users/migrate", apiKeyMiddleware, basicMiddleware, controllers.UsersMigrate)
	r.GET("/v1/users/export", apiKeyMiddleware, basicMiddleware, controllers.UsersExportData)
	r.POST("/v1/users/language", apiKeyMiddleware, basicMiddleware, controllers.UsersLanguage)
	r.POST("/v1/users/portrait", apiKeyMiddleware, basicMiddleware, controllers.UsersPortrait)
	r.POST("/v1/users/reset_password", noneMiddleware, controllers.UsersResetPassword)
	r.POST("/v1/users/set_password/:token", noneMiddleware, controllers.UsersSetPassword)
	r.POST("/v1/users/request_delete", apiKeyMiddleware, basicMiddleware, controllers.UsersRequestDelete)
	r.DELETE("/v1/users/confirm_delete/:token", apiKeyMiddleware, basicMiddleware, controllers.UsersConfirmDelete)

	r.GET("/health", Health)

	// Competitions
	r.POST("/v1/competitions", apiKeyMiddleware, basicMiddleware, controllers.Competitions)
	r.GET("/v1/competitions/list", apiKeyMiddleware, basicMiddleware, controllers.CompetitionsList)
	r.GET("/v1/competitions/list/me", apiKeyMiddleware, basicMiddleware, controllers.CompetitionsListMe)
	r.GET("/v1/competitions/:uuid", apiKeyMiddleware, basicMiddleware, controllers.CompetitionsGet)
	r.PUT("/v1/competitions/:uuid/join", apiKeyMiddleware, basicMiddleware, controllers.CompetitionsJoin)
	r.PUT("/v1/competitions/:uuid/leave", apiKeyMiddleware, basicMiddleware, controllers.CompetitionsLeave)
	r.PUT("/v1/competitions/:uuid/invite", apiKeyMiddleware, basicMiddleware, controllers.CompetitionsInvite)
	r.PUT("/v1/competitions/:uuid/decline", apiKeyMiddleware, basicMiddleware, controllers.CompetitionsDecline)
	r.PUT("/v1/competitions/:uuid/archive", apiKeyMiddleware, basicMiddleware, controllers.CompetitionsArchive)
	r.GET("/v1/competitions/:uuid/result", noneMiddleware, controllers.CompetitionsResult)
	r.GET("/v1/competitions/basis", apiKeyMiddleware, basicMiddleware, controllers.CompetitionsBasis)
	r.GET("/v1/zones/regions", apiKeyMiddleware, basicMiddleware, controllers.ZonesRegions)
	r.GET("/v1/zones/search", apiKeyMiddleware, basicMiddleware, controllers.ZonesSearch)
	r.POST("/v1/zones/osm", apiKeyMiddleware, basicMiddleware, controllers.ZonesFromOsm)

	// Lists
	r.GET("/v1/lists/list", apiKeyMiddleware, basicMiddleware, controllers.ListsList)
	r.GET("/v1/lists/archived", apiKeyMiddleware, basicMiddleware, controllers.ListsArchived)
	r.POST("/v1/lists", apiKeyMiddleware, basicMiddleware, controllers.Lists)
	r.POST("/v1/lists/:uuid/archive", apiKeyMiddleware, basicMiddleware, controllers.ListsArchive)
	r.POST("/v1/lists/:uuid/restore", apiKeyMiddleware, basicMiddleware, controllers.ListsRestore)
	r.DELETE("/v1/lists/:uuid", apiKeyMiddleware, basicMiddleware, controllers.ListsDelete)
	r.POST("/v1/lists/:uuid/rename", apiKeyMiddleware, basicMiddleware, controllers.ListsRename)
	r.POST("/v1/lists/:uuid/color", apiKeyMiddleware, basicMiddleware, controllers.ListsColor)
	r.POST("/v1/lists/:uuid/timespan", apiKeyMiddleware, basicMiddleware, controllers.ListsTimespan)
	r.GET("/v1/lists/:uuid", apiKeyMiddleware, basicMiddleware, controllers.ListsDetails)
	r.GET("/v1/lists/:uuid/reports", apiKeyMiddleware, basicMiddleware, controllers.ListsReports)
	r.GET("/v1/lists/species/:species_uuid", apiKeyMiddleware, basicMiddleware, controllers.ListsForSpecies)
	r.POST("/v1/lists/compare", noneMiddleware, controllers.ListsCompare)

	// List templates
	r.GET("/v1/list_templates/list", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesList)
	r.GET("/v1/list_templates/list/me", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesListMe)
	r.POST("/v1/list_templates", apiKeyMiddleware, basicMiddleware, controllers.ListTemplates)
	r.POST("/v1/list_templates/:uuid/archive", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesArchive)
	r.POST("/v1/list_templates/:uuid/add_species", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesAddSpecies)
	r.POST("/v1/list_templates/:uuid/remove_species", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesRemoveSpecies)
	r.POST("/v1/list_templates/:uuid/name", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesName)
	r.POST("/v1/list_templates/:uuid/description", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesDescription)
	r.POST("/v1/list_templates/:uuid/species", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesSpecies)
	r.POST("/v1/list_templates/:uuid/visibility", apiKeyMiddleware, adminMiddleware, controllers.ListTemplatesVisibility)
	r.GET("/v1/list_templates/:uuid", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesDetails)

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
	r.GET("/v1/species/list", apiKeyMiddleware, basicMiddleware, controllers.SpeciesList)
	r.GET("/v1/species/:uuid", apiKeyMiddleware, noneMiddleware, controllers.SpeciesDetails)
	r.POST("/v1/species/:uuid/set_rarity", adminMiddleware, controllers.SpeciesSetRarity)

	// Reports
	r.GET("/v1/reports/list", apiKeyMiddleware, basicMiddleware, controllers.ReportsList)
	r.POST("/v1/reports/list", apiKeyMiddleware, basicMiddleware, controllers.ReportsListByList)
	r.POST("/v1/reports", apiKeyMiddleware, basicMiddleware, controllers.Reports)
	r.POST("/v1/reports/:uuid/archive", apiKeyMiddleware, basicMiddleware, controllers.ReportsArchive)
	r.DELETE("/v1/reports/:uuid", apiKeyMiddleware, basicMiddleware, controllers.ReportsDelete)
	r.GET("/v1/reports/:uuid", apiKeyMiddleware, basicMiddleware, controllers.ReportsDetails)
	r.POST("/v1/reports/:uuid/remove_tag", apiKeyMiddleware, basicMiddleware, controllers.ReportsRemoveTag)
	r.POST("/v1/reports/:uuid/add_tag", apiKeyMiddleware, basicMiddleware, controllers.ReportsAddTag)
	r.PUT("/v1/reports/basis", apiKeyMiddleware, basicMiddleware, controllers.ReportsBasis)
	r.PUT("/v1/reports/export", apiKeyMiddleware, basicMiddleware, controllers.ReportsExport)

	// Activities
	r.GET("/v1/activities/list", apiKeyMiddleware, basicMiddleware, controllers.NamedRecordList[models.Activity])

	// Methods
	r.GET("/v1/methods/list", apiKeyMiddleware, basicMiddleware, controllers.NamedRecordList[models.Method])

	// Genders
	r.GET("/v1/genders/:uuid", apiKeyMiddleware, basicMiddleware, controllers.NamedRecordDetails[models.Gender])
	r.GET("/v1/genders/list", apiKeyMiddleware, basicMiddleware, controllers.NamedRecordList[models.Gender])

	// Orders
	r.GET("/v1/orders/list", apiKeyMiddleware, basicMiddleware, controllers.NamedRecordList[models.Order])
	r.GET("/v1/orders/:uuid", apiKeyMiddleware, basicMiddleware, controllers.OrderDetails)

	// Ages
	r.GET("/v1/ages/list", apiKeyMiddleware, basicMiddleware, controllers.NamedRecordList[models.Age])

	// Zones
	r.GET("/v1/zones/list", adminMiddleware, controllers.ZonesList)
	r.GET("/v1/zones/:uuid", adminMiddleware, controllers.NamedRecordDetails[models.Zone])

	// Notifications
	r.POST("/v1/notifications/:uuid/comment", apiKeyMiddleware, basicMiddleware, controllers.NotificationsComment)

	// Contact
	r.POST("/v1/contact", apiKeyMiddleware, basicMiddleware, controllers.Contact)

	// Announcements
	r.PUT("/v1/announcements/:uuid/read", apiKeyMiddleware, basicMiddleware, controllers.AnnouncementsRead)
	r.GET("/v1/announcements/list", apiKeyMiddleware, basicMiddleware, controllers.AnnouncementsList)

	r.POST("/v1/analyze", apiKeyMiddleware, basicMiddleware, controllers.Analyze)
	r.PUT("/v1/analyze/resolve", apiKeyMiddleware, basicMiddleware, controllers.AnalyzeResolve)

	setupPrivateRoutes(r)
}

func setupPrivateRoutes(r gin.IRouter) {
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

	// Emails
	r.GET("/v1/emails/send_all", adminMiddleware, controllers.EmailsSendAll)

	// Announcements
	r.POST("/v1/announcements", adminMiddleware, controllers.Announcements)
	r.GET("/v1/announcements/all", adminMiddleware, controllers.AnnouncementsAll)
	r.PUT("/v1/announcements/:uuid/pause", adminMiddleware, controllers.AnnouncementsPause)
	r.PUT("/v1/announcements/:uuid/resume", adminMiddleware, controllers.AnnouncementsResume)
	r.DELETE("/v1/announcements/:uuid", adminMiddleware, controllers.AnnouncementsDelete)

	r.POST("/v1/invite", adminMiddleware, controllers.Invite)
}
