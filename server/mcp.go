// SPDX-License-Identifier: AGPL-3.0-or-later
package main

// The MCP endpoint, /mcp, lets AI assistants such as Claude use an account
// through tools (mcp_tools.go). Apps connect with OAuth (controllers/oauth.go)
// or, like any API client, with a token from an API key.
//
// It is stateless: every request carries its access token. The tools call
// the API in-process with that token, so they get exactly the checks, scopes
// and call limits any other client gets, and the tools a token is offered
// follow its scopes: a read-only one gets no tools that change anything.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/auth"
	"github.com/klahr/300sf/server/controllers"
	"github.com/klahr/300sf/server/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// apiCaller calls the API as the client that called /mcp.
type apiCaller struct {
	api     http.Handler
	request *http.Request
	scopes  []string
}

func (a *apiCaller) canChange() bool {
	return slices.Contains(a.scopes, auth.AuthScopeUser)
}

// call makes an API request and decodes its JSON answer. An answer that
// isn't a success is an error with the API's message, for the assistant to
// read.
func (a *apiCaller) call(ctx context.Context, method, path string, body any, language string) (any, error) {
	var reader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequestWithContext(ctx, method, path, reader)
	// As the original request, so per-address limits and logs see the client.
	req.RemoteAddr = a.request.RemoteAddr
	req.Host = a.request.Host
	if v := a.request.Header.Get("X-Forwarded-For"); v != "" {
		req.Header.Set("X-Forwarded-For", v)
	}
	req.Header.Set("Authorization", a.request.Header.Get("Authorization"))
	req.Header.Set("Content-Type", "application/json")
	if language != "" {
		req.Header.Set("X-App-Language", language)
	}
	w := httptest.NewRecorder()
	a.api.ServeHTTP(w, req)

	var decoded any
	if w.Body.Len() > 0 {
		if err := json.Unmarshal(w.Body.Bytes(), &decoded); err != nil {
			return nil, fmt.Errorf("%s %s: unreadable answer (HTTP %d)", method, path, w.Code)
		}
	}
	if w.Code < 200 || w.Code > 299 {
		message := http.StatusText(w.Code)
		if m, ok := decoded.(map[string]any); ok {
			if s, ok := m["message"].(string); ok && s != "" {
				message = s
			}
		}
		return nil, fmt.Errorf("%s (HTTP %d)", message, w.Code)
	}
	return decoded, nil
}

type callerKey struct{}

// newMcpHandler serves /mcp, with tools calling api.
func newMcpHandler(api http.Handler) gin.HandlerFunc {
	h := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		caller, _ := r.Context().Value(callerKey{}).(*apiCaller)
		if caller == nil {
			return nil
		}
		return newMcpServer(caller)
	}, &mcp.StreamableHTTPOptions{
		Stateless:    true,
		JSONResponse: true,
		// nginx passes requests on from 127.0.0.1 with the public Host; the
		// protection against DNS rebinding would refuse them all. It guards
		// servers without authentication on a developer's machine; every
		// request here needs a token.
		DisableLocalhostProtection: true,
	})

	return func(c *gin.Context) {
		// Apps start OAuth from this challenge: where to read how to get a token.
		challenge := `Bearer resource_metadata="` + controllers.McpResourceMetadata() + `"`
		unauthorized := func(description string) {
			if c.GetHeader("Authorization") != "" {
				challenge += `, error="invalid_token", error_description="` + description + `"`
			}
			c.Header("WWW-Authenticate", challenge)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": description, "code": controllers.ErrorAuthInvalid})
		}

		claims := parseToken(c.GetHeader("Authorization"))
		if claims == nil {
			unauthorized("missing or invalid access token")
			return
		}
		var apiKey models.ApiKey
		if err := models.DB.Where("client_id=?", claims["aud"]).First(&apiKey).Error; err != nil || apiKey.RevokedAt.Valid {
			unauthorized("the app was disconnected")
			return
		}
		scp, _ := claims["scp"].(string)
		scopes := strings.Fields(scp)
		if !slices.Contains(scopes, auth.AuthScopeUser) && !slices.Contains(scopes, auth.AuthScopeRead) {
			c.Header("WWW-Authenticate", challenge+`, error="insufficient_scope", scope="read"`)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "missing scope", "code": controllers.ErrorMissingScope})
			return
		}

		caller := &apiCaller{api: api, request: c.Request, scopes: scopes}
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), callerKey{}, caller))
		h.ServeHTTP(c.Writer, c.Request)
	}
}
