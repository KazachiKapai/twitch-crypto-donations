package jwt

import (
	"fmt"
	"time"
	"twitch-crypto-donations/internal/pkg/environment"

	"github.com/golang-jwt/jwt/v4"
)

type UserClaims struct {
	Address string `json:"address"`
	jwt.RegisteredClaims
}

type Manager struct {
	tokenExpiration environment.TokenExpirationHours
	jwtSecret       environment.JwtSecret
}

func New(
	tokenExpiration environment.TokenExpirationHours,
	jwtSecret environment.JwtSecret,
) *Manager {
	return &Manager{
		tokenExpiration: tokenExpiration,
		jwtSecret:       jwtSecret,
	}
}

func (m *Manager) GenerateJwt(address string) (string, error) {
	expirationTime := time.Now().UTC().Add(time.Hour * time.Duration(m.tokenExpiration))

	claims := &UserClaims{
		Address: address,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   address,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(m.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
