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

	"github.com/Sirupsen/logrus"
	"github.com/ant0ine/go-json-rest/rest"
)

// LoggingMiddleware logs information about the request
type LoggingMiddleware struct{}

// MiddlewareFunc logs information about the request after it completes
func (mw *LoggingMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	return func(w rest.ResponseWriter, r *rest.Request) {
		statusWriter := statusRW{-1, w}

		h(&statusWriter, r)

		// log the exiting request
		logrus.WithFields(logrus.Fields{
			"request_id":  r.Header.Get(RequestIDHeader),
			"method":      r.Method,
			"url":         r.URL,
			"status_code": statusWriter.status,
		}).Info("Handled request")
	}
}

// wrapper response writer so we can capture the status code
type statusRW struct {
	status int
	rest.ResponseWriter
}

func (w *statusRW) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *statusRW) Write(b []byte) (int, error) {
	return w.ResponseWriter.(http.ResponseWriter).Write(b)
}

func (w *statusRW) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
