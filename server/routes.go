// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"net/http"
	"slices"
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

// requireScope lets the request through when the token has any of the
// scopes, or with none given, whether or not there is a token.
func requireScope(c *gin.Context, requiredScopes ...string) {
	xAccessToken := c.Request.Header.Get("Authorization")

	xAppLanguage := c.Request.Header.Get("X-App-Language")
	c.Set("app_language", xAppLanguage)
	c.Set("request", uuid.New().String())

	ok := len(requiredScopes) == 0
	claims := parseToken(xAccessToken)
	if claims == nil {
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "missing scope", "code": controllers.ErrorAuthInvalid})
			c.Abort()
			return
		}
	} else {
		c.Set("uuid", claims["sub"])

		scopes := strings.Split(claims["scp"].(string), " ")
		for _, s := range scopes {
			if slices.Contains(requiredScopes, s) {
				ok = true
				break
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
	requireScope(c)
}

func basicMiddleware(c *gin.Context) {
	requireScope(c, auth.AuthScopeUser)
}

// readMiddleware is for normal routes that only read, which read-only keys
// may call too.
func readMiddleware(c *gin.Context) {
	requireScope(c, auth.AuthScopeUser, auth.AuthScopeRead)
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

// ServerConfig tells clients what this server offers: the app gets the same
// from its page, other clients need it from the API. No token needed.
func ServerConfig(c *gin.Context) {
	type mapConfig struct {
		Tiles       string `json:"tiles"`
		Attribution string `json:"attribution"`
	}
	type features struct {
		Analyze      bool `json:"analyze"`
		Contact      bool `json:"contact"`
		RegionSearch bool `json:"region_search"`
	}
	type response struct {
		Version   string    `json:"version"`
		SourceUrl string    `json:"source_url"`
		Map       mapConfig `json:"map"`
		Features  features  `json:"features"`
	}

	cfg := config.AppConfig
	c.JSON(http.StatusOK, response{
		Version:   version,
		SourceUrl: cfg.Site.SourceUrl,
		Map:       mapConfig{Tiles: cfg.Site.MapTiles, Attribution: cfg.Site.MapAttribution},
		Features: features{
			Analyze:      cfg.Site.AnalyzeEndpoint != "",
			Contact:      cfg.ContactEmail != "",
			RegionSearch: cfg.GeocoderEndpoint != "",
		},
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
	r.GET("/v1/users/list", apiKeyMiddleware, adminMiddleware, controllers.UsersGetUsers)
	r.GET("/v1/users/me", apiKeyMiddleware, readMiddleware, controllers.UsersMe)
	r.GET("/v1/users/:uuid", apiKeyMiddleware, adminMiddleware, controllers.UsersGetUser)
	r.GET("/v1/users/whoami", apiKeyMiddleware, readMiddleware, controllers.UsersWhoAmI)
	r.GET("/v1/users/export", apiKeyMiddleware, readMiddleware, controllers.UsersExportData)
	r.POST("/v1/users/language", apiKeyMiddleware, basicMiddleware, controllers.UsersLanguage)
	r.POST("/v1/users/portrait", apiKeyMiddleware, basicMiddleware, controllers.UsersPortrait)
	r.POST("/v1/users/display_name", apiKeyMiddleware, basicMiddleware, controllers.UsersDisplayName)
	r.POST("/v1/users/username", apiKeyMiddleware, basicMiddleware, controllers.UsersUsername)
	r.POST("/v1/users/email", apiKeyMiddleware, basicMiddleware, controllers.UsersEmail)
	r.POST("/v1/users/reset_password", noneMiddleware, controllers.UsersResetPassword)
	r.POST("/v1/users/set_password/:token", noneMiddleware, controllers.UsersSetPassword)
	r.POST("/v1/users/request_delete", apiKeyMiddleware, basicMiddleware, controllers.UsersRequestDelete)
	r.DELETE("/v1/users/confirm_delete/:token", apiKeyMiddleware, basicMiddleware, controllers.UsersConfirmDelete)

	r.GET("/health", Health)
	r.GET("/v1/config", noneMiddleware, ServerConfig)

	// Competitions
	r.POST("/v1/competitions", apiKeyMiddleware, basicMiddleware, controllers.Competitions)
	r.GET("/v1/competitions/list", apiKeyMiddleware, readMiddleware, controllers.CompetitionsList)
	r.GET("/v1/competitions/list/me", apiKeyMiddleware, readMiddleware, controllers.CompetitionsListMe)
	r.GET("/v1/competitions/:uuid", apiKeyMiddleware, readMiddleware, controllers.CompetitionsGet)
	r.PUT("/v1/competitions/:uuid/join", apiKeyMiddleware, basicMiddleware, controllers.CompetitionsJoin)
	r.PUT("/v1/competitions/:uuid/leave", apiKeyMiddleware, basicMiddleware, controllers.CompetitionsLeave)
	r.PUT("/v1/competitions/:uuid/invite", apiKeyMiddleware, basicMiddleware, controllers.CompetitionsInvite)
	r.PUT("/v1/competitions/:uuid/decline", apiKeyMiddleware, basicMiddleware, controllers.CompetitionsDecline)
	r.PUT("/v1/competitions/:uuid/archive", apiKeyMiddleware, basicMiddleware, controllers.CompetitionsArchive)
	r.GET("/v1/competitions/:uuid/result", noneMiddleware, controllers.CompetitionsResult)
	r.GET("/v1/competitions/basis", apiKeyMiddleware, readMiddleware, controllers.CompetitionsBasis)
	r.GET("/v1/zones/regions", apiKeyMiddleware, readMiddleware, controllers.ZonesRegions)
	r.GET("/v1/zones/search", apiKeyMiddleware, readMiddleware, controllers.ZonesSearch)
	r.POST("/v1/zones/osm", apiKeyMiddleware, basicMiddleware, controllers.ZonesFromOsm)

	// Lists
	r.GET("/v1/lists/list", apiKeyMiddleware, readMiddleware, controllers.ListsList)
	r.GET("/v1/lists/archived", apiKeyMiddleware, readMiddleware, controllers.ListsArchived)
	r.POST("/v1/lists", apiKeyMiddleware, basicMiddleware, controllers.Lists)
	r.POST("/v1/lists/:uuid/archive", apiKeyMiddleware, basicMiddleware, controllers.ListsArchive)
	r.POST("/v1/lists/:uuid/restore", apiKeyMiddleware, basicMiddleware, controllers.ListsRestore)
	r.DELETE("/v1/lists/:uuid", apiKeyMiddleware, basicMiddleware, controllers.ListsDelete)
	r.POST("/v1/lists/:uuid/rename", apiKeyMiddleware, basicMiddleware, controllers.ListsRename)
	r.POST("/v1/lists/:uuid/color", apiKeyMiddleware, basicMiddleware, controllers.ListsColor)
	r.POST("/v1/lists/:uuid/timespan", apiKeyMiddleware, basicMiddleware, controllers.ListsTimespan)
	r.POST("/v1/lists/:uuid/visibility", apiKeyMiddleware, basicMiddleware, controllers.ListsVisibility)
	r.GET("/v1/lists/:uuid", apiKeyMiddleware, readMiddleware, controllers.ListsDetails)
	r.GET("/v1/lists/:uuid/reports", apiKeyMiddleware, readMiddleware, controllers.ListsReports)
	r.GET("/v1/lists/species/:species_uuid", apiKeyMiddleware, readMiddleware, controllers.ListsForSpecies)
	r.POST("/v1/lists/compare", noneMiddleware, controllers.ListsCompare)

	// List templates
	r.GET("/v1/list_templates/list", apiKeyMiddleware, readMiddleware, controllers.ListTemplatesList)
	r.GET("/v1/list_templates/list/me", apiKeyMiddleware, readMiddleware, controllers.ListTemplatesListMe)
	r.POST("/v1/list_templates", apiKeyMiddleware, basicMiddleware, controllers.ListTemplates)
	r.POST("/v1/list_templates/:uuid/archive", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesArchive)
	r.POST("/v1/list_templates/:uuid/add_species", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesAddSpecies)
	r.POST("/v1/list_templates/:uuid/remove_species", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesRemoveSpecies)
	r.POST("/v1/list_templates/:uuid/name", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesName)
	r.POST("/v1/list_templates/:uuid/description", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesDescription)
	r.POST("/v1/list_templates/:uuid/species", apiKeyMiddleware, basicMiddleware, controllers.ListTemplatesSpecies)
	r.POST("/v1/list_templates/:uuid/visibility", apiKeyMiddleware, adminMiddleware, controllers.ListTemplatesVisibility)
	r.GET("/v1/list_templates/:uuid", apiKeyMiddleware, readMiddleware, controllers.ListTemplatesDetails)

	// Tags
	r.GET("/v1/tags/list", apiKeyMiddleware, readMiddleware, controllers.TagsList)
	r.POST("/v1/tags", apiKeyMiddleware, basicMiddleware, controllers.Tags)
	r.POST("/v1/tags/:uuid/archive", apiKeyMiddleware, basicMiddleware, controllers.TagsArchive)
	r.POST("/v1/tags/:uuid/rename", apiKeyMiddleware, basicMiddleware, controllers.TagsRename)
	r.GET("/v1/tags/:uuid", apiKeyMiddleware, readMiddleware, controllers.TagsDetails)

	// Api keys
	r.GET("/v1/api_keys/list", apiKeyMiddleware, readMiddleware, controllers.ApiKeysList)
	r.POST("/v1/api_keys", apiKeyMiddleware, basicMiddleware, controllers.ApiKeys)
	r.POST("/v1/api_keys/password", noneMiddleware, controllers.ApiKeysPassword)
	r.POST("/v1/api_keys/:prefix/revoke", apiKeyMiddleware, basicMiddleware, controllers.ApiKeysRevoke)

	// Species
	r.GET("/v1/species/list", apiKeyMiddleware, readMiddleware, controllers.SpeciesList)
	r.GET("/v1/species/:uuid", apiKeyMiddleware, noneMiddleware, controllers.SpeciesDetails)
	r.POST("/v1/species/:uuid/set_rarity", apiKeyMiddleware, adminMiddleware, controllers.SpeciesSetRarity)

	// Reports
	r.GET("/v1/reports/list", apiKeyMiddleware, readMiddleware, controllers.ReportsList)
	r.POST("/v1/reports/list", apiKeyMiddleware, basicMiddleware, controllers.ReportsListByList)
	r.POST("/v1/reports", apiKeyMiddleware, basicMiddleware, controllers.Reports)
	r.POST("/v1/reports/:uuid/archive", apiKeyMiddleware, basicMiddleware, controllers.ReportsArchive)
	r.POST("/v1/reports/:uuid/add_list", apiKeyMiddleware, basicMiddleware, controllers.ReportsAddList)
	r.POST("/v1/reports/:uuid/remove_list", apiKeyMiddleware, basicMiddleware, controllers.ReportsRemoveList)
	r.DELETE("/v1/reports/:uuid", apiKeyMiddleware, basicMiddleware, controllers.ReportsDelete)
	r.PUT("/v1/reports/:uuid", apiKeyMiddleware, basicMiddleware, controllers.ReportsUpdate)
	r.GET("/v1/reports/:uuid", apiKeyMiddleware, readMiddleware, controllers.ReportsDetails)
	r.POST("/v1/reports/:uuid/remove_tag", apiKeyMiddleware, basicMiddleware, controllers.ReportsRemoveTag)
	r.POST("/v1/reports/:uuid/add_tag", apiKeyMiddleware, basicMiddleware, controllers.ReportsAddTag)
	r.PUT("/v1/reports/basis", apiKeyMiddleware, basicMiddleware, controllers.ReportsBasis)
	r.PUT("/v1/reports/export", apiKeyMiddleware, basicMiddleware, controllers.ReportsExport)

	// Activities
	r.GET("/v1/activities/list", apiKeyMiddleware, readMiddleware, controllers.NamedRecordList[models.Activity])

	// Methods
	r.GET("/v1/methods/list", apiKeyMiddleware, readMiddleware, controllers.NamedRecordList[models.Method])

	// Genders
	r.GET("/v1/genders/:uuid", apiKeyMiddleware, readMiddleware, controllers.NamedRecordDetails[models.Gender])
	r.GET("/v1/genders/list", apiKeyMiddleware, readMiddleware, controllers.NamedRecordList[models.Gender])

	// Orders
	r.GET("/v1/orders/list", apiKeyMiddleware, readMiddleware, controllers.NamedRecordList[models.Order])
	r.GET("/v1/orders/:uuid", apiKeyMiddleware, readMiddleware, controllers.OrderDetails)

	// Ages
	r.GET("/v1/ages/list", apiKeyMiddleware, readMiddleware, controllers.NamedRecordList[models.Age])

	// Zones
	r.GET("/v1/zones/list", apiKeyMiddleware, adminMiddleware, controllers.ZonesList)
	r.GET("/v1/zones/:uuid", apiKeyMiddleware, adminMiddleware, controllers.NamedRecordDetails[models.Zone])

	// Notifications
	r.POST("/v1/notifications/:uuid/comment", apiKeyMiddleware, basicMiddleware, controllers.NotificationsComment)

	// Contact
	r.POST("/v1/contact", apiKeyMiddleware, basicMiddleware, controllers.Contact)

	// Announcements
	r.PUT("/v1/announcements/:uuid/read", apiKeyMiddleware, basicMiddleware, controllers.AnnouncementsRead)
	r.GET("/v1/announcements/list", apiKeyMiddleware, readMiddleware, controllers.AnnouncementsList)
	r.GET("/v1/announcements/feed", apiKeyMiddleware, readMiddleware, controllers.AnnouncementsFeed)
	r.PUT("/v1/announcements/read_all", apiKeyMiddleware, basicMiddleware, controllers.AnnouncementsReadAll)

	r.POST("/v1/analyze", apiKeyMiddleware, basicMiddleware, controllers.Analyze)
	r.PUT("/v1/analyze/resolve", apiKeyMiddleware, basicMiddleware, controllers.AnalyzeResolve)

	setupPrivateRoutes(r)
}

func setupPrivateRoutes(r gin.IRouter) {
	// Species
	r.POST("/v1/activities", apiKeyMiddleware, adminMiddleware, controllers.CreateNamedRecord[models.Activity])
	r.POST("/v1/methods", apiKeyMiddleware, adminMiddleware, controllers.CreateNamedRecord[models.Method])
	r.POST("/v1/species", apiKeyMiddleware, adminMiddleware, controllers.Species)
	r.POST("/v1/genders", apiKeyMiddleware, adminMiddleware, controllers.CreateNamedRecord[models.Gender])
	r.POST("/v1/orders", apiKeyMiddleware, adminMiddleware, controllers.CreateNamedRecord[models.Order])
	r.GET("/v1/translations/list", apiKeyMiddleware, adminMiddleware, controllers.TranslationList)
	r.POST("/v1/translations/update", apiKeyMiddleware, adminMiddleware, controllers.TranslationUpdate)
	r.POST("/v1/ages", apiKeyMiddleware, adminMiddleware, controllers.CreateNamedRecord[models.Age])

	// Zones
	r.POST("/v1/zones", apiKeyMiddleware, adminMiddleware, controllers.Zones)

	// Emails
	r.GET("/v1/emails/send_all", apiKeyMiddleware, adminMiddleware, controllers.EmailsSendAll)

	// Announcements
	r.POST("/v1/announcements", apiKeyMiddleware, adminMiddleware, controllers.Announcements)
	r.GET("/v1/announcements/all", apiKeyMiddleware, adminMiddleware, controllers.AnnouncementsAll)
	r.PUT("/v1/announcements/:uuid", apiKeyMiddleware, adminMiddleware, controllers.AnnouncementsUpdate)
	r.PUT("/v1/announcements/:uuid/pause", apiKeyMiddleware, adminMiddleware, controllers.AnnouncementsPause)
	r.PUT("/v1/announcements/:uuid/resume", apiKeyMiddleware, adminMiddleware, controllers.AnnouncementsResume)
	r.DELETE("/v1/announcements/:uuid", apiKeyMiddleware, adminMiddleware, controllers.AnnouncementsDelete)

	r.POST("/v1/invite", apiKeyMiddleware, adminMiddleware, controllers.Invite)
}

// setupConnectRoutes lets apps such as Claude connect to accounts: OAuth
// (controllers/oauth.go) and the MCP endpoint (mcp.go). Users approve apps
// on /oauth/authorize, a page of the website, which answers through app, the
// API with the website's session.
func setupConnectRoutes(r *gin.Engine, app gin.IRouter) {
	r.GET("/.well-known/oauth-authorization-server", controllers.OAuthServerMetadata)
	r.GET("/.well-known/oauth-protected-resource", controllers.OAuthResourceMetadata)
	r.GET("/.well-known/oauth-protected-resource/mcp", controllers.OAuthResourceMetadata)
	r.POST("/oauth/register", controllers.OAuthRegister)
	r.POST("/oauth/token", controllers.OAuthToken)
	r.POST("/oauth/revoke", controllers.OAuthRevoke)
	app.POST("/v1/oauth/approve", apiKeyMiddleware, basicMiddleware, controllers.OAuthApprove)

	mcp := newMcpHandler(r)
	r.POST("/mcp", mcp)
	r.GET("/mcp", mcp)
	r.DELETE("/mcp", mcp)
}
