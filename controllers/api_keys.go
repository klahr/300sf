package controllers

import (
	"net/http"
	"time"

	"git.sr.ht/~klahr/300sf/config"
	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/table"
	"git.sr.ht/~klahr/300sf/utils"
	"github.com/gin-gonic/gin"
	"gopkg.in/guregu/null.v4/zero"
)

func ApiKeys(c *gin.Context) {
	type request struct {
		Name string `json:"name"`
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

	originalKey := utils.RandomSha256()
	hashedOriginalKey := utils.ComputeHmac256(originalKey, "0b8879c940c5af8076712b19cc614f732086ad822d41572e39f93be31ef7f550")

	prefix := utils.GenerateUuidWithFixedLength(7)
	apiKey := models.ApiKey{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		Name:      r.Name,
		Prefix:    prefix,
		Hash:      hashedOriginalKey,
		UserId:    user.Id,
		SoftLimit: config.AppConfig.DefaultApiSoftLimit,
		HardLimit: config.AppConfig.DefaultApiHardLimit,
	}

	if err := models.DB.Create(&apiKey).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	type response struct {
		ApiKey string `json:"api_key"`
	}
	c.JSON(http.StatusOK, response{
		ApiKey: prefix + "." + originalKey,
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
