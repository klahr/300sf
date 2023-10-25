// SPDX-License-Identifier: GPL-2.0-or-later
package controllers

import (
	"net/http"
	"os"
	"time"

	"git.sr.ht/~klahr/300sf/auth"
	"git.sr.ht/~klahr/300sf/config"
	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/table"
	"git.sr.ht/~klahr/300sf/utils"
	"github.com/gin-gonic/gin"
)

func Auth(c *gin.Context) {
	authorization := c.Request.Header.Get("Authorization")
	var user models.User

	if authorization != "" {
		email, password, ok := c.Request.BasicAuth()
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}

		if err := models.DB.Where("email=?", email).Preload("Language").First(&user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -30002,
				"message": "invalid email or password",
			})
			return
		}

		if user.DeactivatedAt.Valid {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    ErrorAccountDeactivated,
				"message": "account is deactivated",
			})
			return
		}

		p := utils.Password{}
		if !p.IsMatching(password, user.Password, user.PasswordSalt) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -30002,
				"message": "invalid email or password",
			})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	privateKey, err := os.ReadFile(config.AppConfig.PrivateKeyFilename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	authenticationResponse, err := auth.Authenticate(user.Uuid, user.Permissions, config.AppConfig.AccessTokenLifetime, config.AppConfig.RefreshTokenLifetime, privateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	models.DB.Table(table.Users).Where("id=?", user.Id).UpdateColumn("last_seen_at", time.Now())

	c.Header("Content-Language", user.Language.Code)
	c.JSON(http.StatusOK, SingleItemDataResponse{
		Data: authenticationResponse,
	})
}
