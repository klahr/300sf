package wikipedia

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"git.sr.ht/~klahr/300sf/config"
	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/utils"
	gowiki "github.com/trietmn/go-wiki"
)

type Data struct {
	Url       string   `json:"url"`
	Thumbnail string   `json:"thumbnail"`
	Summary   string   `json:"summary"`
	Images    []string `json:"images"`
}

func GetData(lang string, species models.Species) Data {
	var data Data

	gowiki.SetLanguage(lang)
	gowiki.SetCacheDuration(time.Second * time.Duration(config.AppConfig.WikipediaCacheDuration))

	name := utils.Translation2Value(species.Name, lang)
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

	q := fmt.Sprintf("https://%s.wikipedia.org/w/api.php?action=query&titles=%s&prop=pageimages&format=json&pithumbsize=200", lang, url.QueryEscape(utils.Translation2Value(species.Name, lang)))

	r, err := http.Get(q)
	if err == nil {
		defer r.Body.Close()
		body, _ := io.ReadAll(r.Body)
		begin := strings.Index(string(body), "\"source\":\"") + 10
		end := strings.Index(string(body[begin:]), "\"")
		thumbnail := string(body[begin : begin+end])
		if strings.HasPrefix(thumbnail, "http") {
			data.Thumbnail = thumbnail
		}
	}

	return data
}
