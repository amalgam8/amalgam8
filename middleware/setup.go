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
