// SPDX-License-Identifier: GPL-2.0-or-later
package controllers

import (
	"net/http"
	"strings"
	"time"

	"git.sr.ht/~klahr/300sf/config"
	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func ValidateToken(token string) (*jwt.Token, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		secretKey := []byte(config.AppConfig.HashingSecret)
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !t.Valid {
		return nil, nil
	}
	return t, nil
}

func Auth(c *gin.Context) {
	type request struct {
		GrantType    string   `json:"grant_type"`
		ClientId     string   `json:"client_id"`
		ClientSecret string   `json:"client_secret"`
		Email        string   `json:"email"`
		Password     string   `json:"password"`
		RefreshToken string   `json:"refresh_token"`
		Scopes       []string `json:"scopes"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{
			Message: err.Error(),
			Code:    ErrorParse,
		})
		return
	}

	var userUuid string
	var lang string
	var scopes string
	if r.GrantType == "refresh_token" {
		if token, err := ValidateToken(r.RefreshToken); token != nil {
			t, _ := jwt.Parse(r.RefreshToken, func(token *jwt.Token) (interface{}, error) {
				secretKey := []byte(config.AppConfig.HashingSecret)
				return secretKey, nil
			})
			claims := t.Claims.(jwt.MapClaims)
			userUuid = claims["sub"].(string)
			scopes = claims["scp"].(string)

			var user models.User
			if err := models.DB.Where("uuid=?", userUuid).Preload("Language").First(&user).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    ErrorAuthInvalid,
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

			lang = user.Language.Code
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    ErrorAuthInvalid,
				"message": err.Error(),
			})
		}
	} else if r.GrantType == "password" {
		if r.ClientId == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "missing client_id", "code": ErrorMissingApiKey})
			c.Abort()
			return
		}
		v := strings.Split(r.ClientSecret, ".")
		if len(v) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid client_secret", "code": ErrorInvalidApiKey})
			c.Abort()
			return
		}

		var apiKey models.ApiKey
		hashedOrignalKey := utils.ComputeHmac256(v[1], config.AppConfig.HashingSecret)
		if err := models.DB.Where("hash=?", hashedOrignalKey).Where("prefix=?", v[0]).Where("client_id=?", r.ClientId).First(&apiKey).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid api key", "code": ErrorInvalidApiKey})
			c.Abort()
			return
		}

		if apiKey.RevokedAt.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "api key has been revoked", "code": ErrorInvalidApiKey})
			c.Abort()
			return
		}

		var user models.User
		if err := models.DB.Where("email=?", r.Email).Preload("Language").First(&user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    ErrorAuthInvalid,
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
		if !p.IsMatching(r.Password, user.Password, user.PasswordSalt) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    ErrorAuthInvalid,
				"message": "invalid email or password",
			})
			return
		}
		userUuid = user.Uuid
		lang = user.Language.Code

		apiKeyScopes := strings.Split(apiKey.Scopes, " ")
		userScopes := strings.Split(user.Scopes, " ")
		var first = true
		for _, s := range r.Scopes {
			for _, t := range userScopes {
				if s == t {
					for _, u := range apiKeyScopes {
						if s == u {
							if first {
								first = false
							} else {
								scopes += " "
							}
							scopes += s
						}
					}
				}
			}
		}
	} else {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{
			Message: "invalid grant_type",
			Code:    ErrorInvalidRequest,
		})
		return
	}

	accessToken, err := generateToken(userUuid, r.ClientId, scopes, time.Now().Add(time.Second*time.Duration(config.AppConfig.AccessTokenLifetime)).Unix())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    ErrorUnknown,
			"message": err.Error(),
		})
		return
	}
	refreshToken, err := generateToken(userUuid, r.ClientId, scopes, time.Now().Add(time.Second*time.Duration(config.AppConfig.RefreshTokenLifetime)).Unix())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    ErrorUnknown,
			"message": err.Error(),
		})
		return
	}

	response := struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		Scope        string `json:"scope"`
		Lang         string `json:"lang"`
	}{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    config.AppConfig.AccessTokenLifetime,
		Scope:        scopes,
		Lang:         lang,
	}
	c.JSON(http.StatusOK, response)
}

func generateToken(userUuid, clientId string, scopes string, expiry int64) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = userUuid
	claims["exp"] = expiry
	claims["scp"] = scopes
	claims["aud"] = clientId
	secretKey := []byte(config.AppConfig.HashingSecret)
	t, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return t, nil
}
