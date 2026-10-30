// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDatabaseErrorsBecomeBadRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(dbErrors)
	r.GET("/v1/long", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error 1406 (22001): Data too long for column 'site' at row 1"})
	})
	r.GET("/v1/other", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "boom"})
	})
	r.GET("/v1/empty", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusInternalServerError)
	})
	r.GET("/v1/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Error 1406 (22001): Data too long for column 'x'"})
	})
	for path, want := range map[string]struct {
		status int
		body   string
	}{
		"/v1/long":  {http.StatusBadRequest, `{"message":"site is too long","code":2001}`},
		"/v1/other": {http.StatusInternalServerError, `{"message":"boom"}`},
		"/v1/empty": {http.StatusInternalServerError, ``},
		"/v1/ok":    {http.StatusOK, `{"message":"Error 1406 (22001): Data too long for column 'x'"}`},
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != want.status || w.Body.String() != want.body {
			t.Errorf("%s = %d %s, want %d %s", path, w.Code, w.Body.String(), want.status, want.body)
		}
	}
}
