// SPDX-License-Identifier: AGPL-3.0-or-later
package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type App struct {
	DataSource           string `yaml:"data-source"`
	Address              string `yaml:"address"`
	AccessTokenLifetime  int64  `yaml:"access-token-lifetime"`
	RefreshTokenLifetime int64  `yaml:"refresh-token-lifetime"`
	// Stores API keys and email checksums hashed; changing it breaks every
	// API key.
	HashingSecret string `yaml:"hashing-secret"`
	// Signs sessions (JWTs); changing it only signs everyone out, so it can
	// be rotated whenever it may have leaked.
	TokenSecret string `yaml:"token-secret"`
	Domain      string `yaml:"domain"`
	RedisHost   string `yaml:"redis-host"`
	// Redis holds the sessions; set its requirepass so nothing else on the
	// host can read or change them.
	RedisPassword          string   `yaml:"redis-password"`
	LogLevel               string   `yaml:"log-level"`
	DefaultApiSoftLimit    int64    `yaml:"default-api-soft-limit"`
	DefaultApiHardLimit    int64    `yaml:"default-api-hard-limit"`
	ApiCallCounterTTL      int64    `yaml:"api-call-counter-ttl"`
	DefaultLists           []string `yaml:"default-lists"`
	WikipediaCacheDuration int64    `yaml:"wikipedia-cache-duration"`
	MailjetPublicKey       string   `yaml:"mailjet-public-key"`
	MailjetSecretKey       string   `yaml:"mailjet-secret-key"`
	Mode                   string   `yaml:"mode"`
	SendEmailInterval      int64    `yaml:"send-email-interval"`
	DeleteAccountsInterval int64    `yaml:"delete-accounts-interval"`
	// Nominatim-compatible service that names the country, region and
	// municipality of sightings in places no zone covers yet. Empty disables
	// it. See geocoder/ for the public service's usage policy.
	GeocoderEndpoint string `yaml:"geocoder-endpoint"`
	// Where the contact form's messages go, sent from this address too.
	// Empty turns the form off.
	ContactEmail string `yaml:"contact-email"`
	Site         Site   `yaml:"site"`
}

// Site configures the websites served next to the API.
type Site struct {
	// The one public host, e.g. 300sf.se.
	Host string `yaml:"host"`
	// Old hosts that redirect permanently to Host, keeping the path. The value
	// replaces "/" (app.300sf.se/ -> /home). API requests on these hosts are
	// answered rather than redirected, since API clients don't follow
	// redirects on POST.
	RedirectHosts map[string]string `yaml:"redirect-hosts"`
	// API key the app signs in with; added on the server so it never reaches
	// the browser.
	ClientId     string `yaml:"client-id"`
	ClientSecret string `yaml:"client-secret"`
	// Bird sound identification service; empty disables /analyze.
	AnalyzeEndpoint string `yaml:"analyze-endpoint"`
	// The maps' raster tiles ({z}/{x}/{y} URL) and their attribution (HTML).
	// OpenStreetMap's when empty; set them to move to another tile service.
	MapTiles       string `yaml:"map-tiles"`
	MapAttribution string `yaml:"map-attribution"`
	// Where visitors get the source of the code running here, which the
	// AGPL requires: point it at your fork if you change the code.
	SourceUrl string `yaml:"source-url"`
}

const (
	defaultSourceUrl      = "https://github.com/klahr/300sf"
	defaultMapTiles       = "https://tile.openstreetmap.org/{z}/{x}/{y}.png"
	defaultMapAttribution = `&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors`
)

func (c *App) Parse(data []byte) error {
	if err := yaml.Unmarshal(data, c); err != nil {
		return err
	}
	if c.Site.SourceUrl == "" {
		c.Site.SourceUrl = defaultSourceUrl
	}
	if c.Site.MapTiles == "" {
		c.Site.MapTiles = defaultMapTiles
		c.Site.MapAttribution = defaultMapAttribution
	}
	return nil
}

var AppConfig *App

const minTokenSecretLength = 32

func SetupAppConfig(filename string) error {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	AppConfig = new(App)
	if err := AppConfig.Parse(raw); err != nil {
		return err
	}
	return AppConfig.Validate()
}

// Validate refuses settings that would make the server unsafe to run.
func (c *App) Validate() error {
	// The secret signs every token: with a known one anyone can sign in as
	// anyone, admins included.
	if c.HashingSecret == "" || c.HashingSecret == "change-me" {
		return errors.New("hashing-secret is not set; generate one with: openssl rand -hex 32")
	}
	// The token secret signs every session: with a known or guessable one
	// anyone can sign in as anyone, admins included. Any signed-in user holds
	// a token to brute-force it against, so it must be long and random.
	if c.TokenSecret == "" || c.TokenSecret == "change-me" {
		return errors.New("token-secret is not set; generate one with: openssl rand -hex 32")
	}
	if len(c.TokenSecret) < minTokenSecretLength {
		return fmt.Errorf("token-secret is too short; use at least %d characters: openssl rand -hex 32", minTokenSecretLength)
	}
	// Older installations signed sessions with hashing-secret, which can't be
	// rotated without breaking API keys.
	if c.TokenSecret == c.HashingSecret {
		return errors.New("token-secret must differ from hashing-secret; generate one with: openssl rand -hex 32")
	}
	return nil
}

func GetLogLevelAsInt() int {
	switch AppConfig.LogLevel {
	case "silent":
		return 1
	case "error":
		return 2
	case "warn":
		return 3
	case "info":
		return 4
	}
	return 3
}
