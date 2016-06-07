package eureka

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/amalgam8/registry/utils/i18n"
)

func (routes *Routes) listApps(w rest.ResponseWriter, r *rest.Request) {

	catalog := routes.catalog(w, r)
	if catalog == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env["REMOTE_USER"],
			"error":     "catalog is nil",
		}).Error("Failed to list applications")

		return
	}

	services := catalog.ListServices(nil)
	if services == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env["REMOTE_USER"],
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
				"namespace": r.Env["REMOTE_USER"],
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
			"namespace": r.Env["REMOTE_USER"],
			"error":     err,
		}).Warn("Failed to encode applications list")

		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorEncoding)
		return
	}

	routes.logger.WithFields(log.Fields{
		"namespace": r.Env["REMOTE_USER"],
	}).Infof("List applications (%d apps, %d insts)", len(apps.Application), instsCount)
}

func (routes *Routes) listAppInstances(w rest.ResponseWriter, r *rest.Request) {
	appid := r.PathParam(RouteParamAppID)
	if appid == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env["REMOTE_USER"],
			"error":     "application id is required",
		}).Warn("Failed to lookup application")

		i18n.Error(r, w, http.StatusBadRequest, i18n.EurekaErrorApplicationIdentifierMissing)
		return
	}

	catalog := routes.catalog(w, r)
	if catalog == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env["REMOTE_USER"],
			"error":     "catalog is nil",
		}).Errorf("Failed to lookup application %s", appid)

		return
	}

	insts, err := catalog.List(appid, nil)
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env["REMOTE_USER"],
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
			"namespace": r.Env["REMOTE_USER"],
			"error":     err,
		}).Warn("Failed to encode application")

		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorEncoding)
		return
	}

	routes.logger.WithFields(log.Fields{
		"namespace": r.Env["REMOTE_USER"],
	}).Infof("List application instances (%d)", len(app.Instances))
}
