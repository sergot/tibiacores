package auth

import (
	"testing"
	"time"
)

// MockOAuthProvider implements OAuth functionality for testing
type MockOAuthProvider struct {
	ValidateStateFn       func(state string) bool
	ExchangeCodeForUserFn func(provider string, code string) (*OAuthUserInfo, error)
}

func (m *MockOAuthProvider) ValidateState(state string) bool {
	if m.ValidateStateFn != nil {
		return m.ValidateStateFn(state)
	}
	return false
}

func (m *MockOAuthProvider) ExchangeCode(provider string, code string) (*OAuthUserInfo, error) {
	if m.ExchangeCodeForUserFn != nil {
		return m.ExchangeCodeForUserFn(provider, code)
	}
	return nil, nil
}

func TestValidateOAuthState(t *testing.T) {
	// Reset state store
	stateStore.Lock()
	stateStore.states = make(map[string]time.Time)
	stateStore.Unlock()

	// Add test state
	testState := "test-state"
	stateStore.Lock()
	stateStore.states[testState] = time.Now().Add(5 * time.Minute)
	stateStore.Unlock()

	// Test valid state
	if !ValidateOAuthState(testState) {
		t.Error("Expected state to be valid")
	}

	// Test invalid state
	if ValidateOAuthState("invalid-state") {
		t.Error("Expected state to be invalid")
	}

	// Test expired state
	stateStore.Lock()
	stateStore.states["expired-state"] = time.Now().Add(-6 * time.Minute)
	stateStore.Unlock()

	if ValidateOAuthState("expired-state") {
		t.Error("Expected expired state to be invalid")
	}
}
