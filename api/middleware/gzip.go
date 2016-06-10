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
	"bufio"
	"compress/gzip"
	"net"
	"net/http"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"
)

// GzipMiddleware is responsible for compressing the payload with gzip and setting the proper
// headers when supported by the client.
type GzipMiddleware struct{}

// MiddlewareFunc makes GzipMiddleware implement the Middleware interface.
func (mw *GzipMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	return func(w rest.ResponseWriter, r *rest.Request) {
		// Always set the Vary header, even if this particular request
		// is not gzipped.
		w.Header().Add("Vary", "Accept-Encoding")

		if isGzipEnabled(r) {
			w.Header().Set("Content-Encoding", "gzip")
			writer := &gzipResponseWriter{w, false, nil}
			h(writer, r)

			// Close the gzipWriter (if exists) to set the EOF
			if writer.gzipWriter != nil {
				writer.gzipWriter.Close()
			}
		} else {
			// No need to wrap the responseWriter if gzip is disabled
			h(w, r)
		}
	}
}

type gzipResponseWriter struct {
	rest.ResponseWriter
	wroteHeader bool
	gzipWriter  *gzip.Writer
}

// WriteHeader sends an HTTP response header with status code.
// Set the right headers for gzip encoded responses.
func (w *gzipResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.wroteHeader = true
}

// WriteJson uses the EncodeJson to generate the payload and writes it to the local writer.
func (w *gzipResponseWriter) WriteJson(v interface{}) error {
	b, err := w.EncodeJson(v)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	if err != nil {
		return err
	}

	return nil
}

// Flush sends any buffered data to the client.
// Make sure the local WriteHeader is called, and call the parent Flush.
// Provided in order to implement the http.Flusher interface.
func (w *gzipResponseWriter) Flush() {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	flusher := w.ResponseWriter.(http.Flusher)
	flusher.Flush()
}

// CloseNotify returns a channel that receives a single value
// when the client connection has gone away.
// Provided in order to implement the http.CloseNotifier interface.
func (w *gzipResponseWriter) CloseNotify() <-chan bool {
	notifier := w.ResponseWriter.(http.CloseNotifier)
	return notifier.CloseNotify()
}

// Hijack lets the caller take over the connection.
// Provided in order to implement the http.Hijacker interface.
func (w *gzipResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker := w.ResponseWriter.(http.Hijacker)
	return hijacker.Hijack()
}

// Write writes the data to the connection as part of an HTTP reply.
// Make sure the local WriteHeader is called, and encode the payload if necessary.
// Provided in order to implement the http.ResponseWriter interface.
func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	writer := w.ResponseWriter.(http.ResponseWriter)

	if w.gzipWriter == nil {
		w.gzipWriter = gzip.NewWriter(writer)
	}
	count, errW := w.gzipWriter.Write(b)
	errF := w.gzipWriter.Flush()
	if errW != nil {
		return count, errW
	}
	if errF != nil {
		return count, errF
	}
	return count, nil
}

func isGzipEnabled(r *rest.Request) bool {
	val := r.Header.Get("Accept-Encoding")
	if val == "gzip" {
		return true
	}

	if val == "" {
		return false
	}

	for _, val = range r.Header["Accept-Encoding"] {
		splitVal := strings.Split(val, ",")
		for _, fv := range splitVal {
			if fv == "gzip" {
				return true
			}
		}
	}
	return false
}
