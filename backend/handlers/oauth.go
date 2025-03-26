package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/sergot/tibiacores/backend/auth"
	"github.com/sergot/tibiacores/backend/db"
)

type OAuthHandler struct {
	connPool *pgxpool.Pool
}

func NewOAuthHandler(connPool *pgxpool.Pool) *OAuthHandler {
	return &OAuthHandler{connPool}
}

// Login initiates OAuth2 flow for the specified provider
func (h *OAuthHandler) Login(c echo.Context) error {
	provider := c.Param("provider")
	redirectURL, err := auth.GetOAuthRedirect(provider)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.String(http.StatusOK, redirectURL)
}

// Callback handles OAuth2 callback from providers
func (h *OAuthHandler) Callback(c echo.Context) error {
	provider := c.Param("provider")
	state := c.QueryParam("state")
	code := c.QueryParam("code")

	if !auth.ValidateOAuthState(state) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid oauth state")
	}

	queries := db.New(h.connPool)
	ctx := context.Background()

	// Exchange code for token
	userInfo, err := auth.ExchangeCodeForUser(provider, code)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to authenticate with provider")
	}

	// Find or create user
	var email pgtype.Text
	email.String = userInfo.Email
	email.Valid = true

	// First check if a user exists with this email
	existingUser, err := queries.GetUserByEmail(ctx, email)
	if err == nil {
		// User exists with this email
		if existingUser.Password.Valid {
			// User exists with password, meaning it's not an OAuth user
			return echo.NewHTTPError(http.StatusConflict, "email already in use with a different account type")
		}
		// Existing OAuth user, generate token and return
		token, err := auth.GenerateToken(existingUser.ID.String(), true)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
		}

		c.Response().Header().Set("X-Auth-Token", token)
		return c.JSON(http.StatusOK, map[string]interface{}{
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
			existingAnonymousUser, err := queries.GetUserByID(ctx, userID)
			if err == nil && existingAnonymousUser.IsAnonymous {
				// Migrate the anonymous user to OAuth user
				user, err = queries.MigrateAnonymousUser(ctx, db.MigrateAnonymousUserParams{
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
						return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
					}

					c.Response().Header().Set("X-Auth-Token", token)
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
	user, err = queries.CreateUser(ctx, db.CreateUserParams{
		Email:                      email,
		Password:                   pgtype.Text{},        // No password for OAuth users
		EmailVerificationToken:     uuid.Nil,             // OAuth users don't need verification
		EmailVerificationExpiresAt: pgtype.Timestamptz{}, // No expiry needed
		EmailVerified:              true,                 // OAuth users are already verified
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID.String(), true)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
	}

	// Set token in X-Auth-Token header
	c.Response().Header().Set("X-Auth-Token", token)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":        user.ID,
		"has_email": true,
	})
}
