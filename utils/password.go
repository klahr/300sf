package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

type Password struct {
	Value string
	Salt  string
}

func (p *Password) Create(str string) {
	salt := p.randomSha256()
	p.Salt = salt
	p.Value = p.computeHmac256(str, salt)
}

func (p *Password) IsMatching(str, password, salt string) bool {
	return p.computeHmac256(str, salt) == password
}

func (p *Password) randomSha256() string {
	hash := hmac.New(sha256.New, []byte(strconv.FormatInt(time.Now().Unix(), 10)))
	return hex.EncodeToString(hash.Sum(nil))
}

func (p *Password) computeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
