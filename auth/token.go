// SPDX-License-Identifier: GPL-2.0-or-later
package auth

import (
	"github.com/dgrijalva/jwt-go"
)

type Token struct {
	Uuid        string
	Permissions int64
	jwt.StandardClaims
}

func ParseToken(str string, claims *Token, publicKey []byte) error {
	token, err := jwt.ParseWithClaims(str, claims, func(token *jwt.Token) (interface{}, error) {
		key, err := jwt.ParseRSAPublicKeyFromPEM(publicKey)
		return key, err
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return jwt.ValidationError{}
	}
	return nil
}
