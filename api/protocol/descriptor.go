package protocol

import (
	"github.com/ant0ine/go-json-rest/rest"
)

// APIDescriptor encapsulates information associated with an API call.
type APIDescriptor struct {

	// The path part of the URL.
	Path string

	// The HTTP method (e.g. "GET"). Valid values are defined in the MethodXXX constants in the http package.
	Method string

	// The protocol (e.g., Eureka).
	Protocol Type

	// The logic operation (e.g., RegisterInstance).
	Operation Operation

	// The handler registered for this API call.
	Handler rest.HandlerFunc
}

// AsRoute converts this APIDescriptor into a rest.Route object.
func (desc *APIDescriptor) AsRoute() *rest.Route {
	return &rest.Route{
		HttpMethod: desc.Method,
		PathExp:    desc.Path,
		Func:       desc.Handler,
	}
}
