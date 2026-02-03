package auth

import (
	"testing"
)

// MockOAuthProvider implements OAuth functionality for testing
type MockOAuthProvider struct {
	ValidateStateFn       func(cookieState, queryState string) bool
	ExchangeCodeForUserFn func(provider string, code string) (*OAuthUserInfo, error)
}

func (m *MockOAuthProvider) ValidateState(cookieState, queryState string) bool {
	if m.ValidateStateFn != nil {
		return m.ValidateStateFn(cookieState, queryState)
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
	tests := []struct {
		name        string
		cookieState string
		queryState  string
		want        bool
	}{
		{
			name:        "Valid matching states",
			cookieState: "test-state-123",
			queryState:  "test-state-123",
			want:        true,
		},
		{
			name:        "Mismatched states",
			cookieState: "test-state-123",
			queryState:  "different-state",
			want:        false,
		},
		{
			name:        "Empty cookie state",
			cookieState: "",
			queryState:  "test-state-123",
			want:        false,
		},
		{
			name:        "Empty query state",
			cookieState: "test-state-123",
			queryState:  "",
			want:        false,
		},
		{
			name:        "Both empty",
			cookieState: "",
			queryState:  "",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateOAuthState(tt.cookieState, tt.queryState); got != tt.want {
				t.Errorf("ValidateOAuthState() = %v, want %v", got, tt.want)
			}
		})
	}
}
