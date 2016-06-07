package middleware

import (
	"github.com/Sirupsen/logrus"
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
	"regexp"
)

// AuthHeader control plane authorization header
const AuthHeader = "Authorization"

// AuthMiddleware authenticates incoming requests
type AuthMiddleware struct {
	handler    http.Handler
	key        string
	whitelists []*regexp.Regexp
}

// MiddlewareFunc rejects unauthenticated requests and returns an error code.
// Otherwise it propagates the request.
func (mw *AuthMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	return func(w rest.ResponseWriter, r *rest.Request) {
		for _, whitelist := range mw.whitelists {
			if whitelist.MatchString(r.URL.Path) {
				h(w, r)
				return
			}
		}

		reqID := r.Header.Get(RequestIDHeader)

		// Check the header
		if r.Header.Get(AuthHeader) == mw.key {
			h(w, r)
		} else {
			logrus.WithFields(logrus.Fields{
				"remote_address": r.RemoteAddr,
				"request_id":     reqID,
				"method":         r.Method,
				"url":            r.URL,
			}).Error("Invalid authentication token")
			w.WriteHeader(http.StatusUnauthorized)
		}
	}
}
