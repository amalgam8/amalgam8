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
	"net/http/httptest"
	"testing"

	"github.com/amalgam8/registry/auth"
	"github.com/stretchr/testify/assert"
)

const (
	validToken   = "valid-token"
	invalidToken = "invalid-token"
)

type mockAuthenticator struct {
	namespace *auth.Namespace
}

func (ma *mockAuthenticator) Authenticate(token string) (*auth.Namespace, error) {
	if token == validToken {
		return ma.namespace, nil
	}
	return nil, auth.ErrUnauthorized
}

func TestNoToken(t *testing.T) {
	namespace := auth.Namespace(validToken)
	ma := &mockAuthenticator{namespace: &namespace}
	authMw := &AuthMiddleware{Authenticator: ma}

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/", nil)

	jrestServer(authMw, "/").ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestTokenNotValid(t *testing.T) {
	namespace := auth.Namespace(invalidToken)
	ma := &mockAuthenticator{namespace: &namespace}
	authMw := &AuthMiddleware{Authenticator: ma}

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/", nil)
	req.Header.Set("Authorization", "Bearer "+invalidToken)

	jrestServer(authMw, "/").ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestDefaultAuthenticator(t *testing.T) {
	authMw := &AuthMiddleware{}

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/", nil)

	jrestServer(authMw, "/").ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
}
