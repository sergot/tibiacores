package auth

import (
	"testing"
	"time"
)

func TestTokenRevocation(t *testing.T) {
	// Create a fresh blacklist for testing
	tokenBlacklist = &TokenBlacklist{
		blacklist: make(map[string]time.Time),
	}

	// Test token
	testToken := "test-token"

	// Initially the token shouldn't be revoked
	if IsRevoked(testToken) {
		t.Errorf("Expected token not to be revoked initially")
	}

	// Revoke the token
	expiryTime := time.Now().Add(time.Hour)
	RevokeToken(testToken, expiryTime)

	// Now it should be revoked
	if !IsRevoked(testToken) {
		t.Errorf("Expected token to be revoked after revocation")
	}

	// Test expired token cleanup
	expiredToken := "expired-token"
	RevokeToken(expiredToken, time.Now().Add(-time.Second)) // Already expired

	// Wait a moment to ensure time passes
	time.Sleep(10 * time.Millisecond)

	// The expired token should be auto-removed on check
	if IsRevoked(expiredToken) {
		t.Errorf("Expected expired token to be removed from blacklist")
	}

	// Manual cleanup should work
	RevokeToken("to-cleanup", time.Now().Add(-time.Second))
	CleanupBlacklist()
	if IsRevoked("to-cleanup") {
		t.Errorf("Expected token to be removed after cleanup")
	}

	// The non-expired token should still be revoked
	if !IsRevoked(testToken) {
		t.Errorf("Expected valid token to remain revoked after cleanup")
	}
}

func TestExtractTokenClaims(t *testing.T) {
	// We can't easily test this with real tokens without reimplementing token generation,
	// but we can verify the function handles error cases gracefully

	// Empty token
	claims, err := ExtractTokenClaims("")
	if err == nil || claims != nil {
		t.Errorf("Expected error for empty token")
	}

	// Invalid token format
	claims, err = ExtractTokenClaims("invalid-token-format")
	if err == nil || claims != nil {
		t.Errorf("Expected error for invalid token format")
	}
}
