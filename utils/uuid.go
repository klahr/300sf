package utils

import (
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

func GenerateUuid() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghjkmnpqrstuvwxyz123456789"
const (
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

func GenerateUuidWithFixedLength(n int) string {
	b := make([]byte, n)
	l := len(letterBytes)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < l {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}
