// SPDX-License-Identifier: AGPL-3.0-or-later
package site

import (
	"context"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/controllers"
	"github.com/klahr/300sf/server/live"
)

// The share pages' live connections, /share/<list>/live and
// /share/competition/<id>/live: a websocket that says "changed" whenever what
// the page shows may have, so that it fetches itself again (/scripts/live.js).

func (s *Site) liveList(c *gin.Context) {
	owner, err := s.listOwner(c.Param("uuid"))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	s.serveLive(c, []string{"user:" + owner}, nil)
}

func (s *Site) liveCompetition(c *gin.Context) {
	uuid := c.Param("uuid")
	keys, err := s.raceKeys(uuid)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	s.serveLive(c, keys, func() []string {
		// A competition that is no longer public is watched for nothing.
		keys, _ := s.raceKeys(uuid)
		return keys
	})
}

func (s *Site) serveLive(c *gin.Context, keys []string, refresh func() []string) {
	if controllers.Throttled("live/"+controllers.ClientNetwork(c.ClientIP()), 30, time.Minute) {
		c.Status(http.StatusTooManyRequests)
		return
	}
	// Watched from before the page is told it's connected, so no change falls in between.
	changes, stop := live.Watch(keys, refresh)
	defer stop()
	conn, err := websocket.Accept(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.CloseNow()
	// The page sends nothing; reading only notices when it goes away.
	ctx := conn.CloseRead(c.Request.Context())
	ping := time.NewTicker(30 * time.Second)
	defer ping.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-changes:
			if err := conn.Write(ctx, websocket.MessageText, []byte("changed")); err != nil {
				return
			}
		case <-ping.C:
			pctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := conn.Ping(pctx)
			cancel()
			if err != nil {
				return
			}
		}
	}
}
