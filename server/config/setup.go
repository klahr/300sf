// SPDX-License-Identifier: AGPL-3.0-or-later
package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type App struct {
	DataSource             string   `yaml:"data-source"`
	Address                string   `yaml:"address"`
	AccessTokenLifetime    int64    `yaml:"access-token-lifetime"`
	RefreshTokenLifetime   int64    `yaml:"refresh-token-lifetime"`
	HashingSecret          string   `yaml:"hashing-secret"`
	Domain                 string   `yaml:"domain"`
	RedisHost              string   `yaml:"redis-host"`
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
	Site                   Site     `yaml:"site"`
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
}

func (c *App) Parse(data []byte) error {
	return yaml.Unmarshal(data, c)
}

var AppConfig *App

func SetupAppConfig(filename string) error {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	AppConfig = new(App)
	if err := AppConfig.Parse(raw); err != nil {
		return err
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
