package middleware

import (
	"github.com/ant0ine/go-json-rest/rest"
)

const (
	// SpHeader special header to indicate our microservice
	SpHeader = "X-Service-Proxy"
	// SpHeaderVal special header to indicate our microservice
	SpHeaderVal = "service_proxy"
)

// ServiceProxyMiddleware appending Service Proxy service header to
type ServiceProxyMiddleware struct{}

// MiddlewareFunc help us uniquely identify that requests made it to our microservices
// vs getting caught with a 404 from gorouter or something
func (mw *ServiceProxyMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	return func(w rest.ResponseWriter, r *rest.Request) {
		w.Header().Set(SpHeader, SpHeaderVal)
		h(w, r)
	}
}
