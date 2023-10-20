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
	type request struct {
		RefreshToken string `json:"refresh_token"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

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
	} else if r.RefreshToken != "" {
		refreshToken := auth.Token{}
		publicKey, _ := os.ReadFile(config.AppConfig.PublicKeyFilename)
		if err := auth.ParseToken(r.RefreshToken, &refreshToken, publicKey); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid_grant"})
			return
		}

		if err := models.DB.Where("uuid=?", refreshToken.Uuid).First(&user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
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
	c.JSON(http.StatusOK, authenticationResponse)
}
