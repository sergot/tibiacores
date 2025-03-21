package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// In development, use a default secret
		if os.Getenv("APP_ENV") != "production" {
			secret = "dev-secret-key"
		} else {
			panic("JWT_SECRET environment variable must be set in production")
		}
	}
	jwtSecret = []byte(secret)
}

type Claims struct {
	UserID   string `json:"user_id"`
	HasEmail bool   `json:"has_email"`
	jwt.RegisteredClaims
}

func GenerateToken(userID string, hasEmail bool) (string, error) {
	claims := Claims{
		UserID:   userID,
		HasEmail: hasEmail,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 30)), // 30 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("empty token")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func GenerateAnonymousToken() (string, string) {
	anonID := fmt.Sprintf("anon_%d", time.Now().UnixNano())
	token, _ := GenerateToken(anonID, false)
	return token, anonID
}
