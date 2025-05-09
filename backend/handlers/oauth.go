package handlers

import (
	"context"
	"net/http"
	"strings"

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
	redirectURL, err := auth.GetOAuthRedirect(provider)
	if err != nil {
		return apperror.ValidationError("Unable to get OAuth redirect URL", err)
	}

	return c.String(http.StatusOK, redirectURL)
}

// Callback handles OAuth2 callback from providers
func (h *OAuthHandler) Callback(c echo.Context) error {
	provider := c.Param("provider")
	state := c.QueryParam("state")
	code := c.QueryParam("code")

	if !h.oauthProvider.ValidateState(state) {
		return apperror.ValidationError("Invalid OAuth state", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "state",
				Value:  state,
				Reason: "State validation failed",
			})
	}

	ctx := context.Background()

	// Exchange code for token
	userInfo, err := h.oauthProvider.ExchangeCode(provider, code)
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

		// Generate token pair for cookies
		tokenPair, err := auth.GenerateTokenPair(existingUser.ID.String(), true)
		if err != nil {
			return apperror.InternalError("Failed to generate token pair", err).
				WithContext(apperror.ErrorContext{
					Operation: "GenerateTokenPair",
					UserID:    existingUser.ID.String(),
				}).
				Wrap(err)
		}

		// Set cookies for authentication
		auth.SetTokenCookies(c, tokenPair.AccessToken, tokenPair.RefreshToken, tokenPair.ExpiresIn)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":        existingUser.ID,
			"has_email": true,
		})
	}

	// Check for existing session token
	authHeader := c.Request().Header.Get("Authorization")
	var existingUserID *string
	if authHeader != "" {
		if claims, err := auth.ValidateAccessToken(strings.TrimPrefix(authHeader, "Bearer ")); err == nil {
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
					// Generate token pair for cookies
					tokenPair, err := auth.GenerateTokenPair(user.ID.String(), true)
					if err != nil {
						return apperror.InternalError("Failed to generate token pair", err).
							WithContext(apperror.ErrorContext{
								Operation: "GenerateTokenPair",
								UserID:    user.ID.String(),
							}).
							Wrap(err)
					}

					// Set cookies for authentication
					auth.SetTokenCookies(c, tokenPair.AccessToken, tokenPair.RefreshToken, tokenPair.ExpiresIn)

					return c.JSON(http.StatusOK, map[string]interface{}{
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

	// Generate token pair for cookies
	tokenPair, err := auth.GenerateTokenPair(user.ID.String(), true)
	if err != nil {
		return apperror.InternalError("Failed to generate token pair", err).
			WithContext(apperror.ErrorContext{
				Operation: "GenerateTokenPair",
				UserID:    user.ID.String(),
			}).
			Wrap(err)
	}

	// Set cookies for authentication
	auth.SetTokenCookies(c, tokenPair.AccessToken, tokenPair.RefreshToken, tokenPair.ExpiresIn)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":        user.ID,
		"has_email": true,
	})
}
