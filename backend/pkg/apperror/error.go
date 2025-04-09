package apperror

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// Error types
	ErrorTypeValidation    ErrorType = "validation"
	ErrorTypeAuthorization ErrorType = "authorization"
	ErrorTypeNotFound      ErrorType = "not_found"
	ErrorTypeDatabase      ErrorType = "database"
	ErrorTypeInternal      ErrorType = "internal"
	ErrorTypeExternal      ErrorType = "external"
)

// ErrorDetails interface for type-safe error details
type ErrorDetails interface {
	Validate() error
	ToClientSafe() map[string]interface{}
}

// ValidationErrorDetails represents validation error details
type ValidationErrorDetails struct {
	Field  string `json:"field"`
	Value  string `json:"value,omitempty"`
	Reason string `json:"reason"`
}

func (d *ValidationErrorDetails) Validate() error {
	if d.Field == "" || d.Reason == "" {
		return fmt.Errorf("field and reason are required for validation errors")
	}
	return nil
}

func (d *ValidationErrorDetails) ToClientSafe() map[string]interface{} {
	return map[string]interface{}{
		"field":  d.Field,
		"reason": d.Reason,
	}
}

// DatabaseErrorDetails represents database error details
type DatabaseErrorDetails struct {
	Operation string `json:"operation"`
	Table     string `json:"table"`
	Query     string `json:"query,omitempty"`
}

func (d *DatabaseErrorDetails) Validate() error {
	if d.Operation == "" || d.Table == "" {
		return fmt.Errorf("operation and table are required for database errors")
	}
	return nil
}

func (d *DatabaseErrorDetails) ToClientSafe() map[string]interface{} {
	return map[string]interface{}{
		"operation": d.Operation,
		"table":     d.Table,
	}
}

// ExternalServiceErrorDetails represents external service error details
type ExternalServiceErrorDetails struct {
	Service   string `json:"service"`
	Operation string `json:"operation"`
	Endpoint  string `json:"endpoint,omitempty"`
}

func (d *ExternalServiceErrorDetails) Validate() error {
	if d.Service == "" || d.Operation == "" {
		return fmt.Errorf("service and operation are required for external service errors")
	}
	return nil
}

func (d *ExternalServiceErrorDetails) ToClientSafe() map[string]interface{} {
	return map[string]interface{}{
		"service":   d.Service,
		"operation": d.Operation,
	}
}

// AuthorizationErrorDetails represents authorization error details
type AuthorizationErrorDetails struct {
	Reason string `json:"reason"`
	Field  string `json:"field,omitempty"`
}

func (d *AuthorizationErrorDetails) Validate() error {
	if d.Reason == "" {
		return fmt.Errorf("reason is required for authorization errors")
	}
	return nil
}

func (d *AuthorizationErrorDetails) ToClientSafe() map[string]interface{} {
	result := map[string]interface{}{
		"reason": d.Reason,
	}
	if d.Field != "" {
		result["field"] = d.Field
	}
	return result
}

// ErrorContext provides additional context about where and when an error occurred
type ErrorContext struct {
	Function  string    `json:"function"`
	File      string    `json:"file"`
	Line      int       `json:"line"`
	Timestamp time.Time `json:"timestamp"`
	Operation string    `json:"operation,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
	UserID    string    `json:"user_id,omitempty"`
}

// AppError represents an application error
type AppError struct {
	Type       ErrorType    `json:"type"`
	Code       string       `json:"code"`
	Message    string       `json:"message"`
	StatusCode int          `json:"status_code"`
	Err        error        `json:"-"`
	Source     string       `json:"-"`
	Details    ErrorDetails `json:"details,omitempty"`
	Context    ErrorContext `json:"-"`
	Wrapped    error        `json:"-"`
}

// Error returns the error message
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Wrapped
}

// Is checks if the error is of a specific type
func (e *AppError) Is(target error) bool {
	if target == nil {
		return false
	}
	if err, ok := target.(*AppError); ok {
		return e.Type == err.Type && e.Code == err.Code
	}
	return false
}

// WithContext adds context to the error
func (e *AppError) WithContext(ctx ErrorContext) *AppError {
	e.Context = ctx
	return e
}

// WithDetails adds type-safe details to the error
func (e *AppError) WithDetails(details ErrorDetails) *AppError {
	if err := details.Validate(); err != nil {
		log.Printf("Invalid error details: %v", err)
		return e
	}
	e.Details = details
	return e
}

// Wrap wraps another error
func (e *AppError) Wrap(err error) *AppError {
	e.Wrapped = err
	return e
}

// NewError creates a new application error
func NewError(errType ErrorType, code string, message string, statusCode int, err error) *AppError {
	pc, file, line, _ := runtime.Caller(1)
	function := runtime.FuncForPC(pc).Name()

	return &AppError{
		Type:       errType,
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
		Context: ErrorContext{
			Function:  function,
			File:      file,
			Line:      line,
			Timestamp: time.Now(),
		},
	}
}

// LogError logs the error with structured logging
func (e *AppError) LogError() {
	logData := map[string]interface{}{
		"type":        e.Type,
		"code":        e.Code,
		"message":     e.Message,
		"status_code": e.StatusCode,
		"context":     e.Context,
	}

	if e.Details != nil {
		logData["details"] = e.Details
	}

	if e.Err != nil {
		logData["error"] = e.Err.Error()
	}

	if e.Wrapped != nil {
		logData["wrapped_error"] = e.Wrapped.Error()
	}

	jsonData, err := json.Marshal(logData)
	if err != nil {
		log.Printf("Failed to marshal error log: %v", err)
		return
	}

	log.Printf("Error: %s", string(jsonData))
}

// ToHTTPResponse converts the error to a client-friendly HTTP response format
func (e *AppError) ToHTTPResponse() map[string]interface{} {
	response := map[string]interface{}{
		"type":    e.Type,
		"code":    e.Code,
		"message": e.Message,
	}

	if e.Details != nil {
		response["details"] = e.Details.ToClientSafe()
	}

	return response
}

// Helper functions for creating specific error types
func ValidationError(message string, err error) *AppError {
	return NewError(ErrorTypeValidation, "validation_error", message, http.StatusBadRequest, err)
}

func AuthorizationError(message string, err error) *AppError {
	return NewError(ErrorTypeAuthorization, "authorization_error", message, http.StatusUnauthorized, err)
}

func NotFoundError(message string, err error) *AppError {
	return NewError(ErrorTypeNotFound, "not_found_error", message, http.StatusNotFound, err)
}

func DatabaseError(message string, err error) *AppError {
	return NewError(ErrorTypeDatabase, "database_error", message, http.StatusInternalServerError, err)
}

func InternalError(message string, err error) *AppError {
	return NewError(ErrorTypeInternal, "internal_error", message, http.StatusInternalServerError, err)
}

func ExternalServiceError(message string, err error) *AppError {
	return NewError(ErrorTypeExternal, "external_service_error", message, http.StatusBadGateway, err)
}

// ErrorResponse converts various error types to a consistent Echo HTTP error
func ErrorResponse(err error) *echo.HTTPError {
	switch e := err.(type) {
	case *AppError:
		// Log the error with full details
		e.LogError()

		// Return a structured error response
		response := e.ToHTTPResponse()
		return echo.NewHTTPError(e.StatusCode, response)

	case *echo.HTTPError:
		// Already an Echo HTTP error
		return e

	default:
		// Unknown error type, convert to internal error
		appErr := InternalError("An unexpected error occurred", err)
		appErr.LogError()
		return echo.NewHTTPError(appErr.StatusCode, appErr.Message)
	}
}
