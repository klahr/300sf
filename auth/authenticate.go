package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type AuthenticationResponse struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

func Authenticate(uuid string, availablePermissions int64, accessTokenLifetime int64, refreshTokenLifetime int64, privateKey []byte) *AuthenticationResponse {
	now := time.Now().Unix()
	expiresAt := now + accessTokenLifetime
	accessTokenClaims := Token{
		uuid,
		0,
		jwt.StandardClaims{
			ExpiresAt: expiresAt,
			IssuedAt:  now,
		},
	}
	refreshTokenClaims := Token{
		uuid,
		0,
		jwt.StandardClaims{
			ExpiresAt: now + refreshTokenLifetime,
			IssuedAt:  now,
		},
	}

	accessTokenClaims.Permissions = availablePermissions
	refreshTokenClaims.Permissions = accessTokenClaims.Permissions

	accessTokenWithClaims := jwt.NewWithClaims(jwt.SigningMethodRS256, accessTokenClaims)
	key, err := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		return nil
	}
	accessToken, err := accessTokenWithClaims.SignedString(key)
	if err != nil {
		return nil
	}
	refreshTokenWithClaims := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshTokenClaims)
	refreshToken, err := refreshTokenWithClaims.SignedString(key)
	if err != nil {
		return nil
	}
	authenticationResponseSuccess := AuthenticationResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		ExpiresAt:    expiresAt,
	}

	return &authenticationResponseSuccess
}
