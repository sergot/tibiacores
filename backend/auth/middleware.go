package auth

import (
	"strings"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")

		if authHeader == "" {
			return next(c)
		}

		// Validate JWT token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ValidateToken(tokenString)

		if err != nil {
			return next(c)
		}

		// Set authenticated user information in context
		c.Set("user_id", claims.UserID)
		c.Set("has_email", claims.HasEmail)
		return next(c)
	}
}
