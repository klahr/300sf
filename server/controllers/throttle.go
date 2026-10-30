// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/models"
)

// Throttled counts a call against key and reports whether more than limit
// calls were made within window. If Redis is unavailable the call is let
// through rather than failing the request.
func Throttled(key string, limit int64, window time.Duration) bool {
	if models.Cache == nil {
		return false
	}
	key = "/throttle/" + key
	n, err := models.Cache.Incr(key).Result()
	if err != nil {
		return false
	}
	// Also repair a counter left without expiry, or the key would be locked forever.
	if n == 1 || models.Cache.TTL(key).Val() < 0 {
		models.Cache.Expire(key, window)
	}
	return n > limit
}

// throttle answers 429 and returns true when the call is over the limit.
func throttle(c *gin.Context, key string, limit int64, window time.Duration) bool {
	if !Throttled(key, limit, window) {
		return false
	}
	c.JSON(http.StatusTooManyRequests, GeneralErrorResponse{Code: ErrorTooManyRequests, Message: "too many requests, try again later"})
	return true
}

// ClientNetwork is what limits per client count against: the address for
// IPv4, and its /64 for IPv6, since one connection is usually given a whole
// /64 to pick addresses from.
func ClientNetwork(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ip
	}
	if parsed.To4() != nil {
		return parsed.String()
	}
	return parsed.Mask(net.CIDRMask(64, 128)).String() + "/64"
}
