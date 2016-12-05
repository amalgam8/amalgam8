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

	"github.com/amalgam8/amalgam8/controller/rules"
	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/sidecar/proxy"
	"github.com/ant0ine/go-json-rest/rest"
)

// DebugAPI handles debugging API calls to sidecar for checking state
type DebugAPI struct {
	nginxProxy proxy.NGINXProxy
}

// NewDebugAPI creates struct
func NewDebugAPI(nginxProxy proxy.NGINXProxy) *DebugAPI {
	return &DebugAPI{
		nginxProxy: nginxProxy,
	}
}

// Routes for debug API
func (d *DebugAPI) Routes(middlewares ...rest.Middleware) []*rest.Route {
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
func (d *DebugAPI) checkState(w rest.ResponseWriter, req *rest.Request) {

	cachedInstances, cachedRules := d.nginxProxy.GetState()

	state := struct {
		Instances []api.ServiceInstance `json:"instances"`
		Rules     []rules.Rule          `json:"rules"`
	}{
		Instances: cachedInstances,
		Rules:     cachedRules,
	}

	w.WriteHeader(http.StatusOK)
	w.WriteJson(&state)

}
