// SPDX-License-Identifier: AGPL-3.0-or-later
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

	// Wrong passwords allowed within loginFailWindow: for one account from
	// one network, from one network for any accounts (password spraying),
	// and for one account from anywhere (guessing spread over many networks).
	// The last is high so that nobody can lock someone out with a few tries.
	maxLoginFailures       = 10
	maxLoginFailuresPerNet = 30
	loginFailWindow        = 15 * time.Minute
	maxLoginFailuresAnyNet = 100
	loginFailWindowAnyNet  = time.Hour
)

// ValidateToken checks the signature (HS256 only) and expiry of token.
func ValidateToken(token string) (*jwt.Token, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}
		return []byte(config.AppConfig.TokenSecret), nil
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
	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, false
	}
	if _, ok := claims["scp"].(string); !ok {
		return nil, false
	}
	// Tokens issued before the user was signed out everywhere. If Redis is
	// unavailable the token is accepted, like the rest of the API does.
	if models.Cache != nil {
		if revokedBefore, err := models.Cache.Get(revokedBeforeKey(sub)).Int64(); err == nil {
			if iat, _ := claims["iat"].(float64); int64(iat) < revokedBefore {
				return nil, false
			}
		}
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
	// The client's address, for limiting password guessing; set by the server.
	ClientIP string `json:"-"`
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

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// loginFailKeys are the counters of wrong passwords for this account from
// this network, from this network, and for this account.
func loginFailKeys(email, ip string) (accountNet, net, account string) {
	n := ClientNetwork(ip)
	return "/login-fail/" + n + "/" + normalizeEmail(email), "/login-fail-net/" + n, "/login-fail-account/" + normalizeEmail(email)
}

// userSessionsKey holds the jti of every refresh token a user has out.
func userSessionsKey(userUuid string) string {
	return "/user-sessions/" + userUuid
}

// revokedBeforeKey holds the time before which the user's access tokens are
// no longer accepted.
func revokedBeforeKey(userUuid string) string {
	return "/revoked-before/" + userUuid
}

// PasswordLocked reports whether too many wrong passwords have been tried
// for the account with this email, or from the client's network.
func PasswordLocked(email, ip string) bool {
	accountNet, net, account := loginFailKeys(email, ip)
	count := func(key string) int64 {
		n, _ := models.Cache.Get(key).Int64()
		return n
	}
	return count(accountNet) >= maxLoginFailures || count(net) >= maxLoginFailuresPerNet || count(account) >= maxLoginFailuresAnyNet
}

// PasswordFailed counts a wrong password for the account with this email,
// tried from the client at ip.
func PasswordFailed(email, ip string) {
	accountNet, net, account := loginFailKeys(email, ip)
	incr := func(key string, window time.Duration) {
		// Also repair a counter left without expiry, or the key would be locked forever.
		if n, err := models.Cache.Incr(key).Result(); err == nil && (n == 1 || models.Cache.TTL(key).Val() < 0) {
			models.Cache.Expire(key, window)
		}
	}
	incr(accountNet, loginFailWindow)
	incr(net, loginFailWindow)
	incr(account, loginFailWindowAnyNet)
}

// passwordSucceeded forgets the wrong passwords tried for this account from
// this network. The other counters are kept: signing in to an account of
// one's own must not reset them.
func passwordSucceeded(email, ip string) {
	accountNet, _, _ := loginFailKeys(email, ip)
	models.Cache.Del(accountNet)
}

// RevokeUserSessions signs the user out everywhere: every refresh token is
// revoked and access tokens issued until now are rejected.
func RevokeUserSessions(userUuid string) {
	if jtis, err := models.Cache.SMembers(userSessionsKey(userUuid)).Result(); err == nil {
		for _, jti := range jtis {
			models.Cache.Del(refreshKey(jti))
		}
	}
	models.Cache.Del(userSessionsKey(userUuid))
	models.Cache.Set(revokedBeforeKey(userUuid), time.Now().Unix(), time.Duration(config.AppConfig.AccessTokenLifetime)*time.Second)
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
		models.Cache.SRem(userSessionsKey(userUuid), jti)

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
		if PasswordLocked(r.Email, r.ClientIP) {
			return nil, authError(http.StatusTooManyRequests, ErrorTooManyLoginAttempts, "too many failed sign-ins, try again later")
		}
		fail := func() *AuthError {
			PasswordFailed(r.Email, r.ClientIP)
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
		passwordSucceeded(r.Email, r.ClientIP)

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
	refreshLifetime := time.Duration(config.AppConfig.RefreshTokenLifetime) * time.Second
	if err := models.Cache.Set(refreshKey(jti), user.Uuid, refreshLifetime).Err(); err != nil {
		return nil, authError(http.StatusServiceUnavailable, ErrorUnknown, "could not store session")
	}
	models.Cache.SAdd(userSessionsKey(user.Uuid), jti)
	models.Cache.Expire(userSessionsKey(user.Uuid), refreshLifetime)

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
	claims := token.Claims.(jwt.MapClaims)
	if jti, _ := claims["jti"].(string); jti != "" {
		models.Cache.Del(refreshKey(jti))
		if sub, _ := claims["sub"].(string); sub != "" {
			models.Cache.SRem(userSessionsKey(sub), jti)
		}
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

	r.ClientIP = c.ClientIP()
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
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.AppConfig.TokenSecret))
}
