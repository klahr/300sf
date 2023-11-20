// SPDX-License-Identifier: GPL-2.0-or-later
package models

import (
	"strings"

	"gopkg.in/guregu/null.v4/zero"
)

type ApiKey struct {
	Model
	Name      string    `json:"name"`
	Prefix    string    `json:"prefix"`
	ClientId  string    `json:"client_id"`
	Hash      string    `json:"-"`
	Scopes    string    `json:"scopes"`
	SoftLimit int64     `json:"soft_limit"`
	HardLimit int64     `json:"hard_limit"`
	UserId    int64     `json:"-"`
	RevokedAt zero.Time `json:"revoked_at"`
}

type ApiKeyClean struct {
	Name      string    `json:"name"`
	Prefix    string    `json:"prefix"`
	ClientId  string    `json:"client_id"`
	Hash      string    `json:"-"`
	Scopes    string    `json:"scopes"`
	SoftLimit int64     `json:"soft_limit"`
	HardLimit int64     `json:"hard_limit"`
	RevokedAt zero.Time `json:"revoked_at"`
}

func (a *ApiKey) HasScope(scope string) bool {
	s := strings.Split(a.Scopes, " ")
	for i := range s {
		if s[i] == scope {
			return true
		}
	}
	return false
}
