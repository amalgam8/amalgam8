package middleware

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/pborman/uuid"
)

// RequestIDHeader constant for "request-id"
const RequestIDHeader = "request-id"

// RequestIDMiddleware appends a request ID header to incoming request for log tracing across services
type RequestIDMiddleware struct{}

// MiddlewareFunc makes RequestIDMiddleware implement the Middleware interface.
func (mw *RequestIDMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	return func(w rest.ResponseWriter, r *rest.Request) {
		reqID := r.Header.Get(RequestIDHeader)

		// Generate a request ID if none is present
		if reqID == "" {
			reqID = uuid.New()
			r.Header.Set(RequestIDHeader, reqID)
		}

		// Add the request ID to the response headers
		w.Header().Set(RequestIDHeader, reqID)

		// Handle the request
		h(w, r)
	}
}
