// SPDX-License-Identifier: GPL-2.0-or-later
package controllers

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"strings"
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
	user := models.User{}

	if authorization != "" {
		fields := strings.Fields(authorization)
		if len(fields) != 2 {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -30001,
				"message": "invalid authorization header",
			})
			return
		}

		usernameAndPassword, _ := base64.StdEncoding.DecodeString(fields[1])
		splitIndex := strings.Index(string(usernameAndPassword), ":")
		if splitIndex < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}

		var email = string(usernameAndPassword[0:splitIndex])
		var password = string(usernameAndPassword[splitIndex+1:])

		if err := models.DB.Where("email=?", email).First(&user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -30002,
				"message": "invalid email or password",
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
		publicKey, _ := ioutil.ReadFile(config.AppConfig.PublicKeyFilename)
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

	privateKey, err := ioutil.ReadFile(config.AppConfig.PrivateKeyFilename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}

	authenticationResponse := auth.Authenticate(user.Uuid, user.Permissions, config.AppConfig.AccessTokenLifetime, config.AppConfig.RefreshTokenLifetime, privateKey)

	models.DB.Table(table.Users).Where("id=?", user.Id).UpdateColumn("last_seen_at", time.Now())

	c.JSON(http.StatusOK, authenticationResponse)
}
