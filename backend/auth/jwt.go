package auth

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSecret []byte
var refreshSecret []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	refreshSecretStr := os.Getenv("REFRESH_TOKEN_SECRET")

	if secret == "" {
		// In development, use a default secret
		if os.Getenv("APP_ENV") != "production" {
			secret = "dev-secret-key"
		} else {
			panic("JWT_SECRET environment variable must be set in production")
		}
	}

	if refreshSecretStr == "" {
		// In development, use a default secret
		if os.Getenv("APP_ENV") != "production" {
			refreshSecretStr = "dev-refresh-secret-key"
		} else {
			panic("REFRESH_TOKEN_SECRET environment variable must be set in production")
		}
	}

	// If not in development, decode the base64 secrets
	if os.Getenv("APP_ENV") == "production" {
		decoded, err := base64.StdEncoding.DecodeString(secret)
		if err != nil {
			panic(fmt.Sprintf("JWT_SECRET must be a valid base64 string in production: %v", err))
		}
		jwtSecret = decoded

		decodedRefresh, err := base64.StdEncoding.DecodeString(refreshSecretStr)
		if err != nil {
			panic(fmt.Sprintf("REFRESH_TOKEN_SECRET must be a valid base64 string in production: %v", err))
		}
		refreshSecret = decodedRefresh
	} else {
		// In development, use the raw string
		jwtSecret = []byte(secret)
		refreshSecret = []byte(refreshSecretStr)
	}

	// Initialize token blacklist cleanup
	InitTokenBlacklistCleanup()
}

type Claims struct {
	UserID    string `json:"user_id"`
	HasEmail  bool   `json:"has_email"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds until access token expires
}

// GenerateAccessToken creates a short-lived JWT access token
func GenerateAccessToken(userID string, hasEmail bool) (string, error) {
	claims := Claims{
		UserID:    userID,
		HasEmail:  hasEmail,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // 15 minutes
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// GenerateRefreshToken creates a longer-lived refresh token
func GenerateRefreshToken(userID string, hasEmail bool) (string, error) {
	claims := Claims{
		UserID:    userID,
		HasEmail:  hasEmail,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(refreshSecret)
}

// GenerateTokenPair creates both access and refresh tokens
func GenerateTokenPair(userID string, hasEmail bool) (TokenPair, error) {
	accessToken, err := GenerateAccessToken(userID, hasEmail)
	if err != nil {
		return TokenPair{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := GenerateRefreshToken(userID, hasEmail)
	if err != nil {
		return TokenPair{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    15 * 60, // 15 minutes in seconds
	}, nil
}

// ValidateAccessToken validates an access token
func ValidateAccessToken(tokenString string) (*Claims, error) {
	return validateToken(tokenString, jwtSecret, "access")
}

// ValidateRefreshToken validates a refresh token
func ValidateRefreshToken(tokenString string) (*Claims, error) {
	return validateToken(tokenString, refreshSecret, "refresh")
}

// GenerateToken is an alias for GenerateAccessToken for backward compatibility
func GenerateToken(userID string, hasEmail bool) (string, error) {
	return GenerateAccessToken(userID, hasEmail)
}

// ValidateToken is an alias for ValidateAccessToken for backward compatibility
func ValidateToken(tokenString string) (*Claims, error) {
	return ValidateAccessToken(tokenString)
}

// validateToken validates a token with the specified secret and expected type
func validateToken(tokenString string, secret []byte, expectedType string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("empty token")
	}

	// Check if token is blacklisted (revoked)
	if IsRevoked(tokenString) {
		return nil, fmt.Errorf("token has been revoked")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Verify token type
		if claims.TokenType != expectedType {
			return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func GenerateAnonymousToken() (string, string) {
	anonID := fmt.Sprintf("anon_%d", time.Now().UnixNano())
	token, _ := GenerateAccessToken(anonID, false)
	return token, anonID
}
