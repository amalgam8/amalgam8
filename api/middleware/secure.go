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

	"github.com/ant0ine/go-json-rest/rest"
)

// CheckRequest defines the request validation configuration used by RequireHTTPS middleware
type CheckRequest struct {
	// Defines a function that, given a request, determines if the request is secure or should be redirected.
	// Criteria may include, for example, scheme, X-Forward* headers, URL paths, etc. If unset, a default
	// criteria is applied, checking the request's scheme and TLS configuration
	IsSecure func(r *rest.Request) bool
	// When developing, various options can cause unwanted effects (e.g., usually testing happens on localhost with
	// http, not https, etc.). The following flag can set to true in a development environment to disable checks
	// and redirection. Default is false (i.e., it is enabled)
	Disabled bool
}

// RequireHTTPS provide HTTP middleware to ensure requests are received over secure communication.
// By default, all requests using HTTP are redirected to the corresponding HTTPS URL.
type RequireHTTPS struct {
	check CheckRequest
}

// NewRequireHTTPS creates a new middleware with the given checks configured
func NewRequireHTTPS(checkConfig CheckRequest) *RequireHTTPS {
	c := checkConfig
	if c.IsSecure == nil {
		c.IsSecure = IsUsingSecureConnection
	}

	return &RequireHTTPS{
		check: c,
	}
}

// MiddlewareFunc returns a go-json-rest HTTP Handler function, wrapping calls to the provided HandlerFunc
func (secure *RequireHTTPS) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(writer rest.ResponseWriter, request *rest.Request) { secure.handler(writer, request, handler) }
}

func (secure *RequireHTTPS) handler(w rest.ResponseWriter, r *rest.Request, h rest.HandlerFunc) {
	isSecure := secure.check.Disabled || secure.check.IsSecure(r)

	if !isSecure {
		url := r.BaseUrl()
		url.Scheme = "https"
		url.Path = r.URL.Path
		url.RawQuery = r.URL.RawQuery

		w.Header().Set("Location", url.String())
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}
	h(w, r)
}

// IsUsingSecureConnection returns true if the request is using secure connection or false otherwise
func IsUsingSecureConnection(r *rest.Request) bool {
	return IsHTTPS(r) || IsProxiedHTTPS(r)
}

// IsHTTPS returns true if the request is using https protocol or false otherwise
func IsHTTPS(r *rest.Request) bool {
	return r.URL.Scheme == "https" || r.TLS != nil
}

// IsProxiedHTTPS returns true if the original request is using https protocol or false otherwise
func IsProxiedHTTPS(r *rest.Request) bool {
	return r.Header.Get("X-Forwarded-Proto") == "https" || r.Header.Get("$WSSC") == "https"
}
