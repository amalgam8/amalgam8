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

func (routes *Routes) listApps(w rest.ResponseWriter, r *rest.Request) {

	catalog := routes.catalog(w, r)
	if catalog == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "catalog is nil",
		}).Error("Failed to list applications")

		return
	}

	services := catalog.ListServices(nil)
	if services == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "services list is nil",
		}).Error("Failed to list applications")

		i18n.Error(r, w, http.StatusInternalServerError, i18n.EurekaErrorApplicationEnumeration)
		return
	}

	var instsCount int
	apps := &Applications{Application: make([]*Application, len(services))}
	listRes := &ApplicationsList{Applications: apps}

	for index, svc := range services {
		insts, err := catalog.List(svc.ServiceName, nil)
		if err != nil {
			routes.logger.WithFields(log.Fields{
				"namespace": r.Env[env.Namespace],
				"error":     err,
			}).Errorf("Failed to lookup application %s", svc.ServiceName)

			i18n.Error(r, w, http.StatusInternalServerError, i18n.EurekaErrorApplicationEnumeration)
			return
		}

		app := &Application{Name: svc.ServiceName, Instances: make([]*Instance, len(insts))}
		for idx, inst := range insts {
			app.Instances[idx] = buildInstanceFromRegistry(inst)
			instsCount++
		}
		apps.Application[index] = app
	}

	err := w.WriteJson(listRes)
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warn("Failed to encode applications list")

		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorEncoding)
		return
	}

	routes.logger.WithFields(log.Fields{
		"namespace": r.Env[env.Namespace],
	}).Infof("List applications (%d apps, %d insts)", len(apps.Application), instsCount)
}

func (routes *Routes) listAppInstances(w rest.ResponseWriter, r *rest.Request) {
	appid := r.PathParam(RouteParamAppID)
	if appid == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "application id is required",
		}).Warn("Failed to lookup application")

		i18n.Error(r, w, http.StatusBadRequest, i18n.EurekaErrorApplicationIdentifierMissing)
		return
	}

	catalog := routes.catalog(w, r)
	if catalog == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "catalog is nil",
		}).Errorf("Failed to lookup application %s", appid)

		return
	}

	insts, err := catalog.List(appid, nil)
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Errorf("Failed to lookup application %s", appid)

		i18n.Error(r, w, http.StatusInternalServerError, i18n.EurekaErrorApplicationNotFound)
		return
	}

	app := &Application{Name: appid, Instances: make([]*Instance, len(insts))}
	for index, inst := range insts {
		app.Instances[index] = buildInstanceFromRegistry(inst)
	}

	err = w.WriteJson(map[string]*Application{"application": app})
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warn("Failed to encode application")

		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorEncoding)
		return
	}

	routes.logger.WithFields(log.Fields{
		"namespace": r.Env[env.Namespace],
	}).Infof("List application instances (%d)", len(app.Instances))
}
