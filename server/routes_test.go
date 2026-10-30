// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/klahr/300sf/server/config"
	"github.com/klahr/300sf/server/controllers"
)

func token(t *testing.T, scopes string) string {
	tok := jwt.New(jwt.SigningMethodHS256)
	claims := tok.Claims.(jwt.MapClaims)
	claims["sub"] = "user-1"
	claims["aud"] = "client-1"
	claims["scp"] = scopes
	claims["typ"] = controllers.TokenTypeAccess
	claims["exp"] = time.Now().Add(time.Minute).Unix()
	s, err := tok.SignedString([]byte(config.AppConfig.TokenSecret))
	if err != nil {
		t.Fatal(err)
	}
	return "Bearer " + s
}

func newScopeRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	config.AppConfig = &config.App{HashingSecret: "test-secret", TokenSecret: "test-token-secret"}
	r := gin.New()
	ok := func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"uuid": c.GetString("uuid")}) }
	r.GET("/none", noneMiddleware, ok)
	r.GET("/user", basicMiddleware, ok)
	r.GET("/admin", adminMiddleware, ok)
	r.GET("/health", Health)
	return r
}

func do(r *gin.Engine, path, auth string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	r.ServeHTTP(w, req)
	return w
}

func TestScopeMiddleware(t *testing.T) {
	r := newScopeRouter()
	cases := []struct {
		path   string
		auth   string
		status int
	}{
		{"/none", "", http.StatusOK},
		{"/none", token(t, "user"), http.StatusOK},
		{"/user", "", http.StatusUnauthorized},
		{"/user", "Bearer garbage", http.StatusUnauthorized},
		{"/user", "Basic " + token(t, "user")[7:], http.StatusUnauthorized},
		{"/user", token(t, "user"), http.StatusOK},
		{"/user", token(t, "admin"), http.StatusUnauthorized},
		{"/user", token(t, "user admin"), http.StatusOK},
		{"/admin", token(t, "user"), http.StatusUnauthorized},
		{"/admin", token(t, "user admin"), http.StatusOK},
		{"/admin", token(t, "administrator"), http.StatusUnauthorized},
	}
	for _, c := range cases {
		if w := do(r, c.path, c.auth); w.Code != c.status {
			t.Errorf("GET %s with %q: status = %d, want %d (%s)", c.path, c.auth, w.Code, c.status, w.Body.String())
		}
	}
}

func TestScopeMiddlewareSetsUser(t *testing.T) {
	r := newScopeRouter()
	w := do(r, "/user", token(t, "user"))
	if body := w.Body.String(); body != `{"uuid":"user-1"}` {
		t.Errorf("body = %s", body)
	}
}

func TestHealth(t *testing.T) {
	w := do(newScopeRouter(), "/health", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if body := w.Body.String(); body != `{"message":"healthy","data":{"version":"dev"}}` {
		t.Errorf("body = %s", body)
	}
}

func TestServerConfig(t *testing.T) {
	config.AppConfig = &config.App{
		ContactEmail: "owner@example.com",
		Site:         config.Site{MapTiles: "https://tiles.example/{z}/{x}/{y}.png", MapAttribution: "Tiles", SourceUrl: "https://example.com/src"},
	}
	r := gin.New()
	r.GET("/v1/config", noneMiddleware, ServerConfig)
	w := do(r, "/v1/config", "")
	want := `{"version":"dev","source_url":"https://example.com/src","map":{"tiles":"https://tiles.example/{z}/{x}/{y}.png","attribution":"Tiles"},` +
		`"features":{"analyze":false,"contact":true,"region_search":false}}`
	if w.Code != http.StatusOK || w.Body.String() != want {
		t.Errorf("config = %d %s", w.Code, w.Body.String())
	}
}
