package handlers

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/sergot/tibiacores/backend/auth"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
)

type OAuthHandler struct {
	store         db.Store
	oauthProvider auth.OAuthProvider
}

func NewOAuthHandler(store db.Store) *OAuthHandler {
	return &OAuthHandler{
		store:         store,
		oauthProvider: auth.NewDefaultOAuthProvider(),
	}
}

// SetOAuthProvider allows setting a custom OAuth provider (useful for testing)
func (h *OAuthHandler) SetOAuthProvider(provider auth.OAuthProvider) {
	h.oauthProvider = provider
}

// Login initiates OAuth2 flow for the specified provider
func (h *OAuthHandler) Login(c echo.Context) error {
	provider := c.Param("provider")
	redirectURL, state, err := auth.GetOAuthRedirect(provider)
	if err != nil {
		return apperror.ValidationError("Unable to get OAuth redirect URL", err)
	}

	// Set state cookie for CSRF protection
	// Note: SameSite=None requires Secure=true (HTTPS only, except localhost)
	cookie := new(http.Cookie)
	cookie.Name = "oauth_state"
	cookie.Value = state
	cookie.Path = "/"
	cookie.Expires = time.Now().Add(5 * time.Minute)
	cookie.HttpOnly = true
	cookie.Secure = true                    // Required for SameSite=None
	cookie.SameSite = http.SameSiteNoneMode // Required for cross-site OAuth redirects
	c.SetCookie(cookie)

	slog.Info("OAuth state cookie set",
		"provider", provider,
		"state", state,
	)

	return c.String(http.StatusOK, redirectURL)
}

// Callback handles OAuth2 callback from providers
func (h *OAuthHandler) Callback(c echo.Context) error {
	provider := c.Param("provider")
	state := c.QueryParam("state")
	code := c.QueryParam("code")

	// Get state from cookie
	cookie, err := c.Cookie("oauth_state")
	var cookieState string
	if err == nil {
		cookieState = cookie.Value
	}

	slog.Info("OAuth callback received",
		"provider", provider,
		"state_from_query", state,
		"state_from_cookie", cookieState,
		"cookie_found", err == nil,
	)

	// Clear the state cookie
	clearCookie := new(http.Cookie)
	clearCookie.Name = "oauth_state"
	clearCookie.Value = ""
	clearCookie.Path = "/"
	clearCookie.Expires = time.Now().Add(-1 * time.Hour)
	clearCookie.HttpOnly = true
	clearCookie.Secure = true
	clearCookie.SameSite = http.SameSiteNoneMode
	c.SetCookie(clearCookie)

	if !h.oauthProvider.ValidateState(cookieState, state) {
		return apperror.ValidationError("Invalid OAuth state", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "state",
				Value:  state,
				Reason: "State validation failed or expired",
			})
	}

	ctx := c.Request().Context()

	// Exchange code for token
	userInfo, err := h.oauthProvider.ExchangeCode(ctx, provider, code)
	if err != nil {
		return apperror.ExternalServiceError("Failed to authenticate with provider", err).
			WithDetails(&apperror.ExternalServiceErrorDetails{
				Service:   "OAuth",
				Operation: "ExchangeCode",
				Endpoint:  provider,
			}).
			Wrap(err)
	}

	// Find or create user
	var email pgtype.Text
	email.String = userInfo.Email
	email.Valid = true

	// First check if a user exists with this email
	existingUser, err := h.store.GetUserByEmail(ctx, email)
	if err == nil {
		// User exists with this email
		if existingUser.Password.Valid {
			// User exists with password, meaning it's not an OAuth user
			return apperror.ValidationError("Email already in use with a different account type", nil).
				WithDetails(&apperror.ValidationErrorDetails{
					Field:  "email",
					Value:  userInfo.Email,
					Reason: "Account type mismatch",
				})
		}

		// Existing OAuth user, generate token and return
		token, err := auth.GenerateToken(existingUser.ID.String(), true)
		if err != nil {
			return apperror.InternalError("Failed to generate token", err).
				WithContext(apperror.ErrorContext{
					Operation: "GenerateToken",
					UserID:    existingUser.ID.String(),
				}).
				Wrap(err)
		}

		c.Response().Header().Set("X-Auth-Token", token)
		return c.JSON(http.StatusOK, map[string]any{
			"id":        existingUser.ID,
			"has_email": true,
		})
	}

	// Check for existing session token
	authHeader := c.Request().Header.Get("Authorization")
	var existingUserID *string
	if authHeader != "" {
		if claims, err := auth.ValidateToken(strings.TrimPrefix(authHeader, "Bearer ")); err == nil {
			existingUserID = &claims.UserID
		}
	}

	var user db.User
	if existingUserID != nil {
		// Try to migrate existing anonymous user
		userID, parseErr := uuid.Parse(*existingUserID)
		if parseErr == nil {
			// Check if user exists and is anonymous
			existingAnonymousUser, err := h.store.GetUserByID(ctx, userID)
			if err == nil && existingAnonymousUser.IsAnonymous {
				// Migrate the anonymous user to OAuth user
				user, err = h.store.MigrateAnonymousUser(ctx, db.MigrateAnonymousUserParams{
					Email:                      email,
					Password:                   pgtype.Text{}, // No password for OAuth users
					EmailVerificationToken:     uuid.Nil,      // OAuth users don't need verification
					EmailVerificationExpiresAt: pgtype.Timestamptz{},
					ID:                         userID,
				})
				if err == nil {
					// Successfully migrated anonymous user
					token, err := auth.GenerateToken(user.ID.String(), true)
					if err != nil {
						return apperror.InternalError("Failed to generate token", err).
							WithContext(apperror.ErrorContext{
								Operation: "GenerateToken",
								UserID:    user.ID.String(),
							}).
							Wrap(err)
					}

					c.Response().Header().Set("X-Auth-Token", token)
					return c.JSON(http.StatusOK, map[string]any{
						"id":        user.ID,
						"has_email": true,
					})
				}
			}
		}
	}

	// If we get here, either there was no anonymous user to migrate or migration failed
	// Create new user
	user, err = h.store.CreateUser(ctx, db.CreateUserParams{
		Email:                      email,
		Password:                   pgtype.Text{},        // No password for OAuth users
		EmailVerificationToken:     uuid.Nil,             // OAuth users don't need verification
		EmailVerificationExpiresAt: pgtype.Timestamptz{}, // No expiry needed
		EmailVerified:              true,                 // OAuth users are already verified
	})
	if err != nil {
		return apperror.DatabaseError("Failed to create user", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "CreateUser",
				Table:     "users",
			}).
			Wrap(err)
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID.String(), true)
	if err != nil {
		return apperror.InternalError("Failed to generate token", err).
			WithContext(apperror.ErrorContext{
				Operation: "GenerateToken",
				UserID:    user.ID.String(),
			}).
			Wrap(err)
	}

	// Set token in X-Auth-Token header
	c.Response().Header().Set("X-Auth-Token", token)

	return c.JSON(http.StatusOK, map[string]any{
		"id":        user.ID,
		"has_email": true,
	})
}
