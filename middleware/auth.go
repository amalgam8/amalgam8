package middleware

import (
	"github.com/Sirupsen/logrus"
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
)

// AuthHeader control plane authorization header
const (
	AuthHeader = "Authorization"
 	AuthEnv = "TENANT_ID"
	TenantHeader = "SP-Tenant-ID"
)

type Authenticator interface {
	Authenticate(token string) (string, error)
}

// AuthMiddleware authenticates incoming requests
type AuthMiddleware struct {
	Key        string
	Auth Authenticator
}

// MiddlewareFunc rejects unauthenticated requests and returns an error code.
// Otherwise it propagates the request.
func (mw *AuthMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	return func(w rest.ResponseWriter, r *rest.Request) {

		reqID := r.Header.Get(RequestIDHeader)
		authToken := r.Header.Get(AuthHeader)
		// Check the header
		if authToken == mw.Key {
			tenantID := r.Header.Get(TenantHeader)
			if tenantID != "" {
				r.Env[AuthEnv] = tenantID
				h(w, r)
				return
			} else {
				logrus.WithFields(logrus.Fields{
					"remote_address": r.RemoteAddr,
					"request_id":     reqID,
					"method":         r.Method,
					"url":            r.URL,
				}).Error("Missing tenant ID header")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		} else {

			if authToken != "" {
				id, err := mw.Auth.Authenticate(authToken)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"err": err,
						"remote_address": r.RemoteAddr,
						"request_id":     reqID,
						"method":         r.Method,
						"url":            r.URL,
					}).Error("Invalid authentication token")
					w.WriteHeader(http.StatusUnauthorized)
					return
				} else {
					r.Env[AuthEnv] = id
					h(w, r)
					return
				}

			} else {
				logrus.WithFields(logrus.Fields{
					"remote_address": r.RemoteAddr,
					"request_id":     reqID,
					"method":         r.Method,
					"url":            r.URL,
				}).Error("Invalid authentication token")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
	}
}