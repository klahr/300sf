package auth

import (
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type AuthenticationResponse struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func Authenticate(uuid string, availablePermissions int64, permissions string, accessTokenLifetime int64, refreshTokenLifetime int64, privateKey []byte) *AuthenticationResponse {
	now := time.Now().Unix()
	accessTokenClaims := Token{
		uuid,
		0,
		jwt.StandardClaims{
			ExpiresAt: now + accessTokenLifetime,
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

	requestedPermissions := strings.Split(permissions, ",")
	for i := range requestedPermissions {
		permission := requestedPermissions[i]
		var add int64 = 0
		switch permission {
		case "basic":
			add = AuthPermissionBasic
		case "edit":
			add = AuthPermissionEdit
		case "admin":
			add = AuthPermissionAdmin
		}

		if add > 0 {
			if availablePermissions&add != 0 {
				accessTokenClaims.Permissions |= add
			} else {
				//				Logger.Warning(ctx, "ApiAuthenticationV1.Auth", ErrorCheat, "user '"+*userDto.Uuid+"' is trying to request non-granted permission")
			}
		}
	}

	refreshTokenClaims.Permissions = accessTokenClaims.Permissions

	accessTokenWithClaims := jwt.NewWithClaims(jwt.SigningMethodRS256, accessTokenClaims)
	key, err := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		return nil
	}
	accessToken, err := accessTokenWithClaims.SignedString(key)

	refreshTokenWithClaims := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshTokenClaims)
	refreshToken, err := refreshTokenWithClaims.SignedString(key)

	authenticationResponseSuccess := AuthenticationResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
	}

	return &authenticationResponseSuccess
}
