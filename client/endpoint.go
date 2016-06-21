package client

import (
	"fmt"
	"net/url"
)

// ServiceEndpoint describes a network endpoint of service instance.
type ServiceEndpoint struct {

	// Type is the endpoint's type. Valid values are "http", "tcp", or "user".
	Type  string `json:"type"`

	// Value is the endpoint's value according to its type,
	// e.g. "172.135.10.1:8080" or "http://myapp.ng.bluemix.net/api/v1".
	Value string `json:"value"`
}

// NewHTTPEndpoint creates a new HTTP(S) network endpoint with the specified URL.
// The specified URL is assumed to have an "http" or "https" scheme.
func NewHTTPEndpoint(httpURL url.URL) ServiceEndpoint {
	return ServiceEndpoint{
		Type:  "http",
		Value: httpURL.String(),
	}
}

// NewTCPEndpoint creates a new TCP network endpoint with the specified host and optional port.
func NewTCPEndpoint(host string, port int) ServiceEndpoint {
	var value string
	if port > 0 {
		value = fmt.Sprintf("%s:%d", host, port)
	} else {
		value = host
	}
	return ServiceEndpoint{
		Type:  "tcp",
		Value: value,
	}
}

// NewCustomEndpoint creates a new network endpoint of a custom type.
// The value may be any arbitrary description of a network endpoint which makes sense in the context of the application.
func NewCustomEndpoint(endpoint string) ServiceEndpoint {
	return ServiceEndpoint{
		Type:  "user",
		Value: endpoint,
	}
}
