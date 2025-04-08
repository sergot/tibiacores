package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Error types
var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrNotFound       = errors.New("not found")
	ErrConflict       = errors.New("conflict")
	ErrInternal       = errors.New("internal server error")
	ErrValidation     = errors.New("validation error")
	ErrDatabase       = errors.New("database error")
	ErrEmailService   = errors.New("email service error")
	// Creature-specific errors
	ErrCreatureNotFound    = errors.New("creature not found")
	ErrCreatureInvalid     = errors.New("invalid creature data")
	ErrCreatureDuplicate   = errors.New("duplicate creature")
	ErrCreatureInUse       = errors.New("creature is in use")
	ErrCreatureUnavailable = errors.New("creature unavailable")
	// OAuth-specific errors
	ErrOAuthProviderInvalid = errors.New("invalid OAuth provider")
	ErrOAuthStateInvalid    = errors.New("invalid OAuth state")
	ErrOAuthCodeInvalid     = errors.New("invalid OAuth code")
	ErrOAuthTokenInvalid    = errors.New("invalid OAuth token")
	ErrOAuthEmailInUse      = errors.New("email already in use with different account type")
	// List-specific errors
	ErrListNotFound    = errors.New("list not found")
	ErrListInvalid     = errors.New("invalid list data")
	ErrListDuplicate   = errors.New("duplicate list")
	ErrListInUse       = errors.New("list is in use")
	ErrListUnavailable = errors.New("list unavailable")
	ErrListForbidden   = errors.New("list access forbidden")
)

// ValidationError represents a field-level validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// APIError represents an error response from the API
type APIError struct {
	Code             int               `json:"code"`                        // HTTP status code
	Message          string            `json:"message"`                     // User-friendly message
	ErrorType        string            `json:"error"`                       // Error type/code
	DebugInfo        string            `json:"-"`                           // Debug info (not sent to client)
	InternalErr      error             `json:"-"`                           // Original error (not sent to client)
	ValidationErrors []ValidationError `json:"validation_errors,omitempty"` // Field-level validation errors
	Operation        string            `json:"-"`                           // Operation that failed
	Resource         string            `json:"-"`                           // Resource that was affected
}

// NewAPIError creates a new APIError
func NewAPIError(code int, message string, errorType string, debugInfo string, err error) *APIError {
	return &APIError{
		Code:        code,
		Message:     message,
		ErrorType:   errorType,
		DebugInfo:   debugInfo,
		InternalErr: err,
	}
}

// WithOperation adds operation context to the error
func (e *APIError) WithOperation(op string) *APIError {
	e.Operation = op
	return e
}

// WithResource adds resource context to the error
func (e *APIError) WithResource(resource string) *APIError {
	e.Resource = resource
	return e
}

// WithValidation adds validation errors to the error
func (e *APIError) WithValidation(errors []ValidationError) *APIError {
	e.ValidationErrors = errors
	return e
}

// Error implements the error interface
func (e *APIError) Error() string {
	msg := e.Message
	if e.Operation != "" {
		msg = fmt.Sprintf("%s (operation: %s)", msg, e.Operation)
	}
	if e.Resource != "" {
		msg = fmt.Sprintf("%s (resource: %s)", msg, e.Resource)
	}
	if e.InternalErr != nil {
		msg = fmt.Sprintf("%s: %v (debug: %s)", msg, e.InternalErr, e.DebugInfo)
	}
	return msg
}

// Unwrap implements the error unwrapping interface
func (e *APIError) Unwrap() error {
	return e.InternalErr
}

// HTTPStatus returns the HTTP status code for the error
func (e *APIError) HTTPStatus() int {
	return e.Code
}

// Helper functions for common error types
func NewInvalidRequestError(err error) *APIError {
	return NewAPIError(
		http.StatusBadRequest,
		"Invalid request",
		"invalid_request",
		fmt.Sprintf("Request validation failed: %v", err),
		err,
	)
}

func NewUnauthorizedError(err error) *APIError {
	return NewAPIError(
		http.StatusUnauthorized,
		"Unauthorized access",
		"unauthorized",
		fmt.Sprintf("Authentication failed: %v", err),
		err,
	)
}

func NewForbiddenError(err error) *APIError {
	return NewAPIError(
		http.StatusForbidden,
		"Access forbidden",
		"forbidden",
		fmt.Sprintf("Authorization failed: %v", err),
		err,
	)
}

func NewNotFoundError(err error) *APIError {
	return NewAPIError(
		http.StatusNotFound,
		"Resource not found",
		"not_found",
		fmt.Sprintf("Resource lookup failed: %v", err),
		err,
	)
}

func NewConflictError(err error) *APIError {
	return NewAPIError(
		http.StatusConflict,
		"Resource conflict",
		"conflict",
		fmt.Sprintf("Resource conflict: %v", err),
		err,
	)
}

func NewInternalError(err error) *APIError {
	return NewAPIError(
		http.StatusInternalServerError,
		"Internal server error",
		"internal_error",
		fmt.Sprintf("Internal error occurred: %v", err),
		err,
	)
}

func NewValidationError(err error) *APIError {
	return NewAPIError(
		http.StatusBadRequest,
		"Validation error",
		"validation_error",
		fmt.Sprintf("Validation failed: %v", err),
		err,
	)
}

func NewDatabaseError(err error) *APIError {
	return NewAPIError(
		http.StatusInternalServerError,
		"Database error",
		"database_error",
		fmt.Sprintf("Database operation failed: %v", err),
		err,
	)
}

func NewEmailServiceError(err error) *APIError {
	return NewAPIError(
		http.StatusInternalServerError,
		"Email service error",
		"email_service_error",
		fmt.Sprintf("Email service failed: %v", err),
		err,
	)
}

// Creature-specific error helpers
func NewCreatureNotFoundError(err error) *APIError {
	return NewAPIError(
		http.StatusNotFound,
		"Creature not found",
		"creature_not_found",
		fmt.Sprintf("Creature lookup failed: %v", err),
		err,
	)
}

func NewCreatureInvalidError(err error) *APIError {
	return NewAPIError(
		http.StatusBadRequest,
		"Invalid creature data",
		"creature_invalid",
		fmt.Sprintf("Creature validation failed: %v", err),
		err,
	)
}

func NewCreatureDuplicateError(err error) *APIError {
	return NewAPIError(
		http.StatusConflict,
		"Creature already exists",
		"creature_duplicate",
		fmt.Sprintf("Creature creation failed: %v", err),
		err,
	)
}

func NewCreatureInUseError(err error) *APIError {
	return NewAPIError(
		http.StatusConflict,
		"Creature is in use",
		"creature_in_use",
		fmt.Sprintf("Creature operation failed: %v", err),
		err,
	)
}

func NewCreatureUnavailableError(err error) *APIError {
	return NewAPIError(
		http.StatusServiceUnavailable,
		"Creature service unavailable",
		"creature_unavailable",
		fmt.Sprintf("Creature service error: %v", err),
		err,
	)
}

// OAuth-specific error helpers
func NewOAuthProviderError(err error) *APIError {
	return NewAPIError(
		http.StatusBadRequest,
		"Invalid OAuth provider",
		"oauth_provider_invalid",
		fmt.Sprintf("OAuth provider error: %v", err),
		err,
	)
}

func NewOAuthStateError(err error) *APIError {
	return NewAPIError(
		http.StatusBadRequest,
		"Invalid OAuth state",
		"oauth_state_invalid",
		fmt.Sprintf("OAuth state validation failed: %v", err),
		err,
	)
}

func NewOAuthCodeError(err error) *APIError {
	return NewAPIError(
		http.StatusUnauthorized,
		"Invalid OAuth code",
		"oauth_code_invalid",
		fmt.Sprintf("OAuth code exchange failed: %v", err),
		err,
	)
}

func NewOAuthTokenError(err error) *APIError {
	return NewAPIError(
		http.StatusUnauthorized,
		"Invalid OAuth token",
		"oauth_token_invalid",
		fmt.Sprintf("OAuth token validation failed: %v", err),
		err,
	)
}

func NewOAuthEmailInUseError(err error) *APIError {
	return NewAPIError(
		http.StatusConflict,
		"Email already in use with different account type",
		"oauth_email_in_use",
		fmt.Sprintf("OAuth email conflict: %v", err),
		err,
	)
}

// List-specific error helpers
func NewListNotFoundError(err error) *APIError {
	return NewAPIError(
		http.StatusNotFound,
		"List not found",
		"list_not_found",
		fmt.Sprintf("List lookup failed: %v", err),
		err,
	)
}

func NewListInvalidError(err error) *APIError {
	return NewAPIError(
		http.StatusBadRequest,
		"Invalid list data",
		"list_invalid",
		fmt.Sprintf("List validation failed: %v", err),
		err,
	)
}

func NewListDuplicateError(err error) *APIError {
	return NewAPIError(
		http.StatusConflict,
		"List already exists",
		"list_duplicate",
		fmt.Sprintf("List creation failed: %v", err),
		err,
	)
}

func NewListInUseError(err error) *APIError {
	return NewAPIError(
		http.StatusConflict,
		"List is in use",
		"list_in_use",
		fmt.Sprintf("List operation failed: %v", err),
		err,
	)
}

func NewListUnavailableError(err error) *APIError {
	return NewAPIError(
		http.StatusServiceUnavailable,
		"List service unavailable",
		"list_unavailable",
		fmt.Sprintf("List service error: %v", err),
		err,
	)
}

func NewListForbiddenError(err error) *APIError {
	return NewAPIError(
		http.StatusForbidden,
		"Access to list forbidden",
		"list_forbidden",
		fmt.Sprintf("List access denied: %v", err),
		err,
	)
}

// IsAPIError checks if an error is an APIError
func IsAPIError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr)
}

// GetAPIError returns the APIError if the error is one, or nil otherwise
func GetAPIError(err error) *APIError {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return nil
}
