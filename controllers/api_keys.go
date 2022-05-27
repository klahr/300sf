package controllers

import (
	"net/http"
	"time"

	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/utils"
	"github.com/gin-gonic/gin"
	"gopkg.in/guregu/null.v4/zero"
)

func ApiKey(c *gin.Context) {
	type request struct {
		Name           string `json:"name"`
		ApiKeyTierUuid string `json:"api_key_tier_uuid"`
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

	var apiKeyTier models.ApiKeyTier
	if err := models.DB.Where("uuid=?", r.ApiKeyTierUuid).First(&apiKeyTier).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	originalKey := utils.RandomSha256()
	hashedOriginalKey := utils.ComputeHmac256(originalKey, "0b8879c940c5af8076712b19cc614f732086ad822d41572e39f93be31ef7f550")

	prefix := utils.GenerateUuidWithFixedLength(7)
	apiKey := models.ApiKey{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateUuid(),
		},
		Name:         r.Name,
		Prefix:       prefix,
		Hash:         hashedOriginalKey,
		UserId:       user.Id,
		ApiKeyTierId: apiKeyTier.Id,
	}
	print(hashedOriginalKey)

	if err := models.DB.Create(&apiKey).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	type response struct {
		ApiKey models.ApiKey `json:"api_key"`
		Key    string        `json:"key"`
	}
	c.JSON(http.StatusOK, response{
		ApiKey: apiKey,
		Key:    originalKey,
	})
}

func ApiKeyDeactivate(c *gin.Context) {
	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var apiKey models.ApiKey
	if err := models.DB.Where("uuid=?", c.Param("uuid")).Where("user_id=?", user.Id).First(&apiKey).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "api key could not be found",
		})
		return
	}

	if err := models.DB.Model(&apiKey).Updates(models.ApiKey{
		DeactivatedAt: zero.TimeFrom(time.Now()),
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "could not deactivate api key",
		})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
