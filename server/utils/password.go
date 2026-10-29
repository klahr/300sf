// SPDX-License-Identifier: AGPL-3.0-or-later
package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	Value string
	Salt  string
}

// prehashSalt is stored as the salt of passwords hashed by Create. Older
// hashes have a random salt appended to the password instead, which left only
// 32 bytes of bcrypt's 72-byte input for the password itself.
const prehashSalt = "sha256"

// prehash lets passwords of any length use all of bcrypt's input.
func prehash(str string) []byte {
	sum := sha256.Sum256([]byte(str))
	return []byte(base64.StdEncoding.EncodeToString(sum[:]))
}

func (p *Password) Create(str string) error {
	p.Salt = prehashSalt
	if hashedPassword, err := bcrypt.GenerateFromPassword(prehash(str), bcrypt.DefaultCost); err != nil {
		return err
	} else {
		p.Value = string(hashedPassword)
	}
	return nil
}

func (p *Password) IsMatching(str, password, salt string) bool {
	if salt == prehashSalt {
		return bcrypt.CompareHashAndPassword([]byte(password), prehash(str)) == nil
	}
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

// NewToken returns a random token for a link sent by email, and the hash to
// store in its place, so a copy of the database can't be used to follow it.
func NewToken() (token string, hash string) {
	token = RandomSha256()
	return token, HashToken(token)
}

// HashToken returns the stored form of a token made by NewToken.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
