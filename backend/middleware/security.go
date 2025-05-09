package middleware

import (
	"os"

	"github.com/labstack/echo/v4"
)

// SecurityHeaders adds security-related HTTP headers to all responses
func SecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Prevent browsers from MIME-sniffing
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")

			// Prevent embedding in iframes (clickjacking protection)
			c.Response().Header().Set("X-Frame-Options", "DENY")

			// HSTS (only in production)
			if os.Getenv("APP_ENV") == "production" {
				c.Response().Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			// Content Security Policy
			c.Response().Header().Set("Content-Security-Policy",
				"default-src 'self'; "+
					"img-src 'self' data: https:; "+
					"style-src 'self' 'unsafe-inline'; "+ // Required for Vue's dynamic styles
					"script-src 'self' https://umami.tibiacores.com; "+ // Allow Umami analytics
					"connect-src 'self' https://api.tibiadata.com; "+ // Allow TibiaData API
					"frame-ancestors 'none'")

			return next(c)
		}
	}
}
