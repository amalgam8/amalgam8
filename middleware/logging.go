package middleware

import (
	"github.com/Sirupsen/logrus"
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
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
