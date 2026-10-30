// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

// OAuth 2.1 for apps that connect to an account, such as Claude's
// connectors. This server is both the authorization server and the resource
// (the MCP endpoint, /mcp, and the API behind it):
//
//  1. The app finds the endpoints in /.well-known/oauth-protected-resource
//     and /.well-known/oauth-authorization-server.
//  2. It registers itself (RFC 7591) and gets a client_id. Apps are public
//     clients: they have no secret and prove themselves with PKCE instead.
//  3. It sends the user to /oauth/authorize, a page where they sign in and
//     approve it (OAuthApprove). Approving makes a connection: an API key of
//     the user's own for that app, which Settings lists and can revoke.
//  4. The app swaps the code for tokens on /oauth/token, and refreshes them
//     there. They are the API's usual tokens, made through the connection,
//     so everything that checks API keys (revoking, call limits, scopes)
//     applies to the app too.

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/klahr/300sf/server/auth"
	"github.com/klahr/300sf/server/config"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/utils"
	"gopkg.in/guregu/null.v4/zero"
	"gorm.io/gorm"
)

const (
	oauthCodeLifetime   = 2 * time.Minute
	maxRedirectUris     = 10
	maxRedirectUriLen   = 512
	maxOauthClientName  = 128
	defaultOauthAppName = "Unnamed app"
)

// oauthScopes are what an app can be given. Admin access is never handed
// to an app.
var oauthScopes = []string{auth.AuthScopeRead, auth.AuthScopeUser}

var codeVerifierPattern = regexp.MustCompile(`^[A-Za-z0-9._~-]{43,128}$`)

// Issuer is this server's public address, which names it in OAuth.
func Issuer() string {
	return strings.TrimSuffix(config.AppConfig.Domain, "/")
}

// McpResource is the address of the MCP endpoint, the resource apps get tokens for.
func McpResource() string {
	return Issuer() + "/mcp"
}

// McpResourceMetadata is where an app learns how to get a token for /mcp (RFC 9728).
func McpResourceMetadata() string {
	return Issuer() + "/.well-known/oauth-protected-resource/mcp"
}

type oauthError struct {
	Error       string `json:"error"`
	Description string `json:"error_description,omitempty"`
}

func oauthFail(c *gin.Context, status int, code, description string) {
	c.JSON(status, oauthError{Error: code, Description: description})
}

// OAuthServerMetadata describes the authorization server (RFC 8414).
func OAuthServerMetadata(c *gin.Context) {
	issuer := Issuer()
	c.JSON(http.StatusOK, gin.H{
		"issuer":                                         issuer,
		"authorization_endpoint":                         issuer + "/oauth/authorize",
		"token_endpoint":                                 issuer + "/oauth/token",
		"registration_endpoint":                          issuer + "/oauth/register",
		"revocation_endpoint":                            issuer + "/oauth/revoke",
		"response_types_supported":                       []string{"code"},
		"response_modes_supported":                       []string{"query"},
		"grant_types_supported":                          []string{"authorization_code", "refresh_token"},
		"token_endpoint_auth_methods_supported":          []string{"none"},
		"revocation_endpoint_auth_methods_supported":     []string{"none"},
		"code_challenge_methods_supported":               []string{"S256"},
		"scopes_supported":                               oauthScopes,
		"authorization_response_iss_parameter_supported": true,
		"service_documentation":                          issuer + "/docs",
	})
}

// OAuthResourceMetadata describes the MCP endpoint as a protected resource (RFC 9728).
func OAuthResourceMetadata(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"resource":                 McpResource(),
		"authorization_servers":    []string{Issuer()},
		"scopes_supported":         oauthScopes,
		"bearer_methods_supported": []string{"header"},
		"resource_name":            "300 Svenska Fåglar",
		"resource_documentation":   Issuer() + "/docs",
	})
}

func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// validRedirectUri accepts https addresses, and http ones on this computer
// for apps that run on it (RFC 8252).
func validRedirectUri(raw string) bool {
	if len(raw) > maxRedirectUriLen {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.Fragment != "" || u.User != nil {
		return false
	}
	switch u.Scheme {
	case "https":
		return true
	case "http":
		return isLoopback(u.Hostname())
	}
	return false
}

// matchRedirectUri picks the registered address the app asked to be sent
// back to. It must be registered exactly, except that the port of an
// address on this computer may differ, since apps there get one when they
// start (RFC 8252, section 7.3). With one registered it may be left out.
func matchRedirectUri(registered []string, given string) (string, bool) {
	if given == "" {
		if len(registered) == 1 {
			return registered[0], true
		}
		return "", false
	}
	if slices.Contains(registered, given) {
		return given, true
	}
	g, err := url.Parse(given)
	if err != nil || g.Scheme != "http" || !isLoopback(g.Hostname()) || g.Fragment != "" || g.User != nil {
		return "", false
	}
	for _, r := range registered {
		u, err := url.Parse(r)
		if err == nil && u.Scheme == g.Scheme && u.Hostname() == g.Hostname() && u.Path == g.Path && u.RawQuery == g.RawQuery {
			return given, true
		}
	}
	return "", false
}

// redirectWith adds params to the address the app is sent back to, keeping
// its own query.
func redirectWith(redirectUri string, params map[string]string) string {
	u, err := url.Parse(redirectUri)
	if err != nil {
		return redirectUri
	}
	q := u.Query()
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// OAuthRegister lets an app register itself (RFC 7591). Anyone may, so the
// name it gives is shown to users next to the address it sends them back to,
// which it can't fake.
func OAuthRegister(c *gin.Context) {
	type request struct {
		RedirectUris []string `json:"redirect_uris"`
		ClientName   string   `json:"client_name"`
	}
	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		oauthFail(c, http.StatusBadRequest, "invalid_client_metadata", err.Error())
		return
	}
	if Throttled("oauth-register/"+ClientNetwork(c.ClientIP()), 20, time.Hour) {
		oauthFail(c, http.StatusTooManyRequests, "invalid_request", "too many registrations, try again later")
		return
	}
	if len(r.RedirectUris) == 0 || len(r.RedirectUris) > maxRedirectUris {
		oauthFail(c, http.StatusBadRequest, "invalid_redirect_uri", "give 1 to 10 redirect_uris")
		return
	}
	for _, uri := range r.RedirectUris {
		if !validRedirectUri(uri) {
			oauthFail(c, http.StatusBadRequest, "invalid_redirect_uri", "redirect_uris must be https, or http on localhost: "+uri)
			return
		}
	}
	name := strings.TrimSpace(r.ClientName)
	if name == "" {
		name = defaultOauthAppName
	}
	if len([]rune(name)) > maxOauthClientName {
		name = string([]rune(name)[:maxOauthClientName])
	}

	uris, _ := json.Marshal(r.RedirectUris)
	client := models.OauthClient{
		Model:        models.Model{Timestamp: time.Now(), Uuid: utils.GenerateShortUuid()},
		ClientId:     utils.GenerateUuidWithFixedLength(32),
		Name:         name,
		RedirectUris: string(uris),
	}
	if err := models.DB.Create(&client).Error; err != nil {
		oauthFail(c, http.StatusInternalServerError, "server_error", "could not register")
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"client_id":                  client.ClientId,
		"client_id_issued_at":        client.Timestamp.Unix(),
		"client_name":                client.Name,
		"redirect_uris":              r.RedirectUris,
		"grant_types":                []string{"authorization_code", "refresh_token"},
		"response_types":             []string{"code"},
		"token_endpoint_auth_method": "none",
	})
}

// AuthorizeRequest is what an app asks for when it sends the user to
// /oauth/authorize; the page sends it on to OAuthApprove.
type AuthorizeRequest struct {
	ClientId            string `json:"client_id" form:"client_id"`
	RedirectUri         string `json:"redirect_uri" form:"redirect_uri"`
	ResponseType        string `json:"response_type" form:"response_type"`
	State               string `json:"state" form:"state"`
	CodeChallenge       string `json:"code_challenge" form:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method" form:"code_challenge_method"`
	Scope               string `json:"scope" form:"scope"`
	Resource            string `json:"resource" form:"resource"`
}

// AuthorizeCheck is an AuthorizeRequest looked over.
type AuthorizeCheck struct {
	Client      models.OauthClient
	RedirectUri string
	// AppHost is where the user is sent back to, to show next to the app's name.
	AppHost string
	// AsksToChange is whether the app asks to change things too, not only read.
	AsksToChange bool
	// Problem is set when the app can't be told what went wrong, because the
	// client or the address to send the user back to is wrong. Show it.
	Problem string
	// ErrorRedirect is set when the request is wrong in a way the app is told
	// about: send the user there.
	ErrorRedirect string
}

// CheckAuthorizeRequest looks over an app's request to be let in.
func CheckAuthorizeRequest(r AuthorizeRequest) AuthorizeCheck {
	var check AuthorizeCheck
	if r.ClientId == "" || models.DB.Where("client_id=?", r.ClientId).First(&check.Client).Error != nil {
		check.Problem = "unknown client_id"
		return check
	}
	var registered []string
	_ = json.Unmarshal([]byte(check.Client.RedirectUris), &registered)
	uri, ok := matchRedirectUri(registered, r.RedirectUri)
	if !ok {
		check.Problem = "redirect_uri is not registered for this app"
		return check
	}
	check.RedirectUri = uri
	if u, err := url.Parse(uri); err == nil {
		check.AppHost = u.Hostname()
	}

	fail := func(code, description string) AuthorizeCheck {
		check.ErrorRedirect = redirectWith(uri, map[string]string{"error": code, "error_description": description, "state": r.State, "iss": Issuer()})
		return check
	}
	if r.ResponseType != "code" {
		return fail("unsupported_response_type", "response_type must be code")
	}
	if r.CodeChallenge == "" || r.CodeChallengeMethod != "S256" {
		return fail("invalid_request", "PKCE with code_challenge_method S256 is required")
	}
	if r.Resource != "" && r.Resource != Issuer() && r.Resource != Issuer()+"/" && r.Resource != McpResource() {
		return fail("invalid_target", "unknown resource")
	}
	// Scopes other than ours are ignored; asking for none of ours is asking for all.
	requested := strings.Fields(r.Scope)
	check.AsksToChange = slices.Contains(requested, auth.AuthScopeUser) || !slices.ContainsFunc(requested, func(s string) bool { return slices.Contains(oauthScopes, s) })
	return check
}

// oauthCode is what an authorization code stands for, kept in Redis until
// the app swaps it for tokens.
type oauthCode struct {
	ClientId      string `json:"client_id"`
	KeyClientId   string `json:"key_client_id"`
	UserUuid      string `json:"user_uuid"`
	RedirectUri   string `json:"redirect_uri"`
	CodeChallenge string `json:"code_challenge"`
	Scopes        string `json:"scopes"`
}

// Codes are kept by their hash, so a look at Redis can't use them.
func oauthCodeKey(code string) string {
	sum := sha256.Sum256([]byte(code))
	return "/oauth-code/" + hex.EncodeToString(sum[:])
}

var errTooManyApiKeys = errors.New("too many api keys")

// connect gives the app a connection to the user's account with scopes,
// or changes the scopes of the one it has.
func connect(user models.User, client models.OauthClient, scopes string) (models.ApiKey, error) {
	var key models.ApiKey
	err := models.DB.Where("user_id=?", user.Id).Where("oauth_client_id=?", client.Id).Where("revoked_at IS NULL").First(&key).Error
	if err == nil {
		if key.Scopes != scopes || key.Name != client.Name {
			err = models.DB.Model(&key).Updates(map[string]any{"scopes": scopes, "name": client.Name}).Error
		}
		return key, err
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return key, err
	}
	if tooManyApiKeys(user.Id) {
		return key, errTooManyApiKeys
	}
	// The secret is thrown away: the app signs in through OAuth only.
	key, _ = newApiKey(user.Id, client.Name, scopes)
	key.OauthClientId = zero.IntFrom(client.Id)
	return key, models.DB.Create(&key).Error
}

// OAuthApprove is the user's answer on the authorize page, made with their
// website session. It answers with where to send them back to the app: with
// a code when they let it in, or an error when not.
func OAuthApprove(c *gin.Context) {
	type request struct {
		AuthorizeRequest
		Allow bool `json:"allow"`
		// AllowChanges gives the app the user scope instead of only read.
		AllowChanges bool `json:"allow_changes"`
	}
	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorParse, Message: err.Error()})
		return
	}
	check := CheckAuthorizeRequest(r.AuthorizeRequest)
	if check.Problem != "" {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorInvalidRequest, Message: check.Problem})
		return
	}
	if check.ErrorRedirect != "" {
		c.JSON(http.StatusOK, gin.H{"redirect": check.ErrorRedirect})
		return
	}
	if !r.Allow {
		c.JSON(http.StatusOK, gin.H{"redirect": redirectWith(check.RedirectUri, map[string]string{
			"error": "access_denied", "state": r.State, "iss": Issuer(),
		})})
		return
	}

	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorInvalidRequest, Message: err.Error()})
		return
	}
	scopes := auth.AuthScopeRead
	if r.AllowChanges && check.AsksToChange && hasScope(user.Scopes, auth.AuthScopeUser) {
		scopes = auth.AuthScopeUser
	}
	key, err := connect(user, check.Client, scopes)
	if errors.Is(err, errTooManyApiKeys) {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorTooManyApiKeys, Message: "too many api keys and connected apps, revoke one first"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, GeneralErrorResponse{Code: ErrorUnknown, Message: "could not connect the app"})
		return
	}

	code := utils.RandomSha256()
	stored, _ := json.Marshal(oauthCode{
		ClientId:      check.Client.ClientId,
		KeyClientId:   key.ClientId,
		UserUuid:      user.Uuid,
		RedirectUri:   check.RedirectUri,
		CodeChallenge: r.CodeChallenge,
		Scopes:        scopes,
	})
	if err := models.Cache.Set(oauthCodeKey(code), stored, oauthCodeLifetime).Err(); err != nil {
		c.JSON(http.StatusServiceUnavailable, GeneralErrorResponse{Code: ErrorUnknown, Message: "could not store the code"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"redirect": redirectWith(check.RedirectUri, map[string]string{
		"code": code, "state": r.State, "iss": Issuer(),
	})})
}

// pkceMatches checks the app's code_verifier against the code_challenge it
// started with (RFC 7636, S256).
func pkceMatches(verifier, challenge string) bool {
	if !codeVerifierPattern.MatchString(verifier) {
		return false
	}
	sum := sha256.Sum256([]byte(verifier))
	return subtle.ConstantTimeCompare([]byte(base64.RawURLEncoding.EncodeToString(sum[:])), []byte(challenge)) == 1
}

// OAuthToken swaps an authorization code, or a refresh token, for tokens.
func OAuthToken(c *gin.Context) {
	c.Header("Pragma", "no-cache")
	clientId := c.PostForm("client_id")

	var tokens *AuthTokens
	switch c.PostForm("grant_type") {
	case "authorization_code":
		// Get, then delete: of two requests racing with one code, only the one
		// whose delete removed it goes on.
		key := oauthCodeKey(c.PostForm("code"))
		raw, err := models.Cache.Get(key).Result()
		if err != nil {
			oauthFail(c, http.StatusBadRequest, "invalid_grant", "unknown or expired code")
			return
		}
		if n, err := models.Cache.Del(key).Result(); err != nil || n != 1 {
			oauthFail(c, http.StatusBadRequest, "invalid_grant", "code already used")
			return
		}
		var code oauthCode
		if json.Unmarshal([]byte(raw), &code) != nil || code.ClientId != clientId || code.RedirectUri != c.PostForm("redirect_uri") {
			oauthFail(c, http.StatusBadRequest, "invalid_grant", "the code was given to another client or redirect_uri")
			return
		}
		if !pkceMatches(c.PostForm("code_verifier"), code.CodeChallenge) {
			oauthFail(c, http.StatusBadRequest, "invalid_grant", "code_verifier does not match")
			return
		}

		var user models.User
		if err := models.DB.Where("uuid=?", code.UserUuid).Preload("Language").First(&user).Error; err != nil || user.DeactivatedAt.Valid {
			oauthFail(c, http.StatusBadRequest, "invalid_grant", "the account can't be used")
			return
		}
		var apiKey models.ApiKey
		if err := models.DB.Where("client_id=?", code.KeyClientId).First(&apiKey).Error; err != nil || apiKey.RevokedAt.Valid {
			oauthFail(c, http.StatusBadRequest, "invalid_grant", "the connection was removed")
			return
		}
		scopes := intersectScopes(strings.Fields(code.Scopes), user.Scopes, apiKey.Scopes)
		if scopes == "" {
			oauthFail(c, http.StatusBadRequest, "invalid_grant", "the scopes are no longer allowed")
			return
		}
		var authErr *AuthError
		if tokens, authErr = issueTokens(user, apiKey.ClientId, scopes, true); authErr != nil {
			oauthFail(c, http.StatusServiceUnavailable, "server_error", "could not start a session")
			return
		}

	case "refresh_token":
		refreshToken := c.PostForm("refresh_token")
		// Only a connection's refresh tokens are taken here, and only from its app.
		token, _ := ValidateToken(refreshToken)
		if token == nil {
			oauthFail(c, http.StatusBadRequest, "invalid_grant", "invalid refresh token")
			return
		}
		aud, _ := token.Claims.(jwt.MapClaims)["aud"].(string)
		var apiKey models.ApiKey
		if err := models.DB.Where("client_id=?", aud).First(&apiKey).Error; err != nil || !apiKey.OauthClientId.Valid {
			oauthFail(c, http.StatusBadRequest, "invalid_grant", "invalid refresh token")
			return
		}
		if clientId != "" {
			var client models.OauthClient
			if err := models.DB.Where("client_id=?", clientId).First(&client).Error; err != nil || client.Id != apiKey.OauthClientId.Int64 {
				oauthFail(c, http.StatusBadRequest, "invalid_grant", "the refresh token was given to another client")
				return
			}
		}
		var authErr *AuthError
		if tokens, authErr = Authenticate(AuthRequest{GrantType: "refresh_token", RefreshToken: refreshToken, ClientIP: c.ClientIP()}); authErr != nil {
			oauthFail(c, http.StatusBadRequest, "invalid_grant", authErr.Body.Message)
			return
		}

	default:
		oauthFail(c, http.StatusBadRequest, "unsupported_grant_type", "grant_type must be authorization_code or refresh_token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"token_type":    "Bearer",
		"expires_in":    tokens.ExpiresIn,
		"refresh_token": tokens.RefreshToken,
		"scope":         tokens.Scope,
	})
}

// OAuthRevoke ends the session a refresh token belongs to (RFC 7009). It
// always answers 200, as the RFC asks, so it tells nothing about tokens.
func OAuthRevoke(c *gin.Context) {
	if token := c.PostForm("token"); token != "" {
		RevokeRefreshToken(token)
	}
	c.Status(http.StatusOK)
}
