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

	"github.com/amalgam8/amalgam8/registry/api/env"
	"github.com/amalgam8/amalgam8/registry/utils/i18n"
)

func (routes *Routes) getServiceInstances(w rest.ResponseWriter, r *rest.Request) {
	sname := r.PathParam(RouteParamServiceName)
	if sname == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "service name is required",
		}).Warn("Failed to lookup service")

		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorServiceNameMissing)
		return
	}

	catalog := routes.catalog(w, r)
	if catalog == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "catalog is nil",
		}).Errorf("Failed to lookup service %s", sname)
		// error is set in routes.catalog()
		return
	}

	if instances, err := catalog.List(sname, nil); err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Errorf("Failed to lookup service %s", sname)

		i18n.Error(r, w, statusCodeFromError(err), i18n.ErrorServiceEnumeration)
		return
	} else if instances == nil || len(instances) == 0 {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "no such service name",
		}).Warnf("Failed to lookup service %s", sname)

		i18n.Error(r, w, http.StatusNotFound, i18n.ErrorServiceNotFound)
		return
	} else {
		insts := make([]*ServiceInstance, len(instances))
		for index, si := range instances {
			inst, err := copyInstanceWithFilter(sname, si, nil)
			if err != nil {
				routes.logger.WithFields(log.Fields{
					"namespace": r.Env[env.Namespace],
					"error":     err,
				}).Warnf("Failed to lookup service %s", sname)

				i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorFilterGeneric)
				return
			}
			insts[index] = inst
		}

		if err := w.WriteJson(&InstanceList{ServiceName: sname, Instances: insts}); err != nil {
			routes.logger.WithFields(log.Fields{
				"namespace": r.Env[env.Namespace],
				"error":     err,
			}).Warnf("Failed to encode lookup response for %s", sname)

			i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorInternalServer)
			return
		}

		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
		}).Infof("Lookup service %s (%d)", sname, len(insts))
	}
}

func (routes *Routes) listServices(w rest.ResponseWriter, r *rest.Request) {
	catalog := routes.catalog(w, r)
	if catalog == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "catalog is nil",
		}).Error("Failed to list services")
		// error to user is already set in route.catalog()
		return
	}

	services := catalog.ListServices(nil)
	if services == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "services list is nil",
		}).Error("Failed to list services")

		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorServiceEnumeration)
		return
	}
	listRes := &ServicesList{Services: make([]string, len(services), len(services))}

	for index, svc := range services {
		listRes.Services[index] = svc.ServiceName
	}

	err := w.WriteJson(listRes)
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warn("Failed to encode services list")

		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorEncoding)
		return
	}

	routes.logger.WithFields(log.Fields{
		"namespace": r.Env[env.Namespace],
	}).Infof("List services (%d)", len(listRes.Services))
}
