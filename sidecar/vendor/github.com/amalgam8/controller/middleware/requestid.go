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
	"github.com/amalgam8/controller/util"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/pborman/uuid"
)

// RequestIDMiddleware appends a request ID header to incoming request for log tracing across services
type RequestIDMiddleware struct{}

// MiddlewareFunc makes RequestIDMiddleware implement the Middleware interface.
func (mw *RequestIDMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	return func(w rest.ResponseWriter, r *rest.Request) {
		reqID := r.Header.Get(util.RequestIDHeader)

		// Generate a request ID if none is present
		if reqID == "" {
			reqID = uuid.New()
			r.Header.Set(util.RequestIDHeader, reqID)
		}

		// Add the request ID to the response headers
		w.Header().Set(util.RequestIDHeader, reqID)

		// Handle the request
		h(w, r)
	}
}
