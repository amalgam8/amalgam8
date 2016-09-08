// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package middleware

import (
	"net/http"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/api/env"
	"github.com/amalgam8/amalgam8/registry/utils/i18n"
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
		case auth.ErrCommunicationError:
			i18n.Error(request, writer, http.StatusServiceUnavailable, i18n.ErrorAuthorizationTokenValidationFailed)
		default:
			i18n.Error(request, writer, http.StatusInternalServerError, i18n.ErrorInternalServer)
		}
		return
	}

	request.Env[env.Namespace] = *nsPtr
	h(writer, request)
}
