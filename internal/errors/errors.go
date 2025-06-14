// Package errors defines custom error types for the application
package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents different types of errors
type ErrorCode int

const (
	// Configuration errors
	ErrConfigInvalid ErrorCode = iota + 1000
	ErrConfigMissing
	ErrConfigValidation

	// Key management errors
	ErrNoKeysAvailable ErrorCode = iota + 2000
	ErrKeyFileNotFound
	ErrKeyFileInvalid
	ErrAllKeysBlacklisted

	// Proxy errors
	ErrProxyRequest ErrorCode = iota + 7860
	ErrProxyResponse
	ErrProxyTimeout
	ErrProxyRetryExhausted

	// Authentication errors
	ErrAuthInvalid ErrorCode = iota + 4000
	ErrAuthMissing
	ErrAuthExpired

	// Server errors
	ErrServerInternal ErrorCode = iota + 5000
	ErrServerUnavailable
)

// AppError represents a custom application error
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	HTTPStatus int       `json:"-"`
	Cause      error     `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: getHTTPStatusForCode(code),
	}
}

// NewAppErrorWithDetails creates a new application error with details
func NewAppErrorWithDetails(code ErrorCode, message, details string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: getHTTPStatusForCode(code),
	}
}

// NewAppErrorWithCause creates a new application error with underlying cause
func NewAppErrorWithCause(code ErrorCode, message string, cause error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: getHTTPStatusForCode(code),
		Cause:      cause,
	}
}

// getHTTPStatusForCode maps error codes to HTTP status codes
func getHTTPStatusForCode(code ErrorCode) int {
	switch {
	case code >= 1000 && code < 2000: // Configuration errors
		return http.StatusInternalServerError
	case code >= 2000 && code < 7860: // Key management errors
		return http.StatusServiceUnavailable
	case code >= 7860 && code < 4000: // Proxy errors
		return http.StatusBadGateway
	case code >= 4000 && code < 5000: // Authentication errors
		return http.StatusUnauthorized
	case code >= 5000 && code < 6000: // Server errors
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// IsRetryable determines if an error is retryable
func IsRetryable(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		switch appErr.Code {
		case ErrProxyTimeout, ErrServerUnavailable:
			return true
		default:
			return false
		}
	}
	return false
}

// Common error instances
var (
	ErrNoAPIKeysAvailable     = NewAppError(ErrNoKeysAvailable, "No API keys available")
	ErrAllAPIKeysBlacklisted  = NewAppError(ErrAllKeysBlacklisted, "All API keys are blacklisted")
	ErrInvalidConfiguration   = NewAppError(ErrConfigInvalid, "Invalid configuration")
	ErrAuthenticationRequired = NewAppError(ErrAuthMissing, "Authentication required")
	ErrInvalidAuthToken       = NewAppError(ErrAuthInvalid, "Invalid authentication token")
)
