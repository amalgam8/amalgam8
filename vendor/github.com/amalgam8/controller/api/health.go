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
	"net/http"

	"github.com/amalgam8/controller/metrics"
	"github.com/ant0ine/go-json-rest/rest"
)

// Health handles health API calls
type Health struct {
	reporter metrics.Reporter
}

// NewHealth creates struct
func NewHealth(reporter metrics.Reporter) *Health {
	return &Health{
		reporter: reporter,
	}
}

// Routes for health check API
func (h *Health) Routes(middlewares ...rest.Middleware) []*rest.Route {
	routes := []*rest.Route{
		rest.Get("/health", reportMetric(h.reporter, h.GetHealth, "controller_health")),
	}

	for _, route := range routes {
		route.Func = rest.WrapMiddlewares(middlewares, route.Func)
	}
	return routes
}

// GetHealth performs health check on controller and dependencies
func (h *Health) GetHealth(w rest.ResponseWriter, req *rest.Request) error {
	// TODO: perform checks on cloudant, optionally SD and MH
	w.WriteHeader(http.StatusOK)
	return nil
}
