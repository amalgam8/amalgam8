package store

import (
	"fmt"
)

// ErrorCode represents an error condition that might occur during a registry operation
type ErrorCode int

// ErrorCode predefined values
const (
	ErrorBadRequest ErrorCode = iota
	ErrorNoSuchServiceName
	ErrorNoSuchServiceInstance
	ErrorNamespaceQuotaExceeded
	ErrorInternalServerError
)

// Error is an error implementation that is associated with an ErrorCode
type Error struct {
	Code    ErrorCode
	Message string
	Cause   interface{}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d - %s (%v)", e.Code, e.Message, e.Cause)
}

// NewError creates a new registry.Error with the specified code, message and cause.
func NewError(code ErrorCode, message string, cause interface{}) *Error {
	return &Error{Code: code, Message: message, Cause: cause}
}
