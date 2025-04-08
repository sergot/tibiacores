package errors

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorHandler is a middleware that handles errors consistently
func ErrorHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if err == nil {
			return nil
		}

		// Get request context
		req := c.Request()
		path := req.URL.Path
		method := req.Method
		ip := c.RealIP()
		userAgent := req.UserAgent()
		requestID := c.Response().Header().Get(echo.HeaderXRequestID)

		// Check if it's an APIError
		if apiErr := GetAPIError(err); apiErr != nil {
			// Log the error with debug information
			log.Printf("[ERROR] %s %s %s %s %s - %s (debug: %s, internal: %v)",
				requestID,
				ip,
				method,
				path,
				userAgent,
				apiErr.Message,
				apiErr.DebugInfo,
				apiErr.InternalErr,
			)
			return c.JSON(apiErr.HTTPStatus(), apiErr)
		}

		// Check if it's an echo.HTTPError
		if httpErr, ok := err.(*echo.HTTPError); ok {
			// Log the error
			log.Printf("[ERROR] %s %s %s %s %s - %v",
				requestID,
				ip,
				method,
				path,
				userAgent,
				httpErr.Message,
			)
			return c.JSON(httpErr.Code, NewAPIError(httpErr.Code, httpErr.Message.(string), "http_error", "", err))
		}

		// Default to internal server error
		log.Printf("[ERROR] %s %s %s %s %s - Unhandled error: %v",
			requestID,
			ip,
			method,
			path,
			userAgent,
			err,
		)
		return c.JSON(http.StatusInternalServerError, NewInternalError(err))
	}
}
