package middleware

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/pborman/uuid"
)

// RequestIDHeader constant for "request-id"
const RequestIDHeader = "request-id"

// CSBIncidentIDHeader constant for csb's equivalent of SP's request-id header
const CSBIncidentIDHeader = "X-CSB-Trace"

// RequestIDMiddleware appends a request ID header to incoming request for log tracing across services
type RequestIDMiddleware struct{}

// MiddlewareFunc makes RequestIDMiddleware implement the Middleware interface.
func (mw *RequestIDMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	return func(w rest.ResponseWriter, r *rest.Request) {
		// check if CSB gave us a reqid to use before generating one of our own
		// this would happen on a passthru call to broker for prov/deprov for example
		csbReqID := r.Header.Get(CSBIncidentIDHeader)
		reqID := r.Header.Get(RequestIDHeader)
		if csbReqID != "" && reqID == "" {
			r.Header.Set(RequestIDHeader, csbReqID)
			reqID = csbReqID
		} else if reqID == "" {
			// add request ID if none present
			reqID = uuid.New()
			r.Header.Set(RequestIDHeader, reqID)
		}
		w.Header().Set(RequestIDHeader, reqID)

		h(w, r)

	}
}
