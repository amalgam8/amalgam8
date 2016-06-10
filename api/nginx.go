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
	"bytes"
	"errors"
	"net/http"
	"time"

	"github.com/amalgam8/controller/checker"
	"github.com/amalgam8/controller/metrics"
	"github.com/amalgam8/controller/nginx"
	"github.com/ant0ine/go-json-rest/rest"
)

// NGINXConfig options
type NGINXConfig struct {
	Reporter  metrics.Reporter
	Generator nginx.Generator
	Checker   checker.Checker
}

// NGINX handles NGINX API calls
type NGINX struct {
	reporter  metrics.Reporter
	generator nginx.Generator
	checker   checker.Checker
}

// NewNGINX creates struct
func NewNGINX(nc NGINXConfig) *NGINX {
	return &NGINX{
		reporter:  nc.Reporter,
		generator: nc.Generator,
		checker:   nc.Checker,
	}
}

// Routes for NGINX API calls
func (n *NGINX) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/tenants/#id/nginx", reportMetric(n.reporter, n.GetNGINX, "tenants_nginx")),
	}
}

// GetNGINX returns the NGINX configuration for a given tenant
func (n *NGINX) GetNGINX(w rest.ResponseWriter, req *rest.Request) error {
	var err error

	id := req.PathParam("id")
	queries := req.URL.Query()
	var lastUpdate *time.Time
	if queries.Get("version") != "" {
		update, err := time.Parse(time.RFC3339, queries.Get("version"))
		if err == nil {
			lastUpdate = &update
		}
	}

	catalog, err := n.checker.Get(id)
	if err != nil {
		handleDBError(w, req, err)
		return err
	}

	// if version query is newer than latest rules change, return 204
	if lastUpdate != nil && catalog.LastUpdate.Before(*lastUpdate) {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	// Generate config
	buf := bytes.NewBuffer([]byte{})
	if err = n.generator.Generate(buf, id); err != nil {
		RestError(w, req, http.StatusInternalServerError, "error_nginx_generator_failed")
		return err
	}

	// Write response as text
	httpWriter, ok := w.(http.ResponseWriter)
	if !ok {
		RestError(w, req, http.StatusInternalServerError, "error_internal")
		return errors.New("Could not cast rest.ResponseWriter to http.ResponseWriter")
	}
	httpWriter.Header().Set("Content-type", "text/plain")
	httpWriter.WriteHeader(http.StatusOK)
	httpWriter.Write(buf.Bytes())

	return nil
}
