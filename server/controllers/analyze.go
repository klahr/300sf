package controllers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/config"
	"github.com/klahr/300sf/server/log"
	"github.com/klahr/300sf/server/models"
)

// A 10 second recording is about 1 MB as 48 kHz WAV.
const maxRecordingBytes = 10 << 20

// Each identification ties up the service, which runs one at a time.
const maxAnalysesPerHour = 60

// The service names a handful of candidates; more is not a real answer.
const maxResolveBirds = 50

var analyzeTransport = &http.Transport{
	Proxy:                 http.ProxyFromEnvironment,
	ResponseHeaderTimeout: 60 * time.Second,
}

// Analyze passes a recording (multipart "audio" and "meta") to the bird sound identification service,
// site.analyze-endpoint, and returns its answer unchanged. Signed in only: identification is
// expensive to run.
func Analyze(c *gin.Context) {
	endpoint := config.AppConfig.Site.AnalyzeEndpoint
	target, err := url.Parse(endpoint)
	if endpoint == "" || err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    ErrorAnalyzeUnavailable,
			"message": "identification is not available",
		})
		return
	}

	if throttle(c, "analyze/"+c.GetString("uuid"), maxAnalysesPerHour, time.Hour) {
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRecordingBytes)
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.Out.URL = target
			r.Out.Host = target.Host
			// The service needs the recording, not the user's session.
			r.Out.Header.Del("Cookie")
			r.Out.Header.Del("Authorization")
		},
		Transport: analyzeTransport,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Error(c, ErrorAnalyzeUnavailable, err.Error())
			c.JSON(http.StatusBadGateway, gin.H{
				"code":    ErrorAnalyzeUnavailable,
				"message": "identification is not available",
			})
		},
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}

func AnalyzeResolve(c *gin.Context) {
	type bird struct {
		Name        string  `json:"name"`
		Probability float64 `json:"probability"`
	}
	type request struct {
		Birds []bird `json:"birds"`
	}
	type result struct {
		Species     models.Species `json:"species"`
		Probability float64        `json:"probability"`
	}
	type response struct {
		Results []result `json:"results"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	if len(r.Birds) > maxResolveBirds {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorInvalidRequest,
			"message": "too many birds",
		})
		return
	}

	tx := models.DB.Begin()
	defer tx.Rollback()

	var res response
	for _, b := range r.Birds {
		names := strings.Split(b.Name, "_")
		if len(names) < 2 {
			continue
		}
		var species models.Species
		if err := tx.Preload("Name").Joins("JOIN translations ON translations.id=name_id").Where("translations.en=? OR translations.la=?", names[1], names[0]).First(&species).Error; err != nil {
			continue
		}
		res.Results = append(res.Results, result{
			Species:     species,
			Probability: b.Probability,
		})
	}

	c.JSON(http.StatusOK, res)
}
