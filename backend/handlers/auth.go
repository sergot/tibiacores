package handlers

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/sergot/tibiacores/backend/auth"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
)

type AuthHandler struct {
	store db.Store
}

func NewAuthHandler(store db.Store) *AuthHandler {
	return &AuthHandler{store: store}
}

// RefreshTokenRequest represents the request body for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshToken validates a refresh token and issues a new token pair
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	refreshToken := ""

	// First try to get refresh token from cookie
	refreshCookie, err := c.Cookie("refresh_token")
	if err == nil && refreshCookie.Value != "" {
		refreshToken = refreshCookie.Value
	} else {
		// If no cookie, try to get from request body
		var req RefreshTokenRequest
		if err := c.Bind(&req); err == nil && req.RefreshToken != "" {
			refreshToken = req.RefreshToken
		}
	}

	// Validate that we have a token
	if refreshToken == "" {
		return apperror.ValidationError("Refresh token is required", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "refresh_token",
				Reason: "Missing refresh token",
			})
	}

	// Validate the refresh token
	claims, err := auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		return apperror.AuthorizationError("Invalid or expired refresh token", err).
			WithDetails(&apperror.AuthorizationErrorDetails{
				Reason: "refresh_token_invalid",
				Field:  "refresh_token",
			})
	}

	// Generate a new token pair
	tokenPair, err := auth.GenerateTokenPair(claims.UserID, claims.HasEmail)
	if err != nil {
		return apperror.InternalError("Failed to generate tokens", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "token",
				Reason: "Token generation failed",
			})
	}

	// Set access token cookie
	accessCookie := new(http.Cookie)
	accessCookie.Name = "access_token"
	accessCookie.Value = tokenPair.AccessToken
	accessCookie.Path = "/"
	accessCookie.MaxAge = tokenPair.ExpiresIn
	accessCookie.HttpOnly = true
	accessCookie.Secure = os.Getenv("APP_ENV") == "production"
	accessCookie.SameSite = http.SameSiteStrictMode
	c.SetCookie(accessCookie)

	// Set refresh token cookie
	newRefreshCookie := new(http.Cookie)
	newRefreshCookie.Name = "refresh_token"
	newRefreshCookie.Value = tokenPair.RefreshToken
	newRefreshCookie.Path = "/"
	newRefreshCookie.MaxAge = 7 * 24 * 60 * 60 // 7 days
	newRefreshCookie.HttpOnly = true
	newRefreshCookie.Secure = os.Getenv("APP_ENV") == "production"
	newRefreshCookie.SameSite = http.SameSiteStrictMode
	c.SetCookie(newRefreshCookie)

	// Include tokens in response for backward compatibility
	return c.JSON(http.StatusOK, map[string]interface{}{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    tokenPair.ExpiresIn,
	})
}

// Logout clears authentication cookies
func (h *AuthHandler) Logout(c echo.Context) error {
	// Clear access token cookie
	accessCookie := new(http.Cookie)
	accessCookie.Name = "access_token"
	accessCookie.Value = ""
	accessCookie.Path = "/"
	accessCookie.MaxAge = -1 // Delete the cookie
	accessCookie.HttpOnly = true
	accessCookie.Secure = os.Getenv("APP_ENV") == "production"
	accessCookie.SameSite = http.SameSiteStrictMode
	c.SetCookie(accessCookie)

	// Clear refresh token cookie
	refreshCookie := new(http.Cookie)
	refreshCookie.Name = "refresh_token"
	refreshCookie.Value = ""
	refreshCookie.Path = "/"
	refreshCookie.MaxAge = -1 // Delete the cookie
	refreshCookie.HttpOnly = true
	refreshCookie.Secure = os.Getenv("APP_ENV") == "production"
	refreshCookie.SameSite = http.SameSiteStrictMode
	c.SetCookie(refreshCookie)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
	})
}
