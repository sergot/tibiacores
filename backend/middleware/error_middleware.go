package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/labstack/echo/v4"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
)

// ErrorHandler is a custom error handler that returns JSON responses
func ErrorHandler(err error, c echo.Context) {
	// Generate request ID if not present
	requestID := c.Request().Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = c.Response().Header().Get("X-Request-ID")
	}

	// Get user ID from context
	userID := ""
	if userIDVal := c.Get("user_id"); userIDVal != nil {
		if id, ok := userIDVal.(string); ok {
			userID = id
		}
	}

	// Add request context to error
	ctx := apperror.ErrorContext{
		RequestID: requestID,
		UserID:    userID,
		Operation: c.Request().Method + " " + c.Request().URL.Path,
	}

	switch e := err.(type) {
	case *apperror.AppError:
		// Add request context to the error
		e = e.WithContext(ctx)

		// Log internal error details
		e.LogError()

		// Return client-safe error response
		response := e.ToHTTPResponse()
		c.JSON(e.StatusCode, response)

	case *echo.HTTPError:
		// Convert Echo HTTP error to our AppError format
		appErr := apperror.NewError(
			apperror.ErrorTypeInternal,
			"http_error",
			e.Message.(string),
			e.Code,
			err,
		).WithContext(ctx)

		// Log internal error details
		appErr.LogError()

		// Return client-safe error response
		response := appErr.ToHTTPResponse()
		c.JSON(e.Code, response)

	default:
		// Create a new internal error for unknown error types
		appErr := apperror.NewError(
			apperror.ErrorTypeInternal,
			"internal_error",
			"An unexpected error occurred",
			http.StatusInternalServerError,
			err,
		).WithContext(ctx)

		// Log internal error details
		appErr.LogError()

		// Return client-safe error response
		response := appErr.ToHTTPResponse()
		c.JSON(http.StatusInternalServerError, response)
	}
}

// RecoverWithConfig returns a middleware that recovers from panics
func RecoverWithConfig() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}

					// Generate request ID if not present
					requestID := c.Request().Header.Get("X-Request-ID")
					if requestID == "" {
						requestID = c.Response().Header().Get("X-Request-ID")
					}

					// Get user ID from context
					userID := ""
					if userIDVal := c.Get("user_id"); userIDVal != nil {
						if id, ok := userIDVal.(string); ok {
							userID = id
						}
					}

					// Create error with context
					appErr := apperror.NewError(
						apperror.ErrorTypeInternal,
						"panic",
						"An unexpected error occurred",
						http.StatusInternalServerError,
						err,
					).WithContext(apperror.ErrorContext{
						RequestID: requestID,
						UserID:    userID,
						Operation: c.Request().Method + " " + c.Request().URL.Path,
					})

					// Log internal error details including stack trace
					appErr.LogError()
					debug.PrintStack()

					// Return client-safe error response
					response := appErr.ToHTTPResponse()
					c.JSON(http.StatusInternalServerError, response)
				}
			}()

			return next(c)
		}
	}
}
