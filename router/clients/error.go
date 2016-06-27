package clients

import (
	"fmt"
	"net/http"
)

// ConflictError encompass errors involving conflicts (for example attempting a create on a pre-existing resource)
type ConflictError struct{}

// Error description
func (e *ConflictError) Error() string {
	return "client: conflict performing operation"
}

// TenantNotFoundError is returned when a tenant was not found
type TenantNotFoundError struct{}

// Error description
func (e *TenantNotFoundError) Error() string {
	return "client: tenant not found"
}

// ServiceUnavailable indicates that the endpoint has reported that it is unable to service the request
type ServiceUnavailable struct{}

// Error description
func (e *ServiceUnavailable) Error() string {
	return "client: service temporarily unavailable"
}

// NetworkError encompasses errors originating from sources other than the controller (I.E., the request never made it
// to the controller)
type NetworkError struct {
	Response *http.Response
}

// Error description
func (e *NetworkError) Error() string {
	return fmt.Sprintf("client: %v", e.Response.StatusCode)
}

// ConnectionError TODO
type ConnectionError struct {
	Message string
}

// Error description
func (e *ConnectionError) Error() string {
	return e.Message
}
