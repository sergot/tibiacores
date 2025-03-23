package auth

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// AuthMiddleware requires a valid auth token and sets user info in context
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")

		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
		}

		// Validate JWT token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ValidateToken(tokenString)

		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
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
		authHeader := c.Request().Header.Get("Authorization")

		if authHeader != "" {
			// Try to validate JWT token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if claims, err := ValidateToken(tokenString); err == nil {
				// Set authenticated user information in context
				c.Set("user_id", claims.UserID)
				c.Set("has_email", claims.HasEmail)
			}
		}

		return next(c)
	}
}
