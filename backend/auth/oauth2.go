package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

var (
	frontendURL = os.Getenv("FRONTEND_URL")
	// OAuth2 configuration - will be initialized in init()
	oauthConfigs = make(map[string]*oauth2.Config)
	// Map to store state strings temporarily
	stateStore = struct {
		sync.RWMutex
		states map[string]time.Time
	}{states: make(map[string]time.Time)}
)

type OAuthUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email,omitempty"`
	Provider      string `json:"provider"`
}

type DiscordUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Verified      bool   `json:"verified"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
}

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
}

func init() {
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
}

// generateState creates a new random state string and stores it
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	stateStore.Lock()
	stateStore.states[state] = time.Now().Add(5 * time.Minute)
	stateStore.Unlock()

	// Clean up expired states
	go cleanupExpiredStates()

	return state, nil
}

// cleanupExpiredStates removes expired state strings
func cleanupExpiredStates() {
	stateStore.Lock()
	defer stateStore.Unlock()

	now := time.Now()
	for state, expiry := range stateStore.states {
		if now.After(expiry) {
			delete(stateStore.states, state)
		}
	}
}

// GetOAuthRedirect returns the OAuth2 redirect URL for the specified provider
func GetOAuthRedirect(provider string) (string, error) {
	config, exists := oauthConfigs[provider]
	if !exists {
		return "", fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	state, err := generateState()
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	return config.AuthCodeURL(state), nil
}

// ValidateOAuthState validates the state parameter to prevent CSRF
func ValidateOAuthState(state string) bool {
	stateStore.RLock()
	expiry, exists := stateStore.states[state]
	stateStore.RUnlock()

	if !exists {
		return false
	}

	// Check if state has expired
	if time.Now().After(expiry) {
		stateStore.Lock()
		delete(stateStore.states, state)
		stateStore.Unlock()
		return false
	}

	// Remove the used state
	stateStore.Lock()
	delete(stateStore.states, state)
	stateStore.Unlock()

	return true
}

// PrepareOAuthProviders sets up OAuth2 configurations for supported providers
func PrepareOAuthProviders() {
	discordConfig := &oauth2.Config{
		ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
		ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("DISCORD_REDIRECT_URI"),
		Scopes:       []string{"identify", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}
	oauthConfigs["discord"] = discordConfig

	googleConfig := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}
	oauthConfigs["google"] = googleConfig
}

// GetUserInfoFromToken retrieves user information from the OAuth2 provider
func GetUserInfoFromToken(provider string, token *oauth2.Token) (*OAuthUserInfo, error) {
	switch provider {
	case "discord":
		return getDiscordUserInfo(token)
	case "google":
		return getGoogleUserInfo(token)
	default:
		return nil, fmt.Errorf("provider %s not implemented", provider)
	}
}

func getDiscordUserInfo(token *oauth2.Token) (*OAuthUserInfo, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("failed to close response body: %v\n", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var discordUser DiscordUser
	if err := json.Unmarshal(body, &discordUser); err != nil {
		return nil, err
	}

	return &OAuthUserInfo{
		ID:            discordUser.ID,
		Email:         discordUser.Email,
		VerifiedEmail: discordUser.Verified,
		Provider:      "discord",
	}, nil
}

func getGoogleUserInfo(token *oauth2.Token) (*OAuthUserInfo, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("failed to close response body: %v\n", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var googleUser GoogleUser
	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, err
	}

	return &OAuthUserInfo{
		ID:            googleUser.ID,
		Email:         googleUser.Email,
		VerifiedEmail: googleUser.VerifiedEmail,
		Provider:      "google",
	}, nil
}

// ExchangeCodeForUser exchanges OAuth2 code for user information
func ExchangeCodeForUser(provider string, code string) (*OAuthUserInfo, error) {
	config, exists := oauthConfigs[provider]
	if !exists {
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	token, err := config.Exchange(context.Background(), code)
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

// OAuthProvider defines the interface for OAuth operations
type OAuthProvider interface {
	ValidateState(state string) bool
	ExchangeCode(provider string, code string) (*OAuthUserInfo, error)
}

// DefaultOAuthProvider implements OAuthProvider using the standard OAuth flow
type DefaultOAuthProvider struct{}

func (p *DefaultOAuthProvider) ValidateState(state string) bool {
	return ValidateOAuthState(state)
}

func (p *DefaultOAuthProvider) ExchangeCode(provider string, code string) (*OAuthUserInfo, error) {
	return ExchangeCodeForUser(provider, code)
}

// NewDefaultOAuthProvider creates a new default OAuth provider
func NewDefaultOAuthProvider() OAuthProvider {
	return &DefaultOAuthProvider{}
}

var defaultProvider = NewDefaultOAuthProvider()

// These functions use the default provider for backward compatibility
func ValidateOAuthStateWithProvider(state string) bool {
	return defaultProvider.ValidateState(state)
}

func ExchangeCodeForUserWithProvider(provider string, code string) (*OAuthUserInfo, error) {
	return defaultProvider.ExchangeCode(provider, code)
}
