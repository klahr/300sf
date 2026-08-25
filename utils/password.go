// SPDX-License-Identifier: GPL-2.0-or-later
package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	Value string
	Salt  string
}

func (p *Password) Create(str string) error {
	salt := RandomSha1()
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

// RandomSha1 returns a cryptographically random hex string of the same length
// as a sha1 digest. Existing hashes are unaffected: verification always uses
// the salt stored alongside the hash, never a regenerated one.
func RandomSha1() string {
	return randomHex(sha1.Size)
}

// RandomSha256 returns a cryptographically random hex string of the same length
// as a sha256 digest.
func RandomSha256() string {
	return randomHex(sha256.Size)
}

func randomHex(n int) string {
	b := make([]byte, n)
	// crypto/rand.Read never fails, it fills b entirely or crashes the process.
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func ComputeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
