package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type App struct {
	DataSource           string `json:"data_source"`
	Address              string `json:"address"`
	AccessTokenLifetime  int64  `json:"access_token_lifetime"`
	RefreshTokenLifetime int64  `json:"refresh_token_lifetime"`
	PrivateKeyFilename   string `json:"private_key_filename"`
	PublicKeyFilename    string `json:"public_key_filename"`
	Domain               string `json:"domain"`
	ResourcePath         string `json:"resource_path"`
	RedisHost            string `json:"redis_host"`
	LogLevel             string `json:"log_level"`
	DataStorage          string `json:"data_storage"`
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
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	AppConfig = new(App)
	return json.Unmarshal(raw, AppConfig)
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
