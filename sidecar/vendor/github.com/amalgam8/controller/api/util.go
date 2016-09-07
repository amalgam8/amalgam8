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

package api

import (
	"time"

	"github.com/amalgam8/amalgam8/controller/auth"
	"github.com/amalgam8/amalgam8/controller/metrics"
	"github.com/amalgam8/amalgam8/controller/util"
	"github.com/ant0ine/go-json-rest/rest"
)

// GetNamespace from a request
func GetNamespace(req *rest.Request) string {
	env := req.Env[util.Namespace]
	if namespace, ok := env.(auth.Namespace); ok {
		return namespace.String()
	}

	return ""
}

func reportMetric(reporter metrics.Reporter, f func(rest.ResponseWriter, *rest.Request) error, name string) rest.HandlerFunc {
	return func(w rest.ResponseWriter, req *rest.Request) {
		startTime := time.Now()
		err := f(w, req)
		endTime := time.Since(startTime)
		if err != nil {
			// Report failure
			reporter.Failure(name, endTime, err)
			return
		}
		// Report success
		reporter.Success(name, endTime)
	}
}
