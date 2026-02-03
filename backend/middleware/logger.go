package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

// SlogLogger is a middleware that logs requests using log/slog
func SlogLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			latency := time.Since(start)

			req := c.Request()
			res := c.Response()

			// Determine log level based on status code
			level := slog.LevelInfo
			msg := "request completed"
			if res.Status >= 500 {
				level = slog.LevelError
				msg = "server error"
			} else if res.Status >= 400 {
				level = slog.LevelWarn
				msg = "client error"
			}

			// Add error to attributes if present
			attrs := []any{
				slog.String("method", req.Method),
				slog.String("path", req.URL.Path),
				slog.Int("status", res.Status),
				slog.Duration("latency", latency),
				slog.String("ip", c.RealIP()),
			}

			if err != nil {
				attrs = append(attrs, slog.String("error", err.Error()))
			}

			logger.Log(c.Request().Context(), level, msg, attrs...)

			return err
		}
	}
}
