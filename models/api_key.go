package models

import "gopkg.in/guregu/null.v4/zero"

type ApiKey struct {
	Model
	Name          string     `json:"name"`
	Prefix        string     `json:"prefix"`
	Hash          string     `json:"hash"`
	Scopes        int64      `json:"scopes"`
	ApiKeyTierId  int64      `json:"-"`
	ApiKeyTier    ApiKeyTier `json:"api_key_tier"`
	UserId        int64      `json:"-"`
	User          User       `json:"user"`
	DeactivatedAt zero.Time  `json:"deactivated_at"`
}
