// SPDX-License-Identifier: GPL-2.0-or-later
package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	Value string
	Salt  string
}

func (p *Password) Create(str string) error {
	salt := RandomSha256()
	p.Salt = salt
	if hashedPassword, err := bcrypt.GenerateFromPassword([]byte(str+salt), bcrypt.DefaultCost); err != nil {
		return err
	} else {
		p.Value = string(hashedPassword)
	}
	return nil
}

func (p *Password) IsMatching(str, password, salt string) bool {
	return bcrypt.CompareHashAndPassword([]byte(password), []byte(str+salt)) == nil
}

func RandomSha256() string {
	hash := hmac.New(sha256.New, []byte(strconv.FormatInt(time.Now().Unix(), 10)))
	return hex.EncodeToString(hash.Sum(nil))
}

func ComputeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
