package auth

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
)

// getTokenFromRequest gets the access token from cookie
func getTokenFromRequest(c echo.Context) string {
	cookie, err := c.Cookie("access_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}
	return ""
}

// SetTokenCookies sets access and refresh token cookies
func SetTokenCookies(c echo.Context, accessToken string, refreshToken string, maxAge int) {
	// Set access token cookie
	accessCookie := new(http.Cookie)
	accessCookie.Name = "access_token"
	accessCookie.Value = accessToken
	accessCookie.Path = "/"
	accessCookie.MaxAge = maxAge
	accessCookie.HttpOnly = true
	accessCookie.Secure = os.Getenv("APP_ENV") == "production"
	accessCookie.SameSite = http.SameSiteStrictMode
	c.SetCookie(accessCookie)

	// Set refresh token cookie
	refreshCookie := new(http.Cookie)
	refreshCookie.Name = "refresh_token"
	refreshCookie.Value = refreshToken
	refreshCookie.Path = "/"
	refreshCookie.MaxAge = 7 * 24 * 60 * 60 // 7 days in seconds
	refreshCookie.HttpOnly = true
	refreshCookie.Secure = os.Getenv("APP_ENV") == "production"
	refreshCookie.SameSite = http.SameSiteStrictMode
	c.SetCookie(refreshCookie)
}

// AuthMiddleware requires a valid auth token and sets user info in context
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := getTokenFromRequest(c)

		if tokenString == "" {
			return apperror.AuthorizationError("Missing authentication cookie", nil).
				WithDetails(&apperror.AuthorizationErrorDetails{
					Reason: "missing_auth_cookie",
					Field:  "access_token",
				})
		}

		// Validate JWT token
		claims, err := ValidateAccessToken(tokenString)

		if err != nil {
			return apperror.AuthorizationError("Invalid or expired token", err).
				WithDetails(&apperror.AuthorizationErrorDetails{
					Reason: "token_validation_failed",
					Field:  "access_token",
				})
		}

		// Set authenticated user information in context
		c.Set("user_id", claims.UserID)
		c.Set("has_email", claims.HasEmail)
		return next(c)
	}
}

// OptionalAuthMiddleware tries to validate auth token but allows requests through without it
func OptionalAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := getTokenFromRequest(c)

		if tokenString != "" {
			// Try to validate JWT token
			if claims, err := ValidateAccessToken(tokenString); err == nil {
				// Set authenticated user information in context
				c.Set("user_id", claims.UserID)
				c.Set("has_email", claims.HasEmail)
			}
		}

		return next(c)
	}
}
