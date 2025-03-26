package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/sergot/fiendlist/backend/auth"
	"github.com/sergot/fiendlist/backend/db"
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
		// User exists, check if it's an OAuth user
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

	// Create new user if not found
	user, err := queries.CreateUser(ctx, db.CreateUserParams{
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
