// SPDX-License-Identifier: GPL-2.0-or-later
package models

import "gopkg.in/guregu/null.v4/zero"

type ApiKey struct {
	Model
	Name      string    `json:"name"`
	Prefix    string    `json:"prefix"`
	Hash      string    `json:"-"`
	Scopes    int64     `json:"scopes"`
	SoftLimit int64     `json:"soft_limit"`
	HardLimit int64     `json:"hard_limit"`
	UserId    int64     `json:"-"`
	RevokedAt zero.Time `json:"revoked_at"`
}

type ApiKeyClean struct {
	Name      string    `json:"name"`
	Prefix    string    `json:"prefix"`
	Hash      string    `json:"-"`
	Scopes    int64     `json:"scopes"`
	SoftLimit int64     `json:"soft_limit"`
	HardLimit int64     `json:"hard_limit"`
	RevokedAt zero.Time `json:"revoked_at"`
}
