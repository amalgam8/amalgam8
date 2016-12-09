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

package debug

import (
	"net/http"

	"sync"

	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/ant0ine/go-json-rest/rest"
)

// API handles debugging API calls to sidecar for checking state
type API struct {
	instances []api.ServiceInstance
	rules     []api.Rule
	mutex     sync.Mutex
}

// NewAPI creates struct
func NewAPI() *API {
	return &API{
		instances: []api.ServiceInstance{},
		rules:     []api.Rule{},
	}
}

// Routes for debug API
func (d *API) Routes(middlewares ...rest.Middleware) []*rest.Route {
	routes := []*rest.Route{
		rest.Get("/state", d.checkState),
	}

	for _, route := range routes {
		route.Func = rest.WrapMiddlewares(middlewares, route.Func)
	}
	return routes
}

// checkState returns the cached rules from controller and cached instances
// from registry stored in sidecar memory
func (d *API) checkState(w rest.ResponseWriter, req *rest.Request) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	state := struct {
		Instances []api.ServiceInstance `json:"instances"`
		Rules     []api.Rule            `json:"rules"`
	}{
		Instances: d.instances,
		Rules:     d.rules,
	}

	w.WriteHeader(http.StatusOK)
	w.WriteJson(&state)
}

// CatalogChange updates on a change in the catalog.
func (d *API) CatalogChange(instances []api.ServiceInstance) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.instances = instances
	return nil
}

// RuleChange updates NGINX on a change in the proxy configuration.
func (d *API) RuleChange(rules []api.Rule) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.rules = rules
	return nil
}
