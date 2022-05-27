package controllers

import (
	"net/http"
	"time"

	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/utils"
	"github.com/gin-gonic/gin"
)

func ApiKeyTier(c *gin.Context) {
	type request struct {
		Name      string `json:"name"`
		SoftLimit int64  `json:"soft_limit"`
		HardLimit int64  `json:"hard_limit"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	apiKeyTier := models.ApiKeyTier{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateUuid(),
		},
		Name:      r.Name,
		SoftLimit: r.SoftLimit,
		HardLimit: r.HardLimit,
	}

	if err := models.DB.Create(&apiKeyTier).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	if err := models.DB.Create(&apiKeyTier).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, apiKeyTier)
}
