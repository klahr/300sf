// SPDX-License-Identifier: GPL-2.0-or-later
package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/klahr/300sf/api/config"
)

func setup() {
	gin.SetMode(gin.TestMode)
	config.AppConfig = &config.App{HashingSecret: "test-secret", AccessTokenLifetime: 60, RefreshTokenLifetime: 120}
}

func TestTokenRoundTrip(t *testing.T) {
	setup()
	tok, err := generateToken("user-1", "client-1", "user admin", time.Now().Add(time.Minute).Unix())
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ValidateToken(tok)
	if err != nil || parsed == nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	claims := parsed.Claims.(jwt.MapClaims)
	if claims["sub"] != "user-1" || claims["aud"] != "client-1" || claims["scp"] != "user admin" {
		t.Errorf("claims = %v", claims)
	}
}

func TestValidateTokenRejects(t *testing.T) {
	setup()
	expired, _ := generateToken("u", "c", "user", time.Now().Add(-time.Minute).Unix())
	if tok, _ := ValidateToken(expired); tok != nil {
		t.Error("expired token accepted")
	}

	good, _ := generateToken("u", "c", "user", time.Now().Add(time.Minute).Unix())
	config.AppConfig.HashingSecret = "other-secret"
	if tok, _ := ValidateToken(good); tok != nil {
		t.Error("token signed with another secret accepted")
	}

	if tok, _ := ValidateToken("not.a.token"); tok != nil {
		t.Error("garbage accepted")
	}

	none := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"sub": "u", "scp": "admin"})
	s, _ := none.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if tok, _ := ValidateToken(s); tok != nil {
		t.Error("alg=none token accepted")
	}
}

func postAuth(body string) *httptest.ResponseRecorder {
	r := gin.New()
	r.POST("/v1/auth", Auth)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func errorCode(t *testing.T, w *httptest.ResponseRecorder) int64 {
	var res GeneralErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode %q: %v", w.Body.String(), err)
	}
	return res.Code
}

func TestAuthRejectsBadRequests(t *testing.T) {
	setup()
	cases := []struct {
		name   string
		body   string
		status int
		code   int64
	}{
		{"invalid json", "{", http.StatusBadRequest, ErrorParse},
		{"unknown grant", `{"grant_type":"magic"}`, http.StatusBadRequest, ErrorInvalidRequest},
		{"bad refresh token", `{"grant_type":"refresh_token","refresh_token":"x"}`, http.StatusUnauthorized, ErrorAuthInvalid},
		{"password without client", `{"grant_type":"password","email":"a@b.se","password":"pw"}`, http.StatusUnauthorized, ErrorMissingApiKey},
		{"malformed client secret", `{"grant_type":"password","client_id":"c","client_secret":"nodot"}`, http.StatusUnauthorized, ErrorInvalidApiKey},
	}
	for _, c := range cases {
		w := postAuth(c.body)
		if w.Code != c.status {
			t.Errorf("%s: status = %d, want %d", c.name, w.Code, c.status)
		}
		if code := errorCode(t, w); code != c.code {
			t.Errorf("%s: code = %d, want %d", c.name, code, c.code)
		}
	}
}

func TestIntersectScopes(t *testing.T) {
	cases := []struct {
		requested []string
		user, key string
		want      string
	}{
		{[]string{"user", "admin"}, "user admin", "user admin", "user admin"},
		{[]string{"user", "admin"}, "user", "user admin", "user"},
		{[]string{"user", "admin"}, "user admin", "user", "user"},
		{[]string{"admin"}, "user", "user admin", ""},
		{[]string{"user", ""}, "user", "user", "user"},
		{nil, "user", "user", ""},
		{[]string{"adm"}, "admin", "admin", ""},
	}
	for _, c := range cases {
		if got := intersectScopes(c.requested, c.user, c.key); got != c.want {
			t.Errorf("intersectScopes(%v, %q, %q) = %q, want %q", c.requested, c.user, c.key, got, c.want)
		}
	}
}

func TestListsCompareRejectsBadInput(t *testing.T) {
	setup()
	r := gin.New()
	r.POST("/v1/lists/compare", ListsCompare)
	tooMany := `{"uuids":["a"` + strings.Repeat(`,"a"`, 20) + `]}`
	for _, body := range []string{"{", `{}`, `{"uuids":[]}`, tooMany} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/v1/lists/compare", strings.NewReader(body)))
		if w.Code != http.StatusBadRequest {
			t.Errorf("body %.30s: status = %d, want 400", body, w.Code)
		}
	}
}

func TestNotImplemented(t *testing.T) {
	setup()
	r := gin.New()
	r.GET("/x", NotImplemented)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
	if w.Code != http.StatusNotImplemented {
		t.Errorf("status = %d", w.Code)
	}
}
