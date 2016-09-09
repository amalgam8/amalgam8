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
	"bytes"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/api/env"
	"github.com/amalgam8/amalgam8/registry/api/protocol"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

const (
	module         = "ACCESS"
	clientIPHeader = "X-Client-Ip"
)

var (
	// headersWhitelist holds a set of header names that will be logged.
	headersWhitelist = map[string]struct{}{}
)

// AccessLog produces the access log.
// It depends on TimerMiddleware and RecorderMiddleware that should be in the wrapped
// middlewares. It also uses request.Env[Namespace] set by the auth middlewares.
type AccessLog struct {
	logger *log.Entry
}

// MiddlewareFunc makes AccessLogApacheMiddleware implement the Middleware interface.
func (mw *AccessLog) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	mw.logger = logging.GetLogger(module).WithField("apptype", "service-discovery")

	return func(w rest.ResponseWriter, r *rest.Request) {
		// We log the message in a defer function to make sure that the message
		// is logged even if a panic occurs in some handler in the chain
		defer func() {
			reqID, ok := r.Env[env.RequestID].(string)
			if !ok {
				reqID = "Unknown"
			}

			l := mw.logger.WithFields(log.Fields{
				"sd-request-id": reqID,
				"namespace":     mw.namespace(r),
				"method":        mw.method(r),
				"protocol":      mw.protocol(r),
				"returncode":    mw.statusCode(r),
				"byteswritten":  mw.bytesWritten(r),
				"elapsedtime":   mw.elapsedTime(r)})

			if len(headersWhitelist) > 0 {
				l = l.WithField("headers", mw.headers(r))
			}
			l.Infof("%s %s %s %s", mw.remoteAddr(r), r.Method, r.RequestURI, r.Proto)
		}()

		// call the handler
		h(w, r)
	}
}

func (mw *AccessLog) remoteAddr(r *rest.Request) string {
	remoteAddr := r.Header.Get(clientIPHeader)
	if remoteAddr != "" {
		return remoteAddr
	}
	remoteAddr = r.RemoteAddr
	if remoteAddr != "" {
		parts := strings.SplitN(remoteAddr, ":", 2)
		return parts[0]
	}
	return ""
}

func (mw *AccessLog) namespace(r *rest.Request) auth.Namespace {
	if val, exists := r.Env[env.Namespace]; exists {
		return val.(auth.Namespace)
	}
	return ""
}

func (mw *AccessLog) headers(r *rest.Request) string {
	var buf bytes.Buffer
	first := true
	for name, values := range r.Header {
		if _, exists := headersWhitelist[name]; exists {
			if first {
				first = false
			} else {
				_, _ = buf.WriteString(";")
			}
			_, _ = buf.WriteString(name)
			_, _ = buf.WriteString(":")
			_, _ = buf.WriteString(strings.Join(values, ","))
		}
	}
	return buf.String()
}

func (mw *AccessLog) method(r *rest.Request) string {
	if val, exists := r.Env[env.APIOperation]; exists {
		return val.(protocol.Operation).String()
	}
	return r.Method
}

func (mw *AccessLog) protocol(r *rest.Request) string {
	if val, exists := r.Env[env.APIProtocol]; exists {
		return protocol.NameOf(val.(protocol.Type))
	}
	return r.Proto
}

func (mw *AccessLog) statusCode(r *rest.Request) int {
	if val, exists := r.Env[env.StatusCode]; exists {
		return val.(int)
	}
	return http.StatusInternalServerError
}

func (mw *AccessLog) elapsedTime(r *rest.Request) *time.Duration {
	if val, exists := r.Env[env.ElapsedTime]; exists {
		return val.(*time.Duration)
	}
	return nil
}

func (mw *AccessLog) bytesWritten(r *rest.Request) int64 {
	if val, exists := r.Env[env.BytesWritten]; exists {
		return val.(int64)
	}
	return 0
}
