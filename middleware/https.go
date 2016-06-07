package middleware

import (
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
	"strings"
)

// XForwardedProtoHeader HTTPS headers expected from bluemix
const XForwardedProtoHeader = "X-Forwarded-Proto"

// WSSCHeader HTTPS headers expected from bluemix
const WSSCHeader = "$WSSC"

// HTTPSMiddleware filters non-HTTPS forwarded REST calls
type HTTPSMiddleware struct{}

// MiddlewareFunc checks for either the presence of either of the HTTPS headers in the request.
// If neither is in the request, it returns an error code. Otherwise it propagates the request.
func (mw *HTTPSMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	return func(w rest.ResponseWriter, r *rest.Request) {
		// Check for either HTTPS header
		if strings.ToLower(r.Header.Get(XForwardedProtoHeader)) != "https" &&
			strings.ToLower(r.Header.Get(WSSCHeader)) != "https" {
			w.WriteHeader(http.StatusMovedPermanently)
		} else {
			h(w, r)
		}
	}
}
