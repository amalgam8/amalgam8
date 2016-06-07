package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/stretchr/testify/assert"
)

const internalHandlerBody = "internal"

var encapsulated = rest.HandlerFunc(func(w rest.ResponseWriter, r *rest.Request) {
	// w.WriteJson would just be a hassle here, so we need to explicitly set the content type and cast the jrest
	// type to net/http types
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.(http.ResponseWriter).Write([]byte(internalHandlerBody))
})

func jrestServer(mw rest.Middleware, endpoint string) http.Handler {
	restAPI := rest.NewApi()
	restAPI.Use(mw)

	router, err := rest.MakeRouter(rest.Get(endpoint, encapsulated))

	if err != nil {
		return nil
	}

	restAPI.SetApp(router)
	return restAPI.MakeHandler()
}

func TestEnabledByDefault(t *testing.T) {
	s := NewRequireHTTPS(CheckRequest{})
	assert.NotNil(t, s)
	assert.False(t, s.check.Disabled)

	s = NewRequireHTTPS(CheckRequest{IsSecure: func(r *rest.Request) bool { return true }})
	assert.NotNil(t, s)
	assert.False(t, s.check.Disabled)
}

func TestCreateAlwaysSetsRequestCheck(t *testing.T) {
	s := NewRequireHTTPS(CheckRequest{})
	assert.NotNil(t, s.check.IsSecure)

	s = NewRequireHTTPS(CheckRequest{Disabled: true})
	assert.NotNil(t, s.check.IsSecure)

	s = NewRequireHTTPS(CheckRequest{Disabled: true, IsSecure: func(r *rest.Request) bool { return true }})
	assert.NotNil(t, s.check.IsSecure)
}

func TestRequestCheckSetFromConfig(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com/", nil)
	wrapped := &rest.Request{
		Request:    req,
		PathParams: nil,
		Env:        map[string]interface{}{},
	}

	s := NewRequireHTTPS(CheckRequest{Disabled: true, IsSecure: func(r *rest.Request) bool { return false }})
	assert.False(t, s.check.IsSecure(wrapped))
	s = NewRequireHTTPS(CheckRequest{Disabled: true, IsSecure: func(r *rest.Request) bool { return true }})
	assert.True(t, s.check.IsSecure(wrapped))
}

func TestDisabledDoesNotRedirectHTTP(t *testing.T) {
	// IsSecure will always return false, but it is disabled
	s := NewRequireHTTPS(CheckRequest{Disabled: true, IsSecure: func(r *rest.Request) bool { return false }})
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/", nil)

	jrestServer(s, "/").ServeHTTP(res, req)

	assert.Equal(t, res.Code, http.StatusOK)
	assert.Equal(t, internalHandlerBody, res.Body.String())
}

func TestRedirectsHTTPByDefault(t *testing.T) {
	s := NewRequireHTTPS(CheckRequest{})
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/", nil)

	jrestServer(s, "/").ServeHTTP(res, req)

	assert.Equal(t, res.Code, http.StatusMovedPermanently)
	assert.NotEqual(t, internalHandlerBody, res.Body.String())
}

func TestNoRedirectHTTPSByDefault(t *testing.T) {
	s := NewRequireHTTPS(CheckRequest{})
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "https://example.com/", nil)

	jrestServer(s, "/").ServeHTTP(res, req)

	assert.Equal(t, res.Code, http.StatusOK)
	assert.Equal(t, internalHandlerBody, res.Body.String())
}

func TestRedirectChangesOnlyScheme(t *testing.T) {
	s := NewRequireHTTPS(CheckRequest{})
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/?key1=value&key2=value2", nil)
	original := *req.URL //save a copy

	jrestServer(s, "/").ServeHTTP(res, req)

	assert.Equal(t, res.Code, http.StatusMovedPermanently)
	target, err := url.Parse(res.Header().Get("Location"))
	assert.NoError(t, err)
	assert.Equal(t, original.Host, target.Host)
	assert.Equal(t, original.Path, target.Path)
	assert.Equal(t, original.RawQuery, target.RawQuery)
	assert.NotEqual(t, original.Scheme, target.Scheme)
	assert.Equal(t, target.Scheme, "https")
}

func checkProxiedHTTPSHeaders(r *rest.Request) bool {
	return r.Header.Get("X-Forwarded-Proto") == "https" || r.Header.Get("$WSSC") == "https"
}

func TestCustomRequestHeadersCheck(t *testing.T) {
	s := NewRequireHTTPS(CheckRequest{IsSecure: checkProxiedHTTPSHeaders})
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")

	jrestServer(s, "/").ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusOK)
	assert.Equal(t, internalHandlerBody, res.Body.String())
}

func checkBenignPath(r *rest.Request) bool {
	return r.URL.Path == "/benign"
}

func TestSpecificRequestPathAllowed(t *testing.T) {
	s := NewRequireHTTPS(CheckRequest{IsSecure: checkBenignPath})
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/benign", nil)

	jrestServer(s, "/benign").ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusOK)
	assert.Equal(t, internalHandlerBody, res.Body.String())
}
