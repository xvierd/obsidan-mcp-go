package domain

import (
	"errors"
	"fmt"
)

// DomainError represents an error that occurred in the domain layer.
type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// Error implements the error interface.
func (e *DomainError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Code, e.Err)
	}
	return e.Code
}

// Unwrap returns the underlying error.
func (e *DomainError) Unwrap() error {
	return e.Err
}

// IsNotFound returns true if the error is a not found error.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var domainErr *DomainError
	if As(err, &domainErr) {
		return domainErr.Code == "NOT_FOUND"
	}
	return err == ErrNoteNotFound
}

// IsUnauthorized returns true if the error is an unauthorized error.
func IsUnauthorized(err error) bool {
	if err == nil {
		return false
	}
	var domainErr *DomainError
	if As(err, &domainErr) {
		return domainErr.Code == "UNAUTHORIZED"
	}
	return err == ErrUnauthorized
}

// IsForbidden returns true if the error is a forbidden error.
func IsForbidden(err error) bool {
	if err == nil {
		return false
	}
	var domainErr *DomainError
	if As(err, &domainErr) {
		return domainErr.Code == "FORBIDDEN"
	}
	return err == ErrForbidden
}

// IsTimeout returns true if the error is a timeout error.
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	var domainErr *DomainError
	if As(err, &domainErr) {
		return domainErr.Code == "TIMEOUT"
	}
	return err == ErrTimeout
}

// IsRateLimited returns true if the error is a rate limited error.
func IsRateLimited(err error) bool {
	if err == nil {
		return false
	}
	var domainErr *DomainError
	if As(err, &domainErr) {
		return domainErr.Code == "RATE_LIMITED"
	}
	return err == ErrRateLimited
}

// IsValidation returns true if the error is a validation error.
func IsValidation(err error) bool {
	if err == nil {
		return false
	}
	var domainErr *DomainError
	if As(err, &domainErr) {
		return domainErr.Code == "VALIDATION"
	}
	return err == ErrValidation
}

// IsConnectionFailed returns true if the error is a connection failed error.
func IsConnectionFailed(err error) bool {
	if err == nil {
		return false
	}
	var domainErr *DomainError
	if As(err, &domainErr) {
		return domainErr.Code == "CONNECTION_FAILED"
	}
	return err == ErrConnectionFailed
}

// IsInvalidRequest returns true if the error is an invalid request error.
func IsInvalidRequest(err error) bool {
	if err == nil {
		return false
	}
	var domainErr *DomainError
	if As(err, &domainErr) {
		return domainErr.Code == "INVALID_REQUEST"
	}
	return err == ErrInvalidRequest
}

// IsServerError returns true if the error is a server error.
func IsServerError(err error) bool {
	if err == nil {
		return false
	}
	var domainErr *DomainError
	if As(err, &domainErr) {
		return domainErr.Code == "SERVER_ERROR"
	}
	return err == ErrServerError
}

// NewDomainError creates a new DomainError.
func NewDomainError(code, message string) error {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

// NewDomainErrorWrap wraps an existing error as a DomainError.
func NewDomainErrorWrap(code string, err error) error {
	return &DomainError{
		Code: code,
		Err:  err,
	}
}

// As is a wrapper for errors.As.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
