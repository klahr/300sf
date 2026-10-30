// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/controllers"
	"github.com/klahr/300sf/server/site"
)

// Database errors that are the client's fault. Handlers pass the driver's
// message on; these become a 400 the client can act on.
var (
	tooLongError   = regexp.MustCompile(`Error 1406 \(22001\): Data too long for column '([^']+)'`)
	outOfRange     = regexp.MustCompile(`Error 1264 \(22003\): Out of range value for column '([^']+)'`)
	duplicateError = regexp.MustCompile(`Error 1062 \(23000\): Duplicate entry`)
)

// clientDBError returns the response for a database error in body, if it is one.
func clientDBError(body []byte) (controllers.GeneralErrorResponse, bool) {
	if m := tooLongError.FindSubmatch(body); m != nil {
		return controllers.GeneralErrorResponse{Code: controllers.ErrorParse, Message: string(m[1]) + " is too long"}, true
	}
	if m := outOfRange.FindSubmatch(body); m != nil {
		return controllers.GeneralErrorResponse{Code: controllers.ErrorParse, Message: string(m[1]) + " is out of range"}, true
	}
	if duplicateError.Match(body) {
		return controllers.GeneralErrorResponse{Code: controllers.ErrorInsert, Message: "already exists"}, true
	}
	return controllers.GeneralErrorResponse{}, false
}

// errorWriter holds back the status of an error response until its body is
// known, so a database error can be answered as the client's mistake.
type errorWriter struct {
	gin.ResponseWriter
	status int
}

func (w *errorWriter) WriteHeader(code int) {
	if code >= http.StatusBadRequest {
		w.status = code
		return
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *errorWriter) flush() {
	if w.status != 0 {
		w.ResponseWriter.WriteHeader(w.status)
		w.status = 0
	}
}

func (w *errorWriter) WriteHeaderNow() {
	w.flush()
	w.ResponseWriter.WriteHeaderNow()
}

func (w *errorWriter) Status() int {
	if w.status != 0 {
		return w.status
	}
	return w.ResponseWriter.Status()
}

func (w *errorWriter) Write(b []byte) (int, error) {
	if w.status != 0 {
		if resp, ok := clientDBError(b); ok {
			w.status = http.StatusBadRequest
			w.flush()
			out, _ := json.Marshal(resp)
			if _, err := w.ResponseWriter.Write(out); err != nil {
				return 0, err
			}
			return len(b), nil
		}
	}
	w.flush()
	return w.ResponseWriter.Write(b)
}

func (w *errorWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

// dbErrors answers input the database refuses (too long, out of range,
// duplicate) with 400 instead of a raw driver message.
func dbErrors(c *gin.Context) {
	if !site.IsAPIPath(c.Request.URL.Path) {
		c.Next()
		return
	}
	w := &errorWriter{ResponseWriter: c.Writer}
	c.Writer = w
	c.Next()
	w.flush()
}
