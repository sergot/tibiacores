package auth

import (
	"fmt"
	"net/url"
	"os"

	"golang.org/x/oauth2"
)

var (
	oauthStateString = os.Getenv("OAUTH_STATE_STRING")
	frontendURL      = os.Getenv("FRONTEND_URL")

	// OAuth2 configuration - will be initialized in init()
	oauthConfigs = make(map[string]*oauth2.Config)
)

type OAuthUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email,omitempty"`
	Provider      string `json:"provider"`
}

func init() {
	if oauthStateString == "" {
		if os.Getenv("APP_ENV") != "production" {
			oauthStateString = "dev-state-string"
		} else {
			panic("OAUTH_STATE_STRING environment variable must be set in production")
		}
	}

	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
}

// GetOAuthRedirect returns the OAuth2 redirect URL for the specified provider
func GetOAuthRedirect(provider string) (string, error) {
	config, exists := oauthConfigs[provider]
	if !exists {
		return "", fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	return config.AuthCodeURL(oauthStateString), nil
}

// ValidateOAuthState validates the state parameter to prevent CSRF
func ValidateOAuthState(state string) bool {
	return state == oauthStateString
}

// PrepareOAuthProviders sets up OAuth2 configurations for supported providers
func PrepareOAuthProviders() {
	// Configurations will be added when implementing specific providers
	// not implemented Google/Discord yet
}

// GetUserInfoFromToken retrieves user information from the OAuth2 provider
func GetUserInfoFromToken(provider string, token *oauth2.Token) (*OAuthUserInfo, error) {
	// This will be implemented when adding specific providers
	return nil, fmt.Errorf("provider %s not implemented", provider)
}

// ExchangeCodeForUser exchanges OAuth2 code for user information
func ExchangeCodeForUser(provider string, code string) (*OAuthUserInfo, error) {
	config, exists := oauthConfigs[provider]
	if !exists {
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	token, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	return GetUserInfoFromToken(provider, token)
}

// GetFrontendCallbackURL generates the URL to redirect back to the frontend
func GetFrontendCallbackURL(token string, userID string) string {
	callbackURL, _ := url.Parse(frontendURL)
	callbackURL.Path = "/oauth/callback"

	q := callbackURL.Query()
	q.Set("token", token)
	q.Set("user_id", userID)
	callbackURL.RawQuery = q.Encode()

	return callbackURL.String()
}
