package eureka

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/amalgam8/registry/utils/i18n"
)

func (routes *Routes) listVips(w rest.ResponseWriter, r *rest.Request) {

	vip := r.PathParam(RouterParamVip)
	if vip == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env["REMOTE_USER"],
			"error":     "vip is required",
		}).Warn("Failed to list vip")

		i18n.Error(r, w, http.StatusBadRequest, i18n.EurekaErrorVIPRequired)
		return
	}

	catalog := routes.catalog(w, r)
	if catalog == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env["REMOTE_USER"],
			"error":     "catalog is nil",
		}).Errorf("Failed to list vip %s", vip)

		return
	}

	services := catalog.ListServices(nil)
	if services == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env["REMOTE_USER"],
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
				"namespace": r.Env["REMOTE_USER"],
				"error":     err,
			}).Errorf("Failed to list vips %s", vip)

			i18n.Error(r, w, http.StatusInternalServerError, i18n.EurekaErrorVIPEnumeration)
			return
		}

		app := &Application{Name: svc.ServiceName, Instances: make([]*Instance, 0, len(insts))}
		for _, inst := range insts {
			if vipaddr, ok := inst.Extension["VipAddress"]; ok {
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
			"namespace": r.Env["REMOTE_USER"],
			"error":     err,
		}).Warn("Failed to encode vips list")

		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorEncoding)
		return
	}

	routes.logger.WithFields(log.Fields{
		"namespace": r.Env["REMOTE_USER"],
	}).Infof("List vips (%d apps, %d insts)", len(apps.Application), instsCount)
}
