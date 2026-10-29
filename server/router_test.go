// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/config"
	"github.com/klahr/300sf/server/site"
)

func fullRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	config.AppConfig = &config.App{HashingSecret: "test-secret"}
	web, err := site.New(config.Site{Host: "300sf.se", RedirectHosts: map[string]string{"app.300sf.se": "/home"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Panics here if API and website routes conflict.
	return newRouter(web)
}

func TestRouterServesAPIAndSites(t *testing.T) {
	r := fullRouter(t)
	cases := []struct {
		method, host, path string
		status             int
		contentType        string
	}{
		{"GET", "300sf.se", "/health", 200, "application/json"},
		{"GET", "300sf.se", "/api/health", 200, "application/json"},
		{"GET", "300sf.se", "/api/v1/lists/list", 401, "application/json"},
		{"OPTIONS", "300sf.se", "/v1/lists/list", 204, ""},
		{"GET", "300sf.se", "/", 200, "text/html"},
		{"GET", "300sf.se", "/home", 200, "text/html"},
		{"GET", "300sf.se", "/css/base.css", 200, "text/css"},
		// Old hosts: API clients keep working, pages move.
		{"GET", "api.300sf.se", "/health", 200, "application/json"},
		{"POST", "app.300sf.se", "/api/v1/auth", 403, "application/json"},
		{"GET", "app.300sf.se", "/home", 301, ""},
	}
	for _, c := range cases {
		req := httptest.NewRequest(c.method, c.path, nil)
		req.Host = c.host
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != c.status {
			t.Errorf("%s %s%s = %d, want %d", c.method, c.host, c.path, w.Code, c.status)
		}
		if c.contentType != "" && !strings.HasPrefix(w.Header().Get("Content-Type"), c.contentType) {
			t.Errorf("%s %s%s content type = %q, want %s", c.method, c.host, c.path, w.Header().Get("Content-Type"), c.contentType)
		}
	}
}

func TestOnlyAPIGetsCORS(t *testing.T) {
	r := fullRouter(t)
	// The app's cookie-authenticated API is same-origin only.
	for path, want := range map[string]string{"/health": "*", "/v1/lists/list": "*", "/api/health": "", "/": ""} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != want {
			t.Errorf("%s: CORS = %q, want %q", path, got, want)
		}
	}
}

func TestSecurityHeaders(t *testing.T) {
	r := fullRouter(t)
	for _, path := range []string{"/", "/home", "/health", "/css/base.css"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Header().Get("X-Frame-Options") != "DENY" || !strings.Contains(w.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'") {
			t.Errorf("%s can be framed", path)
		}
		if w.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Errorf("%s: missing nosniff", path)
		}
	}
}

func TestRequestBodyIsCapped(t *testing.T) {
	r := fullRouter(t)
	body := `{"email":"` + strings.Repeat("a", maxBodyBytes) + `"}`
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/v1/auth", strings.NewReader(body)))
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "request body too large") {
		t.Errorf("oversized body: %d %s, want it rejected as too large", w.Code, w.Body.String())
	}
}
