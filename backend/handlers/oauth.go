package handlers

import (
	"context"
	"net/http"

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
	return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
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

	// Exchange code for token (implementation will be added per provider)
	userInfo, err := auth.ExchangeCodeForUser(provider, code)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to authenticate with provider")
	}

	// Find or create user
	var email pgtype.Text
	email.String = userInfo.Email
	email.Valid = true

	user, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		// Create new user if not found
		user, err = queries.CreateUser(ctx, db.CreateUserParams{
			Email:    email,
			Password: pgtype.Text{}, // No password for OAuth users
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
		}
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID.String(), false)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
	}

	// Redirect to frontend with token
	return c.Redirect(http.StatusTemporaryRedirect,
		auth.GetFrontendCallbackURL(token, user.ID.String()))
}
