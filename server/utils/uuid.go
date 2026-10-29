// SPDX-License-Identifier: AGPL-3.0-or-later
package utils

import (
	"crypto/rand"
)

// shortUuidLength gives about 50 random bits. Some of these ids are the only
// thing that keeps a link-shared list private, so they must not be guessable.
// Ids made before were 6 characters and keep working.
const shortUuidLength = 10

func GenerateShortUuid() string {
	return GenerateUuidWithFixedLength(shortUuidLength)
}

// 32 characters, so each random byte maps onto one without bias. Easily
// confused characters (0, o, i, l) are left out.
const letterBytes = "abcdefghjkmnpqrstuvwxyz123456789"

func GenerateUuidWithFixedLength(n int) string {
	b := make([]byte, n)
	// crypto/rand.Read never fails, it fills b entirely or crashes the process.
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = letterBytes[b[i]%byte(len(letterBytes))]
	}
	return string(b)
}
