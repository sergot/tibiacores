package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/sergot/tibiacores/backend/auth"
	mockdb "github.com/sergot/tibiacores/backend/db/mock"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/handlers"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func init() {
	// Initialize OAuth providers for testing
	auth.PrepareOAuthProviders()
}

type mockOAuthProvider struct {
	validateStateFn func(state string) bool
	exchangeCodeFn  func(provider, code string) (*auth.OAuthUserInfo, error)
}

func (m *mockOAuthProvider) ValidateState(state string) bool {
	if m.validateStateFn != nil {
		return m.validateStateFn(state)
	}
	return false
}

func (m *mockOAuthProvider) ExchangeCode(provider, code string) (*auth.OAuthUserInfo, error) {
	if m.exchangeCodeFn != nil {
		return m.exchangeCodeFn(provider, code)
	}
	return nil, nil
}

func TestLogin(t *testing.T) {
	testCases := []struct {
		name          string
		provider      string
		expectedCode  int
		expectedError string
	}{
		{
			name:         "Success - Discord",
			provider:     "discord",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Success - Google",
			provider:     "google",
			expectedCode: http.StatusOK,
		},
		{
			name:          "Invalid Provider",
			provider:      "invalid",
			expectedCode:  http.StatusBadRequest,
			expectedError: "unsupported OAuth provider: invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)

			// Create HTTP request
			req := httptest.NewRequest(http.MethodGet, "/api/auth/oauth/"+tc.provider, nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/auth/oauth/:provider")
			c.SetParamNames("provider")
			c.SetParamValues(tc.provider)

			// Execute handler
			h := handlers.NewOAuthHandler(store)
			err := h.Login(c)

			// Check for expected error response
			if tc.expectedError != "" {
				httpError, ok := err.(*echo.HTTPError)
				require.True(t, ok)
				require.Equal(t, tc.expectedCode, httpError.Code)
				require.Contains(t, httpError.Message, tc.expectedError)
				return
			}

			// Check successful response
			require.NoError(t, err)
			require.Equal(t, tc.expectedCode, rec.Code)

			// Check URL format based on provider
			if tc.provider == "google" {
				require.Contains(t, rec.Body.String(), "accounts.google.com/o/oauth2/v2/auth")
			} else {
				require.Contains(t, rec.Body.String(), "oauth2/authorize")
			}
		})
	}
}

func TestCallback(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, provider *mockOAuthProvider, email pgtype.Text)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response map[string]interface{}, headers http.Header)
	}{
		{
			name: "Success - New User",
			setupRequest: func(c echo.Context) {
				c.SetParamValues("discord")
				c.QueryParams().Set("state", "valid-state")
				c.QueryParams().Set("code", "valid-code")
			},
			setupMocks: func(store *mockdb.MockStore, provider *mockOAuthProvider, email pgtype.Text) {
				provider.validateStateFn = func(state string) bool {
					return state == "valid-state"
				}
				provider.exchangeCodeFn = func(provider, code string) (*auth.OAuthUserInfo, error) {
					if code == "valid-code" {
						return &auth.OAuthUserInfo{
							ID:            "oauth-user-id",
							Email:         "test@example.com",
							VerifiedEmail: true,
							Provider:      provider,
						}, nil
					}
					return nil, errors.New("invalid code")
				}

				// User doesn't exist yet
				store.EXPECT().
					GetUserByEmail(gomock.Any(), email).
					Return(db.User{}, errors.New("not found"))

				// Create new user
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(db.User{
						ID:            uuid.New(),
						Email:         email,
						EmailVerified: true,
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}, headers http.Header) {
				require.NotEmpty(t, response["id"])
				require.True(t, response["has_email"].(bool))
				require.NotEmpty(t, headers.Get("X-Auth-Token"))
			},
		},
		{
			name: "Success - Existing OAuth User",
			setupRequest: func(c echo.Context) {
				c.SetParamValues("discord")
				c.QueryParams().Set("state", "valid-state")
				c.QueryParams().Set("code", "valid-code")
			},
			setupMocks: func(store *mockdb.MockStore, provider *mockOAuthProvider, email pgtype.Text) {
				provider.validateStateFn = func(state string) bool {
					return state == "valid-state"
				}
				provider.exchangeCodeFn = func(provider, code string) (*auth.OAuthUserInfo, error) {
					if code == "valid-code" {
						return &auth.OAuthUserInfo{
							ID:            "oauth-user-id",
							Email:         "test@example.com",
							VerifiedEmail: true,
							Provider:      provider,
						}, nil
					}
					return nil, errors.New("invalid code")
				}

				// User exists with no password (OAuth user)
				store.EXPECT().
					GetUserByEmail(gomock.Any(), email).
					Return(db.User{
						ID:            uuid.New(),
						Email:         email,
						EmailVerified: true,
						Password:      pgtype.Text{}, // No password means OAuth user
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}, headers http.Header) {
				require.NotEmpty(t, response["id"])
				require.True(t, response["has_email"].(bool))
				require.NotEmpty(t, headers.Get("X-Auth-Token"))
			},
		},
		{
			name: "Success - Migrate Anonymous User",
			setupRequest: func(c echo.Context) {
				c.SetParamValues("discord")
				c.QueryParams().Set("state", "valid-state")
				c.QueryParams().Set("code", "valid-code")

				// Use a fixed test UUID
				existingUserID := uuid.MustParse("c6f36c1c-f957-4655-9dcd-89072ebaabda")
				token, err := auth.GenerateToken(existingUserID.String(), false)
				require.NoError(t, err)
				c.Request().Header.Set("Authorization", "Bearer "+token)
			},
			setupMocks: func(store *mockdb.MockStore, provider *mockOAuthProvider, email pgtype.Text) {
				provider.validateStateFn = func(state string) bool {
					return state == "valid-state"
				}
				provider.exchangeCodeFn = func(provider, code string) (*auth.OAuthUserInfo, error) {
					if code == "valid-code" {
						return &auth.OAuthUserInfo{
							ID:            "oauth-user-id",
							Email:         "test@example.com",
							VerifiedEmail: true,
							Provider:      provider,
						}, nil
					}
					return nil, errors.New("invalid code")
				}

				// Use the same fixed test UUID
				existingUserID := uuid.MustParse("c6f36c1c-f957-4655-9dcd-89072ebaabda")

				// User doesn't exist with this email
				store.EXPECT().
					GetUserByEmail(gomock.Any(), email).
					Return(db.User{}, errors.New("not found"))

				// Get existing anonymous user
				store.EXPECT().
					GetUserByID(gomock.Any(), existingUserID).
					Return(db.User{
						ID:          existingUserID,
						IsAnonymous: true,
					}, nil)

				// Migrate anonymous user
				store.EXPECT().
					MigrateAnonymousUser(gomock.Any(), db.MigrateAnonymousUserParams{
						ID:                         existingUserID,
						Email:                      email,
						Password:                   pgtype.Text{},
						EmailVerificationToken:     uuid.Nil,
						EmailVerificationExpiresAt: pgtype.Timestamptz{},
					}).
					Return(db.User{
						ID:            existingUserID,
						Email:         email,
						EmailVerified: true,
						IsAnonymous:   false,
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}, headers http.Header) {
				require.NotEmpty(t, response["id"])
				require.True(t, response["has_email"].(bool))
				require.NotEmpty(t, headers.Get("X-Auth-Token"))
			},
		},
		{
			name: "Invalid State",
			setupRequest: func(c echo.Context) {
				c.SetParamValues("discord")
				c.QueryParams().Set("state", "invalid-state")
				c.QueryParams().Set("code", "valid-code")
			},
			setupMocks: func(store *mockdb.MockStore, provider *mockOAuthProvider, email pgtype.Text) {
				provider.validateStateFn = func(state string) bool {
					return state == "valid-state"
				}
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid oauth state",
		},
		{
			name: "Email Already Used with Password Account",
			setupRequest: func(c echo.Context) {
				c.SetParamValues("discord")
				c.QueryParams().Set("state", "valid-state")
				c.QueryParams().Set("code", "valid-code")
			},
			setupMocks: func(store *mockdb.MockStore, provider *mockOAuthProvider, email pgtype.Text) {
				provider.validateStateFn = func(state string) bool {
					return state == "valid-state"
				}
				provider.exchangeCodeFn = func(provider, code string) (*auth.OAuthUserInfo, error) {
					if code == "valid-code" {
						return &auth.OAuthUserInfo{
							ID:            "oauth-user-id",
							Email:         "test@example.com",
							VerifiedEmail: true,
							Provider:      provider,
						}, nil
					}
					return nil, errors.New("invalid code")
				}

				// User exists with password
				store.EXPECT().
					GetUserByEmail(gomock.Any(), email).
					Return(db.User{
						ID:       uuid.New(),
						Email:    email,
						Password: pgtype.Text{String: "hashed-password", Valid: true},
					}, nil)
			},
			expectedCode:  http.StatusConflict,
			expectedError: "email already in use with a different account type",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			oauthProvider := &mockOAuthProvider{}

			// Create HTTP request
			req := httptest.NewRequest(http.MethodGet, "/api/auth/oauth/callback/discord", nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/auth/oauth/:provider/callback")

			// Custom request setup
			if tc.setupRequest != nil {
				tc.setupRequest(c)
			}

			// Setup mock expectations with the test email
			email := pgtype.Text{String: "test@example.com", Valid: true}
			if tc.setupMocks != nil {
				tc.setupMocks(store, oauthProvider, email)
			}

			// Execute handler
			h := handlers.NewOAuthHandler(store)
			h.SetOAuthProvider(oauthProvider)
			err := h.Callback(c)

			// Check for expected error response
			if tc.expectedError != "" {
				httpError, ok := err.(*echo.HTTPError)
				require.True(t, ok)
				require.Equal(t, tc.expectedCode, httpError.Code)
				require.Contains(t, httpError.Message, tc.expectedError)
				return
			}

			// Check successful response
			require.NoError(t, err)
			require.Equal(t, tc.expectedCode, rec.Code)

			// Check response body and headers
			if tc.checkResponse != nil {
				var response map[string]interface{}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, response, rec.Header())
			}
		})
	}
}
