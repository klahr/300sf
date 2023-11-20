// SPDX-License-Identifier: GPL-2.0-or-later
package controllers

import (
	"net/http"
	"os"
	"strings"
	"time"

	"git.sr.ht/~klahr/300sf/auth"
	"git.sr.ht/~klahr/300sf/config"
	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/table"
	"git.sr.ht/~klahr/300sf/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func Auth(c *gin.Context) {
	authorization := c.Request.Header.Get("Authorization")
	var user models.User

	if authorization != "" {
		email, password, ok := c.Request.BasicAuth()
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    ErrorInvalidRequest,
				"message": "invalid request",
			})
			return
		}

		if err := models.DB.Where("email=?", email).Preload("Language").First(&user).Error; err != nil {
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
		if !p.IsMatching(password, user.Password, user.PasswordSalt) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    ErrorAuthInvalid,
				"message": "invalid email or password",
			})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorMissingHeader,
			"message": "invalid request",
		})
		return
	}

	privateKey, err := os.ReadFile(config.AppConfig.PrivateKeyFilename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    ErrorMissingFile,
			"message": "internal error",
		})
		return
	}

	authenticationResponse, err := auth.Authenticate(user.Uuid, user.Permissions, config.AppConfig.AccessTokenLifetime, privateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    ErrorUnknown,
			"message": "internal error",
		})
		return
	}

	models.DB.Table(table.Users).Where("id=?", user.Id).UpdateColumn("last_seen_at", time.Now())

	authenticationResponse.Lang = user.Language.Code
	c.JSON(http.StatusOK, SingleItemDataResponse{
		Data: authenticationResponse,
	})
}

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

func Oauth2(c *gin.Context) {
	type request struct {
		GrantType    string `json:"grant_type"`
		ClientId     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Email        string `json:"email"`
		Password     string `json:"password"`
		RefreshToken string `json:"refresh_token"`
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
	if r.GrantType == "refresh_token" {
		if token, err := ValidateToken(r.RefreshToken); token == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    ErrorAuthInvalid,
				"message": err.Error(),
			})
			t, _ := jwt.Parse(r.RefreshToken, func(token *jwt.Token) (interface{}, error) {
				secretKey := []byte(config.AppConfig.HashingSecret)
				return secretKey, nil
			})
			claims := t.Claims.(jwt.MapClaims)
			userUuid = claims["sub"].(string)
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
	} else {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{
			Message: "invalid grant_type",
			Code:    ErrorInvalidRequest,
		})
		return
	}

	accessToken, err := generateToken(userUuid, r.ClientId, []string{}, time.Now().Add(time.Minute*15).Unix())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    ErrorUnknown,
			"message": err.Error(),
		})
		return
	}
	refreshToken, err := generateToken(userUuid, r.ClientId, []string{}, time.Now().Add(time.Hour*7*24).Unix())
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
	}{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	c.JSON(http.StatusOK, response)
}

func generateToken(userUuid, clientId string, scopes []string, expiry int64) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = userUuid
	claims["exp"] = expiry
	claims["scope"] = scopes
	claims["aud"] = clientId
	secretKey := []byte(config.AppConfig.HashingSecret)
	t, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return t, nil
}

func Oath2Test(c *gin.Context) {
	/*	accessToken := c.Request.Header.Get("Authorization")
		if ok, err := validateToken(accessToken); !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    ErrorUnknown,
				"message": err.Error(),
			})
			return
		}*/
	c.JSON(http.StatusOK, nil)
}
