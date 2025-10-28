package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"twitch-crypto-donations/internal/pkg/environment"
	_jwt "twitch-crypto-donations/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

const (
	AddressKey = "address"
	ClaimsKey  = "claims"
)

type JwtMiddleware struct {
	jwtSecret environment.JwtSecret
	logger    Logger
}

func NewJwtMiddleware(jwtSecret environment.JwtSecret, logger Logger) *JwtMiddleware {
	return &JwtMiddleware{
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

func (m *JwtMiddleware) Request() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be in format 'Bearer <token>'"})
			return
		}

		tokenString := parts[1]
		token, err := jwt.ParseWithClaims(tokenString, &_jwt.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Method.Alg())
			}

			return []byte(m.jwtSecret), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Invalid token: %v", err)})
			return
		}

		if claims, ok := token.Claims.(*_jwt.UserClaims); ok && token.Valid {
			c.Set(AddressKey, claims.Address)
			c.Set(ClaimsKey, claims)

			m.logger.Info("user claims", "address", claims.Address)

			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		}
	}
}
