package auth

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenBlacklist stores revoked tokens along with their expiration time
type TokenBlacklist struct {
	blacklist map[string]time.Time
	mu        sync.RWMutex
}

// Create a singleton instance of the token blacklist
var tokenBlacklist = &TokenBlacklist{
	blacklist: make(map[string]time.Time),
}

// RevokeToken adds a token to the blacklist
func RevokeToken(tokenString string, expiresAt time.Time) {
	tokenBlacklist.mu.Lock()
	defer tokenBlacklist.mu.Unlock()
	tokenBlacklist.blacklist[tokenString] = expiresAt
}

// IsRevoked checks if a token is in the blacklist
func IsRevoked(tokenString string) bool {
	tokenBlacklist.mu.RLock()
	defer tokenBlacklist.mu.RUnlock()

	expiryTime, exists := tokenBlacklist.blacklist[tokenString]
	if !exists {
		return false
	}

	// If the token's blacklist entry has expired, we can remove it
	if time.Now().After(expiryTime) {
		// Using a goroutine to avoid locking during the read operation
		go func() {
			tokenBlacklist.mu.Lock()
			defer tokenBlacklist.mu.Unlock()
			delete(tokenBlacklist.blacklist, tokenString)
		}()
		return false
	}

	return true
}

// CleanupBlacklist removes expired tokens from the blacklist
func CleanupBlacklist() {
	tokenBlacklist.mu.Lock()
	defer tokenBlacklist.mu.Unlock()

	now := time.Now()
	for token, expiry := range tokenBlacklist.blacklist {
		if now.After(expiry) {
			delete(tokenBlacklist.blacklist, token)
		}
	}
}

// InitTokenBlacklistCleanup starts a periodic cleanup of the token blacklist
func InitTokenBlacklistCleanup() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			CleanupBlacklist()
		}
	}()
}

// ExtractTokenClaims extracts claims from a token without full validation
// This is useful for getting expiry times from potentially invalid tokens
func ExtractTokenClaims(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("empty token")
	}

	token, _ := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// For extraction purposes only, we don't need to validate the signature
		return []byte("extraction-only"), nil
	})

	if token != nil {
		if claims, ok := token.Claims.(*Claims); ok {
			return claims, nil
		}
	}

	return nil, fmt.Errorf("could not extract claims")
}
