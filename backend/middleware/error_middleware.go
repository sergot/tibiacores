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

	var appErr *apperror.AppError
	var statusCode int

	switch e := err.(type) {
	case *apperror.AppError:
		// Add request context to the error
		appErr = e.WithContext(ctx)
		statusCode = e.StatusCode

	case *echo.HTTPError:
		// Convert Echo HTTP error to our AppError format
		appErr = apperror.NewError(
			apperror.ErrorTypeInternal,
			"http_error",
			e.Message.(string),
			e.Code,
			err,
		).WithContext(ctx)
		statusCode = e.Code

	default:
		// Create a new internal error for unknown error types
		appErr = apperror.NewError(
			apperror.ErrorTypeInternal,
			"internal_error",
			"An unexpected error occurred",
			http.StatusInternalServerError,
			err,
		).WithContext(ctx)
		statusCode = http.StatusInternalServerError
	}

	// Log internal error details
	appErr.LogError()

	// Return client-safe error response
	response := appErr.ToHTTPResponse()
	if jsonErr := c.JSON(statusCode, response); jsonErr != nil {
		// If we fail to send JSON, create a new error with the JSON error
		jsonAppErr := apperror.NewError(
			apperror.ErrorTypeInternal,
			"json_serialization_error",
			"Failed to serialize error response",
			statusCode,
			jsonErr,
		).WithContext(ctx)

		// Log the JSON error
		jsonAppErr.LogError()

		// Try to send a simple text response as a last resort
		c.String(statusCode, "Internal Server Error")
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

					// Create error context
					ctx := apperror.ErrorContext{
						RequestID: requestID,
						UserID:    userID,
						Operation: c.Request().Method + " " + c.Request().URL.Path,
					}

					// Create a new internal error for the panic
					appErr := apperror.NewError(
						apperror.ErrorTypeInternal,
						"panic",
						"An unexpected error occurred",
						http.StatusInternalServerError,
						err,
					).WithContext(ctx)

					// Log the panic and stack trace
					appErr.LogError()
					c.Logger().Errorf("Panic recovered: %v\n%s", err, debug.Stack())

					// Return client-safe error response
					response := appErr.ToHTTPResponse()
					if jsonErr := c.JSON(http.StatusInternalServerError, response); jsonErr != nil {
						// If we fail to send JSON, create a new error with the JSON error
						jsonAppErr := apperror.NewError(
							apperror.ErrorTypeInternal,
							"json_serialization_error",
							"Failed to serialize error response",
							http.StatusInternalServerError,
							jsonErr,
						).WithContext(ctx)

						// Log the JSON error
						jsonAppErr.LogError()

						// Try to send a simple text response as a last resort
						c.String(http.StatusInternalServerError, "Internal Server Error")
					}
				}
			}()

			return next(c)
		}
	}
}
