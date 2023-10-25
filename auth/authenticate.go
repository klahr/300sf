// SPDX-License-Identifier: GPL-2.0-or-later
package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type AuthenticationResponse struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
	MaxAge      int64  `json:"max_age"`
}

func Authenticate(uuid string, availablePermissions int64, accessTokenLifetime int64, privateKey []byte) (*AuthenticationResponse, error) {
	now := time.Now().Unix()
	expiresAt := now + accessTokenLifetime
	accessTokenClaims := Token{
		uuid,
		availablePermissions,
		jwt.StandardClaims{
			ExpiresAt: expiresAt,
			IssuedAt:  now,
		},
	}

	accessTokenWithClaims := jwt.NewWithClaims(jwt.SigningMethodRS256, accessTokenClaims)
	key, err := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		return nil, err
	}
	accessToken, err := accessTokenWithClaims.SignedString(key)
	if err != nil {
		return nil, err
	}
	authenticationResponseSuccess := AuthenticationResponse{
		AccessToken: accessToken,
		TokenType:   "bearer",
		ExpiresAt:   expiresAt,
		MaxAge:      accessTokenLifetime,
	}

	return &authenticationResponseSuccess, nil
}
