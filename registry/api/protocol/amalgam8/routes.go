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

package amalgam8

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
	module = "AMALGAM8"
)

// Routes encapsulates information needed for the aykesd protocol routes
type Routes struct {
	catalogMap store.CatalogMap
	logger     *log.Entry
}

// New creates a Routes object for the amalgam8 protocol routes
func New(catalogMap store.CatalogMap) *Routes {
	return &Routes{catalogMap, logging.GetLogger(module)}
}

// RouteHandlers returns an array of route handlers
func (routes *Routes) RouteHandlers(middlewares ...rest.Middleware) []*rest.Route {
	descriptors := []*protocol.APIDescriptor{
		{
			Path:      ServiceNamesURL(),
			Method:    "GET",
			Protocol:  protocol.Amalgam8,
			Operation: protocol.ListServices,
			Handler:   routes.listServices,
		},
		{
			Path:      serviceInstancesTemplateURL(),
			Method:    "GET",
			Protocol:  protocol.Amalgam8,
			Operation: protocol.ListServiceInstances,
			Handler:   routes.getServiceInstances,
		},
		{
			Path:      InstanceCreateURL(),
			Method:    "POST",
			Protocol:  protocol.Amalgam8,
			Operation: protocol.RegisterInstance,
			Handler:   routes.registerInstance,
		},
		{
			Path:      InstancesURL(),
			Method:    "GET",
			Protocol:  protocol.Amalgam8,
			Operation: protocol.ListInstances,
			Handler:   routes.listInstances,
		},
		{
			Path:      instanceTemplateURL(),
			Method:    "DELETE",
			Protocol:  protocol.Amalgam8,
			Operation: protocol.DeregisterInstance,
			Handler:   routes.deregisterInstance,
		},
		{
			Path:      instanceHeartbeatTemplateURL(),
			Method:    "PUT",
			Protocol:  protocol.Amalgam8,
			Operation: protocol.RenewInstance,
			Handler:   routes.renewInstance,
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
