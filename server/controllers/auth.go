// SPDX-License-Identifier: GPL-2.0-or-later
package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/klahr/300sf/server/config"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/utils"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	// Failed sign-ins allowed per email address within loginFailWindow.
	maxLoginFailures = 10
	loginFailWindow  = 15 * time.Minute
)

// ValidateToken checks the signature (HS256 only) and expiry of token.
func ValidateToken(token string) (*jwt.Token, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}
		return []byte(config.AppConfig.HashingSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if !t.Valid {
		return nil, nil
	}
	return t, nil
}

// ValidateAccessToken returns the claims of a valid access token. Refresh
// tokens are rejected, so a leaked refresh token can't be used to call the API.
func ValidateAccessToken(token string) (jwt.MapClaims, bool) {
	t, _ := ValidateToken(token)
	if t == nil {
		return nil, false
	}
	claims := t.Claims.(jwt.MapClaims)
	if typ, _ := claims["typ"].(string); typ != TokenTypeAccess {
		return nil, false
	}
	if _, ok := claims["sub"].(string); !ok {
		return nil, false
	}
	if _, ok := claims["scp"].(string); !ok {
		return nil, false
	}
	return claims, true
}

type AuthRequest struct {
	GrantType    string   `json:"grant_type"`
	ClientId     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Email        string   `json:"email"`
	Password     string   `json:"password"`
	RefreshToken string   `json:"refresh_token"`
	Scopes       []string `json:"scopes"`
	StaySignedIn bool     `json:"stay_signed_in"`
}

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
	Lang         string `json:"lang"`
	// Whether the session should outlive the browser session (the app's
	// "stay signed in"). Carried in the refresh token across refreshes.
	StaySignedIn bool `json:"-"`
}

// AuthError is a failed sign-in, ready to send as JSON.
type AuthError struct {
	Status int
	Body   GeneralErrorResponse
}

func authError(status int, code int64, message string) *AuthError {
	return &AuthError{Status: status, Body: GeneralErrorResponse{Code: code, Message: message}}
}

func refreshKey(jti string) string {
	return "/refresh-token/" + jti
}

func loginFailKey(email string) string {
	return "/login-fail/" + strings.ToLower(strings.TrimSpace(email))
}

// Authenticate signs in with a password or a refresh token and issues a new
// access and refresh token. Refresh tokens are single use: each refresh
// replaces the one it was given, and a revoked or reused one is rejected.
func Authenticate(r AuthRequest) (*AuthTokens, *AuthError) {
	var user models.User
	var scopes string
	var clientId string
	var stay bool

	switch r.GrantType {
	case "refresh_token":
		token, _ := ValidateToken(r.RefreshToken)
		if token == nil {
			return nil, authError(http.StatusUnauthorized, ErrorAuthInvalid, "invalid refresh token")
		}
		claims := token.Claims.(jwt.MapClaims)
		typ, _ := claims["typ"].(string)
		jti, _ := claims["jti"].(string)
		userUuid, _ := claims["sub"].(string)
		oldScopes, _ := claims["scp"].(string)
		clientId, _ = claims["aud"].(string)
		stay, _ = claims["stay"].(bool)
		if typ != TokenTypeRefresh || jti == "" {
			return nil, authError(http.StatusUnauthorized, ErrorAuthInvalid, "invalid refresh token")
		}

		// The refresh token is bound to the API key it was issued for.
		if clientId == "" || (r.ClientId != "" && r.ClientId != clientId) {
			return nil, authError(http.StatusUnauthorized, ErrorAuthInvalid, "invalid client_id")
		}
		var apiKey models.ApiKey
		if err := models.DB.Where("client_id=?", clientId).First(&apiKey).Error; err != nil || apiKey.RevokedAt.Valid {
			return nil, authError(http.StatusUnauthorized, ErrorInvalidApiKey, "invalid api key")
		}

		// Consume the token. DEL is atomic, so of two requests racing with the
		// same token only one gets a new session.
		if n, err := models.Cache.Del(refreshKey(jti)).Result(); err != nil || n != 1 {
			return nil, authError(http.StatusUnauthorized, ErrorAuthInvalid, "refresh token revoked or already used")
		}

		if err := models.DB.Where("uuid=?", userUuid).Preload("Language").First(&user).Error; err != nil {
			return nil, authError(http.StatusUnauthorized, ErrorAuthInvalid, "invalid refresh token")
		}
		if user.DeactivatedAt.Valid {
			return nil, authError(http.StatusBadRequest, ErrorAccountDeactivated, "account is deactivated")
		}

		// Scopes can only shrink: re-check them against the user and the key.
		scopes = intersectScopes(strings.Split(oldScopes, " "), user.Scopes, apiKey.Scopes)

	case "password":
		if r.ClientId == "" {
			return nil, authError(http.StatusUnauthorized, ErrorMissingApiKey, "missing client_id")
		}
		v := strings.Split(r.ClientSecret, ".")
		if len(v) != 2 {
			return nil, authError(http.StatusUnauthorized, ErrorInvalidApiKey, "invalid client_secret")
		}

		var apiKey models.ApiKey
		hashedOrignalKey := utils.ComputeHmac256(v[1], config.AppConfig.HashingSecret)
		if err := models.DB.Where("hash=?", hashedOrignalKey).Where("prefix=?", v[0]).Where("client_id=?", r.ClientId).First(&apiKey).Error; err != nil {
			return nil, authError(http.StatusUnauthorized, ErrorInvalidApiKey, "invalid api key")
		}
		if apiKey.RevokedAt.Valid {
			return nil, authError(http.StatusUnauthorized, ErrorInvalidApiKey, "api key has been revoked")
		}

		// Slow down password guessing per account.
		failKey := loginFailKey(r.Email)
		if failures, _ := models.Cache.Get(failKey).Int64(); failures >= maxLoginFailures {
			return nil, authError(http.StatusTooManyRequests, ErrorTooManyLoginAttempts, "too many failed sign-ins, try again later")
		}
		fail := func() *AuthError {
			if n, err := models.Cache.Incr(failKey).Result(); err == nil && n == 1 {
				models.Cache.Expire(failKey, loginFailWindow)
			}
			return authError(http.StatusBadRequest, ErrorAuthInvalid, "invalid email or password")
		}

		if err := models.DB.Where("email=?", r.Email).Preload("Language").First(&user).Error; err != nil {
			return nil, fail()
		}
		if user.DeactivatedAt.Valid {
			return nil, authError(http.StatusBadRequest, ErrorAccountDeactivated, "account is deactivated")
		}
		p := utils.Password{}
		if !p.IsMatching(r.Password, user.Password, user.PasswordSalt) {
			return nil, fail()
		}
		models.Cache.Del(failKey)

		clientId = r.ClientId
		stay = r.StaySignedIn
		scopes = intersectScopes(r.Scopes, user.Scopes, apiKey.Scopes)

	default:
		return nil, authError(http.StatusBadRequest, ErrorInvalidRequest, "invalid grant_type")
	}

	accessToken, err := generateToken(TokenTypeAccess, user.Uuid, clientId, scopes, "", stay, config.AppConfig.AccessTokenLifetime)
	if err != nil {
		return nil, authError(http.StatusInternalServerError, ErrorUnknown, err.Error())
	}
	jti := uuid.New().String()
	refreshToken, err := generateToken(TokenTypeRefresh, user.Uuid, clientId, scopes, jti, stay, config.AppConfig.RefreshTokenLifetime)
	if err != nil {
		return nil, authError(http.StatusInternalServerError, ErrorUnknown, err.Error())
	}
	if err := models.Cache.Set(refreshKey(jti), user.Uuid, time.Duration(config.AppConfig.RefreshTokenLifetime)*time.Second).Err(); err != nil {
		return nil, authError(http.StatusServiceUnavailable, ErrorUnknown, "could not store session")
	}

	models.DB.Model(&user).UpdateColumn("last_seen_at", time.Now())

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    config.AppConfig.AccessTokenLifetime,
		Scope:        scopes,
		Lang:         user.Language.Code,
		StaySignedIn: stay,
	}, nil
}

// RevokeRefreshToken ends the session the refresh token belongs to.
func RevokeRefreshToken(refreshToken string) {
	token, _ := ValidateToken(refreshToken)
	if token == nil {
		return
	}
	if jti, _ := token.Claims.(jwt.MapClaims)["jti"].(string); jti != "" {
		models.Cache.Del(refreshKey(jti))
	}
}

func Auth(c *gin.Context) {
	var r AuthRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{
			Message: err.Error(),
			Code:    ErrorParse,
		})
		return
	}

	tokens, authErr := Authenticate(r)
	if authErr != nil {
		c.JSON(authErr.Status, authErr.Body)
		return
	}
	c.JSON(http.StatusOK, tokens)
}

// AuthLogout revokes a refresh token so it can't be used again.
func AuthLogout(c *gin.Context) {
	var r struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Message: err.Error(), Code: ErrorParse})
		return
	}
	RevokeRefreshToken(r.RefreshToken)
	c.Status(http.StatusNoContent)
}

// intersectScopes keeps the requested scopes that both the user and the API key have.
func intersectScopes(requested []string, userScopes, apiKeyScopes string) string {
	has := func(list string, scope string) bool {
		for _, s := range strings.Split(list, " ") {
			if s == scope {
				return true
			}
		}
		return false
	}
	var result []string
	for _, s := range requested {
		if s != "" && has(userScopes, s) && has(apiKeyScopes, s) {
			result = append(result, s)
		}
	}
	return strings.Join(result, " ")
}

func generateToken(typ, userUuid, clientId, scopes, jti string, stay bool, lifetimeSeconds int64) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"typ": typ,
		"sub": userUuid,
		"aud": clientId,
		"scp": scopes,
		"iat": now.Unix(),
		"exp": now.Add(time.Duration(lifetimeSeconds) * time.Second).Unix(),
	}
	if jti != "" {
		claims["jti"] = jti
	}
	if stay {
		claims["stay"] = true
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.AppConfig.HashingSecret))
}
