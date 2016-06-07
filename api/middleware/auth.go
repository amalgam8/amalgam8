package middleware

import (
	"net/http"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/amalgam8/registry/auth"
	"github.com/amalgam8/registry/utils/i18n"
)

const (
	// NamespaceKey defines the name of the namespace key in the Env map
	NamespaceKey = "namespace"
)

// AuthMiddleware provides a generic authentication middleware
// On failure, a 401 HTTP response is returned. On success, the wrapped middleware is called.
type AuthMiddleware struct {
	TokenRouteParam string
	Authenticator   auth.Authenticator
}

// MiddlewareFunc returns a go-json-rest HTTP Handler function, wrapping calls to the provided HandlerFunc
func (mw *AuthMiddleware) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	if mw.Authenticator == nil {
		mw.Authenticator = auth.DefaultAuthenticator()
	}

	return func(writer rest.ResponseWriter, request *rest.Request) { mw.handler(writer, request, handler) }
}

func (mw *AuthMiddleware) handler(writer rest.ResponseWriter, request *rest.Request, h rest.HandlerFunc) {
	authHeader := request.Header.Get("Authorization") // for Amalgam8 requests
	token := request.PathParam(mw.TokenRouteParam)    // for Eureka requests

	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || (parts[0] != "Bearer" && parts[0] != "bearer") {
			i18n.Error(request, writer, http.StatusUnauthorized, i18n.ErrorAuthorizationMalformedHeader)
			return
		}
		token = parts[1]
	}

	nsPtr, err := mw.Authenticator.Authenticate(token)
	if err != nil {
		switch err {
		case auth.ErrEmptyToken:
			i18n.Error(request, writer, http.StatusUnauthorized, i18n.ErrorAuthorizationMissingHeader)
		case auth.ErrUnauthorized, auth.ErrUnrecognizedToken:
			i18n.Error(request, writer, http.StatusUnauthorized, i18n.ErrorAuthorizationNotAuthorized)
		default:
			i18n.Error(request, writer, http.StatusInternalServerError, i18n.ErrorInternalServer)
		}
		return
	}

	request.Env[NamespaceKey] = *nsPtr
	request.Env["REMOTE_USER"] = nsPtr.String()
	h(writer, request)
}
