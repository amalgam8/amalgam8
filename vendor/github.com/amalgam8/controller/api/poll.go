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

	"github.com/amalgam8/controller/checker"
	"github.com/amalgam8/controller/metrics"
	"github.com/ant0ine/go-json-rest/rest"
)

// Poll handles poll API
type Poll struct {
	checker  checker.Checker
	reporter metrics.Reporter
}

// NewPoll create struct
func NewPoll(reporter metrics.Reporter, checker checker.Checker) *Poll {
	return &Poll{
		reporter: reporter,
		checker:  checker,
	}
}

// Routes for poll API
func (p *Poll) Routes(middlewares ...rest.Middleware) []*rest.Route {
	routes := []*rest.Route{
		rest.Post("/v1/poll", reportMetric(p.reporter, p.Poll, "poll")),
	}
	for _, route := range routes {
		route.Func = rest.WrapMiddlewares(middlewares, route.Func)
	}
	return routes
}

// Poll Registry for latest changes
func (p *Poll) Poll(w rest.ResponseWriter, req *rest.Request) error {
	if err := p.checker.Check(nil); err != nil {
		RestError(w, req, http.StatusInternalServerError, "failed")
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
