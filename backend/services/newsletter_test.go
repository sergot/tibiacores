package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewsletterService_validateEmail(t *testing.T) {
	service := &NewsletterService{}

	testCases := []struct {
		name        string
		email       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid email",
			email:       "test@example.com",
			expectError: false,
		},
		{
			name:        "valid email with plus",
			email:       "test+newsletter@example.com",
			expectError: false,
		},
		{
			name:        "empty string",
			email:       "",
			expectError: true,
			errorMsg:    "email cannot be empty",
		},
		{
			name:        "whitespace only",
			email:       "   ",
			expectError: true,
			errorMsg:    "email cannot be empty",
		},
		{
			name:        "invalid format - no @",
			email:       "invalid-email",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "invalid format - no domain",
			email:       "test@",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "invalid format - no local part",
			email:       "@example.com",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "invalid format - multiple @",
			email:       "test@@example.com",
			expectError: true,
			errorMsg:    "invalid email format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.validateEmail(tc.email)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewsletterService_Subscribe_EmailValidation(t *testing.T) {
	// This test focuses on the email validation part of Subscribe method
	// We don't need to mock the HTTP client since validation happens before the API call

	service := &NewsletterService{
		apiKey:  "test-key",
		listID:  "test-list",
		baseURL: "https://test.com",
	}

	testCases := []struct {
		name        string
		email       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty email",
			email:       "",
			expectError: true,
			errorMsg:    "invalid email: email cannot be empty",
		},
		{
			name:        "invalid email format",
			email:       "invalid-email",
			expectError: true,
			errorMsg:    "invalid email: invalid email format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.Subscribe(context.Background(), tc.email)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
