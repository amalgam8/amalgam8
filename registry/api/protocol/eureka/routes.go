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

package eureka

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/api/env"
	"github.com/amalgam8/amalgam8/registry/api/protocol"
	"github.com/amalgam8/amalgam8/registry/store"
	"github.com/amalgam8/amalgam8/registry/utils/i18n"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

const (
	module = "EUREKA"
)

// Routes encapsulates information needed for the eureka protocol routes
type Routes struct {
	catalogMap store.CatalogMap
	logger     *log.Entry
}

// New creates a new eureka Server instance
func New(catalogMap store.CatalogMap) *Routes {
	return &Routes{
		catalogMap: catalogMap,
		logger:     logging.GetLogger(module),
	}
}

// RouteHandlers returns an array of routes
func (routes *Routes) RouteHandlers(middlewares ...rest.Middleware) []*rest.Route {
	descriptors := []*protocol.APIDescriptor{
		{
			Path:      applicationTemplateURL(),
			Method:    "POST",
			Protocol:  protocol.Eureka,
			Operation: protocol.RegisterInstance,
			Handler:   routes.registerInstance,
		},
		{
			Path:      instanceTemplateURL(),
			Method:    "DELETE",
			Protocol:  protocol.Eureka,
			Operation: protocol.DeregisterInstance,
			Handler:   routes.deregisterInstance,
		},
		{
			Path:      instanceTemplateURL(),
			Method:    "PUT",
			Protocol:  protocol.Eureka,
			Operation: protocol.RenewInstance,
			Handler:   routes.renewInstance,
		},
		{
			Path:      applicationsURL(),
			Method:    "GET",
			Protocol:  protocol.Eureka,
			Operation: protocol.ListServices,
			Handler:   routes.listApps,
		},
		{
			Path:      applicationsURLTrailingSlash(),
			Method:    "GET",
			Protocol:  protocol.Eureka,
			Operation: protocol.ListServices,
			Handler:   routes.listApps,
		},
		{
			Path:      applicationTemplateURL(),
			Method:    "GET",
			Protocol:  protocol.Eureka,
			Operation: protocol.ListServiceInstances,
			Handler:   routes.listAppInstances,
		},
		{
			Path:      instanceTemplateURL(),
			Method:    "GET",
			Protocol:  protocol.Eureka,
			Operation: protocol.GetInstance,
			Handler:   routes.getInstance,
		},
		{
			Path:      instanceQueryTemplateURL(),
			Method:    "GET",
			Protocol:  protocol.Eureka,
			Operation: protocol.GetInstance,
			Handler:   routes.getInstance,
		},
		{
			Path:      instanceStatusTemplateURL(),
			Method:    "PUT",
			Protocol:  protocol.Eureka,
			Operation: protocol.SetInstanceStatus,
			Handler:   routes.setStatus,
		},
		{
			Path:      vipTemplateURL(),
			Method:    "GET",
			Protocol:  protocol.Eureka,
			Operation: protocol.ListServiceInstances,
			Handler:   routes.listVips,
		},
	}

	rts := make([]*rest.Route, 0, len(descriptors))
	for _, desc := range descriptors {
		desc.Handler = rest.WrapMiddlewares(middlewares, desc.Handler)
		desc.Handler = protocol.APIHandler(desc.Handler, desc.Protocol, desc.Operation)
		rts = append(rts, desc.AsRoute())
	}
	return rts
}

func (routes *Routes) catalog(w rest.ResponseWriter, r *rest.Request) store.Catalog {
	if r.Env[env.Namespace] == nil {
		i18n.Error(r, w, http.StatusUnauthorized, i18n.ErrorNamespaceNotFound)
		return nil
	}
	namespace := r.Env[env.Namespace].(auth.Namespace)
	if catalog, err := routes.catalogMap.GetCatalog(namespace); err != nil {
		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorInternalServer)
		return nil
	} else if catalog == nil {
		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorNamespaceNotFound)
		return nil
	} else {
		return catalog
	}
}
