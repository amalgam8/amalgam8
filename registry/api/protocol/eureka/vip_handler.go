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

	"github.com/amalgam8/amalgam8/registry/api/env"
	"github.com/amalgam8/amalgam8/registry/utils/i18n"
)

func (routes *Routes) listVips(w rest.ResponseWriter, r *rest.Request) {

	vip := r.PathParam(RouterParamVip)
	if vip == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "vip is required",
		}).Warn("Failed to list vip")

		i18n.Error(r, w, http.StatusBadRequest, i18n.EurekaErrorVIPRequired)
		return
	}

	catalog := routes.catalog(w, r)
	if catalog == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "catalog is nil",
		}).Errorf("Failed to list vip %s", vip)

		return
	}

	services := catalog.ListServices(nil)
	if services == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "services list is nil",
		}).Errorf("Failed to list vip %s", vip)

		i18n.Error(r, w, http.StatusInternalServerError, i18n.EurekaErrorVIPEnumeration)
		return
	}

	var instsCount int
	apps := &Applications{Application: make([]*Application, 0, len(services))}
	listRes := &ApplicationsList{Applications: apps}

	for _, svc := range services {
		insts, err := catalog.List(svc.ServiceName, nil)
		if err != nil {
			routes.logger.WithFields(log.Fields{
				"namespace": r.Env[env.Namespace],
				"error":     err,
			}).Errorf("Failed to list vips %s", vip)

			i18n.Error(r, w, http.StatusInternalServerError, i18n.EurekaErrorVIPEnumeration)
			return
		}

		app := &Application{Name: svc.ServiceName, Instances: make([]*Instance, 0, len(insts))}
		for _, inst := range insts {
			if vipaddr, ok := inst.Extension[extVIP]; ok {
				if vipaddr == vip {
					app.Instances = append(app.Instances, buildInstanceFromRegistry(inst))
					instsCount++
				}
			}
		}
		if len(app.Instances) > 0 {
			apps.Application = append(apps.Application, app)
		}
	}

	err := w.WriteJson(listRes)
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warn("Failed to encode vips list")

		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorEncoding)
		return
	}

	routes.logger.WithFields(log.Fields{
		"namespace": r.Env[env.Namespace],
	}).Infof("List vips (%d apps, %d insts)", len(apps.Application), instsCount)
}
