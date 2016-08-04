package middleware

import (
	"net/http"

	"strings"

	"github.com/amalgam8/controller/api"
	"github.com/amalgam8/controller/auth"
	"github.com/amalgam8/controller/util"
	"github.com/ant0ine/go-json-rest/rest"
)

const adminNamespace = "admin"

// AuthMiddleware provides a generic authentication middleware
// On failure, a 401 HTTP response is returned. On success, the wrapped middleware is called.
type AuthMiddleware struct {
	Authenticator auth.Authenticator
}

// MiddlewareFunc returns a go-json-rest HTTP Handler function, wrapping calls to the provided HandlerFunc
func (mw *AuthMiddleware) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	if mw.Authenticator == nil {
		mw.Authenticator = auth.DefaultAuthenticator()
	}

	return func(writer rest.ResponseWriter, request *rest.Request) { mw.handler(writer, request, handler) }
}

func (mw *AuthMiddleware) handler(writer rest.ResponseWriter, request *rest.Request, h rest.HandlerFunc) {
	authHeader := request.Header.Get(util.AuthHeader) // for Amalgam8 requests
	token := ""

	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || (parts[0] != "Bearer" && parts[0] != "bearer") {
			api.RestError(writer, request, http.StatusUnauthorized, "error_auth_header_malformed")
			return
		}
		token = parts[1]
	}

	nsPtr, err := mw.Authenticator.Authenticate(token)
	if err != nil {
		switch err {
		case auth.ErrEmptyToken:
			api.RestError(writer, request, http.StatusUnauthorized, "error_auth_header_missing")
		case auth.ErrUnauthorized, auth.ErrUnrecognizedToken:
			api.RestError(writer, request, http.StatusUnauthorized, "error_auth_not_authorized")
		case auth.ErrCommunicationError:
			api.RestError(writer, request, http.StatusServiceUnavailable, "error_auth_failed_validation")
		default:
			api.RestError(writer, request, http.StatusInternalServerError, "error_internal")
		}
		return
	}

	// Recognize admin namespace and get tenant ID from header
	if nsPtr.String() == adminNamespace {
		tenantID := request.Header.Get(util.TenantHeader)
		if tenantID == "" {
			api.RestError(writer, request, http.StatusBadRequest, "missing_tenant_header")
			return
		}
		tenantNamespace := auth.Namespace(tenantID)
		nsPtr = &tenantNamespace
	}

	request.Env[util.Namespace] = *nsPtr
	h(writer, request)
}
