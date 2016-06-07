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
