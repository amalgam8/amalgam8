package clients

import (
	"fmt"
	"net/http"
)

type ConflictError struct{}

func (e *ConflictError) Error() string {
	return "Conflict performing operation"
}

type TenantNotFoundError struct{}

func (e *TenantNotFoundError) Error() string {
	return "Tenant not found"
}

type ServiceUnavailable struct{}

func (e *ServiceUnavailable) Error() string {
	return "Service temporarily unavailable"
}

type NetworkError struct {
	Response *http.Response
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error: %v", e.Response.StatusCode)
}

type ConnectionError struct {
	Message string
}

func (e *ConnectionError) Error() string {
	return e.Message
}
