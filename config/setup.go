// SPDX-License-Identifier: GPL-2.0-or-later
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
	PrivateKeyFilename     string   `yaml:"private-key-filename"`
	PublicKeyFilename      string   `yaml:"public-key-filename"`
	HashingSecret          string   `yaml:"hashing-secret"`
	Domain                 string   `yaml:"domain"`
	ResourcePath           string   `yaml:"resource-path"`
	RedisHost              string   `yaml:"redis-host"`
	LogLevel               string   `yaml:"log-level"`
	DataStorage            string   `yaml:"data-storage"`
	DefaultApiSoftLimit    int64    `yaml:"default-api-soft-limit"`
	DefaultApiHardLimit    int64    `yaml:"default-api-hard-limit"`
	DefaultLists           []string `yaml:"default-lists"`
	CacheDuration          int64    `yaml:"cache-duration"`
	WikipediaCacheDuration int64    `yaml:"wikipedia-cache-duration"`
	MailjetPublicKey       string   `yaml:"mailjet-public-key"`
	MailjetSecretKey       string   `yaml:"mailjet-secret-key"`
	Mode                   string   `yaml:"mode"`
}

func (c *App) Parse(data []byte) error {
	return yaml.Unmarshal(data, c)
}

func (a *App) GetDataStorage() string {
	if a.DataStorage == "" {
		dataStoragePath := os.Getenv("DATA_STORAGE_PATH")
		if dataStoragePath != "" {
			return dataStoragePath
		} else {
			return "resources/"
		}
	}
	return a.DataStorage
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
