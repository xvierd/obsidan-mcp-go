package obsidian

import (
	"errors"
	"fmt"
)

// Common errors returned by the Obsidian API client.
var (
	ErrNoteNotFound     = errors.New("note not found")
	ErrUnauthorized     = errors.New("unauthorized - check API key")
	ErrForbidden        = errors.New("forbidden")
	ErrConnectionFailed = errors.New("connection failed")
	ErrTimeout          = errors.New("request timeout")
	ErrRateLimited      = errors.New("rate limited")
	ErrInvalidRequest   = errors.New("invalid request")
	ErrServerError      = errors.New("server error")
)

// APIError represents an error returned by the Obsidian REST API.
type APIError struct {
	StatusCode int
	Message    string
	Err        error
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("obsidian API error (status %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("obsidian API error (status %d): %v", e.StatusCode, e.Err)
}

// Unwrap returns the underlying error.
func (e *APIError) Unwrap() error {
	return e.Err
}

// IsNotFound returns true if the error is a not found error.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return errors.Is(err, ErrNoteNotFound)
}

// IsUnauthorized returns true if the error is an unauthorized error.
func IsUnauthorized(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401
	}
	return errors.Is(err, ErrUnauthorized)
}

// NewAPIError creates a new APIError from status code and message.
func NewAPIError(statusCode int, message string) error {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// MapHTTPStatus maps HTTP status codes to appropriate errors.
func MapHTTPStatus(statusCode int, message string) error {
	switch statusCode {
	case 200, 201, 204:
		return nil
	case 400:
		return &APIError{StatusCode: statusCode, Message: message, Err: ErrInvalidRequest}
	case 401:
		return &APIError{StatusCode: statusCode, Message: message, Err: ErrUnauthorized}
	case 403:
		return &APIError{StatusCode: statusCode, Message: message, Err: ErrForbidden}
	case 404:
		return &APIError{StatusCode: statusCode, Message: message, Err: ErrNoteNotFound}
	case 429:
		return &APIError{StatusCode: statusCode, Message: message, Err: ErrRateLimited}
	case 500, 502, 503, 504:
		return &APIError{StatusCode: statusCode, Message: message, Err: ErrServerError}
	default:
		return &APIError{StatusCode: statusCode, Message: message}
	}
}
