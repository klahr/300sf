// SPDX-License-Identifier: GPL-2.0-or-later
package wikipedia

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/klahr/300sf/server/config"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/utils"
	gowiki "github.com/trietmn/go-wiki"
)

type Data struct {
	Url       string   `json:"url"`
	Thumbnail string   `json:"thumbnail"`
	Summary   string   `json:"summary"`
	Images    []string `json:"images"`
}

const userAgent = "300sf/1.0 (https://300sf.se; support@300sf.se)"

// gowiki keeps its language in package state, so concurrent lookups in
// different languages must not interleave.
var gowikiMu sync.Mutex

var client = &http.Client{Timeout: 10 * time.Second}

// Language maps an app language to a supported Wikipedia edition. It is also
// what callers must use in cache keys.
func Language(lang string) string {
	switch lang {
	case "sv", "en", "de":
		return lang
	}
	return "en"
}

func GetData(lang string, pageName models.Translation) Data {
	var data Data
	lang = Language(lang)

	gowikiMu.Lock()
	defer gowikiMu.Unlock()

	gowiki.SetUserAgent(userAgent)
	gowiki.SetLanguage(lang)
	gowiki.SetCacheDuration(time.Second * time.Duration(config.AppConfig.WikipediaCacheDuration))

	name := utils.Translation2Value(pageName, lang)
	var format string
	if strings.ContainsAny(name, " ") {
		format = "\"%s\""
	} else {
		format = "%s"
	}
	p, err := gowiki.GetPage(fmt.Sprintf(format, name), -1, false, true)
	if err != nil {
		fmt.Println(err)
	} else {
		data.Url = p.URL
		data.Summary, _ = p.GetSummary()
		if images, err := p.GetImagesURL(); err == nil {
			for _, i := range images {
				if !strings.HasSuffix(i, ".svg") {
					data.Images = append(data.Images, i)
				}
			}
		}
	}

	q := fmt.Sprintf(
		"https://%s.wikipedia.org/w/api.php?action=query&titles=%s&prop=pageimages&format=json&pithumbsize=200",
		lang,
		url.QueryEscape(utils.Translation2Value(pageName, lang)),
	)

	req, err := http.NewRequest("GET", q, nil)
	if err != nil {
		return data
	}

	req.Header.Set(
		"User-Agent",
		userAgent,
	)
	req.Header.Set("Accept", "application/json")

	r, err := client.Do(req)
	if err == nil {
		defer r.Body.Close()
		body, _ := io.ReadAll(r.Body)

		begin := strings.Index(string(body), "\"source\":\"") + 10
		if begin >= 10 {
			end := strings.Index(string(body[begin:]), "\"")
			if end > 0 {
				thumbnail := string(body[begin : begin+end])
				if strings.HasPrefix(thumbnail, "http") {
					data.Thumbnail = thumbnail
				}
			}
		}
	}

	return data
}
