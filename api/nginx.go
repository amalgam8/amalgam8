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
	"errors"
	"net/http"
	"time"

	"github.com/amalgam8/controller/metrics"
	"github.com/amalgam8/controller/nginx"
	"github.com/ant0ine/go-json-rest/rest"
)

// NGINXConfig options
type NGINXConfig struct {
	Reporter  metrics.Reporter
	Generator nginx.Generator
}

// NGINX handles NGINX API calls
type NGINX struct {
	reporter  metrics.Reporter
	generator nginx.Generator
}

// NewNGINX creates struct
func NewNGINX(nc NGINXConfig) *NGINX {
	return &NGINX{
		reporter:  nc.Reporter,
		generator: nc.Generator,
	}
}

// Routes for NGINX API calls
func (n *NGINX) Routes(middlewares ...rest.Middleware) []*rest.Route {
	routes := []*rest.Route{
		rest.Get("/v1/nginx", reportMetric(n.reporter, n.GetNGINX, "tenants_nginx")),
	}

	for _, route := range routes {
		route.Func = rest.WrapMiddlewares(middlewares, route.Func)
	}
	return routes
}

// GetNGINX returns the NGINX configuration for a given tenant
func (n *NGINX) GetNGINX(w rest.ResponseWriter, req *rest.Request) error {
	var err error

	tenantID := GetTenantID(req)
	if tenantID == "" {
		RestError(w, req, http.StatusBadRequest, "error_invalid_input")
		return errors.New("special error")
	}

	queries := req.URL.Query()
	var lastUpdate *time.Time
	if queries.Get("version") != "" {
		update, err := time.Parse(time.RFC3339, queries.Get("version"))
		if err == nil {
			lastUpdate = &update
		}
	}

	// Generate config
	templateConf, err := n.generator.Generate(tenantID, lastUpdate)
	if err != nil {
		RestError(w, req, http.StatusInternalServerError, "error_nginx_generator_failed")
		return err
	}

	if templateConf == nil {
		// No new config was generated
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	w.WriteHeader(http.StatusOK)
	w.WriteJson(templateConf)

	return nil
}
