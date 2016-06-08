package middleware

import (
	"errors"
	"fmt"
	"net/http"
)

// SetupHandler structure for filtering REST calls during setup of microservice
type SetupHandler struct {
	handler http.Handler
	err     error
}

// NewSetupHandler creates a new SetupHandler
func NewSetupHandler() *SetupHandler {
	h := new(SetupHandler)
	h.handler = nil
	h.err = errors.New("Service setup incomplete")
	return h
}

// SetHandler creates handler
func (h *SetupHandler) SetHandler(handler http.Handler) {
	h.handler = handler
	h.SetError(nil)
}

// SetError for handler
func (h *SetupHandler) SetError(err error) {
	h.err = err
}

// ServeHTTP serve http or return 503 while setup is still in progress
func (h *SetupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if h.err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Header().Set("Content-type", "application/json")
		w.Write([]byte(fmt.Sprintf("{\"error\":\"%v\"}", h.err.Error())))
	} else if h.handler == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		h.handler.ServeHTTP(w, r)
	}
}
