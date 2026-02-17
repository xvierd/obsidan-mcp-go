package domain

import (
	"errors"
	"fmt"
)

// Error codes used in DomainError.
const (
	CodeNotFound         = "NOT_FOUND"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeTimeout          = "TIMEOUT"
	CodeRateLimited      = "RATE_LIMITED"
	CodeValidation       = "VALIDATION"
	CodeConnectionFailed = "CONNECTION_FAILED"
	CodeInvalidRequest   = "INVALID_REQUEST"
	CodeServerError      = "SERVER_ERROR"
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

// Common domain errors as DomainError values for consistent checking.
var (
	ErrNoteNotFound     = &DomainError{Code: CodeNotFound, Message: "note not found"}
	ErrUnauthorized     = &DomainError{Code: CodeUnauthorized, Message: "unauthorized"}
	ErrForbidden        = &DomainError{Code: CodeForbidden, Message: "forbidden"}
	ErrConnectionFailed = &DomainError{Code: CodeConnectionFailed, Message: "connection failed"}
	ErrTimeout          = &DomainError{Code: CodeTimeout, Message: "request timeout"}
	ErrRateLimited      = &DomainError{Code: CodeRateLimited, Message: "rate limited"}
	ErrInvalidRequest   = &DomainError{Code: CodeInvalidRequest, Message: "invalid request"}
	ErrServerError      = &DomainError{Code: CodeServerError, Message: "server error"}
	ErrValidation       = &DomainError{Code: CodeValidation, Message: "validation error"}
)

// isDomainCode checks if an error has the given domain error code.
func isDomainCode(err error, code string) bool {
	if err == nil {
		return false
	}
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code == code
	}
	return false
}

// IsNotFound returns true if the error is a not found error.
func IsNotFound(err error) bool { return isDomainCode(err, CodeNotFound) }

// IsUnauthorized returns true if the error is an unauthorized error.
func IsUnauthorized(err error) bool { return isDomainCode(err, CodeUnauthorized) }

// IsForbidden returns true if the error is a forbidden error.
func IsForbidden(err error) bool { return isDomainCode(err, CodeForbidden) }

// IsTimeout returns true if the error is a timeout error.
func IsTimeout(err error) bool { return isDomainCode(err, CodeTimeout) }

// IsRateLimited returns true if the error is a rate limited error.
func IsRateLimited(err error) bool { return isDomainCode(err, CodeRateLimited) }

// IsValidation returns true if the error is a validation error.
func IsValidation(err error) bool { return isDomainCode(err, CodeValidation) }

// IsConnectionFailed returns true if the error is a connection failed error.
func IsConnectionFailed(err error) bool { return isDomainCode(err, CodeConnectionFailed) }

// IsInvalidRequest returns true if the error is an invalid request error.
func IsInvalidRequest(err error) bool { return isDomainCode(err, CodeInvalidRequest) }

// IsServerError returns true if the error is a server error.
func IsServerError(err error) bool { return isDomainCode(err, CodeServerError) }

// NewDomainError creates a new DomainError with a code and message.
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
