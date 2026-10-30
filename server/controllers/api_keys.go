// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/auth"
	"github.com/klahr/300sf/server/config"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/table"
	"github.com/klahr/300sf/server/utils"
	"gopkg.in/guregu/null.v4/zero"
)

const maxApiKeysPerUser = 10

func ApiKeys(c *gin.Context) {
	type request struct {
		Name string `json:"name"`
		// Optional; only scopes the user has, or just read for a read-only key.
		// Sign-in is still limited to the user's scopes.
		Scopes []string `json:"scopes"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	createApiKey(c, user, r.Name, r.Scopes)
}

// ApiKeysPassword makes an API key with the account's email and password, no
// token needed: signing in takes a key, so without this a new account could
// only get its first key from the website. Such keys get the user scope.
func ApiKeysPassword(c *gin.Context) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorParse, Message: err.Error()})
		return
	}
	if throttle(c, "api-key-password/"+ClientNetwork(c.ClientIP()), 10, time.Hour) {
		return
	}

	user, authErr := CheckPassword(r.Email, r.Password, c.ClientIP())
	if authErr != nil {
		c.JSON(authErr.Status, authErr.Body)
		return
	}
	createApiKey(c, user, r.Name, nil)
}

// createApiKey makes a key for user with the requested scopes (the user scope
// when none) and answers with its secret, which is shown only this once.
func createApiKey(c *gin.Context, user models.User, name string, requested []string) {
	if len(requested) == 0 {
		requested = []string{auth.AuthScopeUser}
	}
	for _, s := range requested {
		if intersectScopes([]string{s}, user.Scopes, user.Scopes) == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    ErrorMissingScope,
				"message": "you can only give a key scopes you have",
			})
			return
		}
	}
	// A read-only key with other scopes would not be read-only.
	if slices.Contains(requested, auth.AuthScopeRead) && slices.ContainsFunc(requested, func(s string) bool { return s != auth.AuthScopeRead }) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorMissingScope,
			"message": "a read-only key can have no other scopes",
		})
		return
	}
	slices.Sort(requested)
	scopes := strings.Join(slices.Compact(requested), " ")

	// Rate limits are per key, so unlimited keys would mean no limit at all.
	var active int64
	models.DB.Model(&models.ApiKey{}).Where("user_id=?", user.Id).Where("revoked_at IS NULL").Count(&active)
	if active >= maxApiKeysPerUser {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorTooManyApiKeys,
			"message": "too many api keys, revoke one first",
		})
		return
	}

	originalKey := utils.RandomSha256()
	hashedOriginalKey := utils.ComputeHmac256(originalKey, config.AppConfig.HashingSecret)

	prefix := utils.GenerateUuidWithFixedLength(7)
	apiKey := models.ApiKey{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		Name:      name,
		ClientId:  utils.GenerateUuidWithFixedLength(32),
		Prefix:    prefix,
		Hash:      hashedOriginalKey,
		UserId:    user.Id,
		SoftLimit: config.AppConfig.DefaultApiSoftLimit,
		HardLimit: config.AppConfig.DefaultApiHardLimit,
		Scopes:    scopes,
	}

	if err := models.DB.Create(&apiKey).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	// Signing in takes the client ID as well as the secret.
	type response struct {
		ApiKey   string `json:"api_key"`
		ClientId string `json:"client_id"`
		Prefix   string `json:"prefix"`
		Scopes   string `json:"scopes"`
	}
	c.JSON(http.StatusOK, response{
		ApiKey:   prefix + "." + originalKey,
		ClientId: apiKey.ClientId,
		Prefix:   prefix,
		Scopes:   scopes,
	})
}

func ApiKeysList(c *gin.Context) {
	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var apiKeys []models.ApiKeyClean
	if err := models.DB.Table(table.ApiKeys).Where("user_id=?", user.Id).Find(&apiKeys).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	for i := range apiKeys {
		apiKeys[i].Website = apiKeys[i].ClientId == config.AppConfig.Site.ClientId
	}

	type response struct {
		ApiKeys []models.ApiKeyClean `json:"api_keys"`
	}

	c.JSON(http.StatusOK, response{
		ApiKeys: apiKeys,
	})
}

func ApiKeysRevoke(c *gin.Context) {
	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var apiKey models.ApiKey
	if err := models.DB.Where("prefix=?", c.Param("prefix")).Where("user_id=?", user.Id).First(&apiKey).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// Revoking it would sign everyone out of the website.
	if apiKey.ClientId == config.AppConfig.Site.ClientId {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "this key is used by the website",
		})
		return
	}

	if err := models.DB.Model(&apiKey).Updates(models.ApiKey{
		RevokedAt: zero.TimeFrom(time.Now()),
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "could not revoke api key",
		})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
