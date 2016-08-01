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
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ant0ine/go-json-rest/rest"

	"strings"

	"github.com/amalgam8/registry/api/env"
	"github.com/amalgam8/registry/store"
	"github.com/amalgam8/registry/utils/i18n"
)

func (routes *Routes) registerInstance(w rest.ResponseWriter, r *rest.Request) {
	var err error
	var reg InstanceWrapper

	appid := r.PathParam(RouteParamAppID)
	if appid == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "application id is required",
		}).Warn("Failed to register instance")

		i18n.Error(r, w, http.StatusBadRequest, i18n.EurekaErrorApplicationIdentifierMissing)
		return
	}

	if err = r.DecodeJsonPayload(&reg); err != nil {

		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warn("Failed to register instance")

		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorInstanceRegistrationFailed)
		return
	}

	inst := reg.Inst
	if inst.Application == "" {
		inst.Application = appid
	}

	if inst.HostName == "" || inst.Application == "" || inst.VIPAddr == "" || inst.IPAddr == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "hostname, application, vipaddress and IPaddress are required",
		}).Warnf("Failed to register instance %+v", inst)

		i18n.Error(r, w, http.StatusBadRequest, i18n.EurekaErrorRequiredFieldsMissing)
		return
	}

	metadataValid := true

	if inst.Metadata != nil {
		metadataValid = validateJSON(inst.Metadata)
	}

	if !metadataValid {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "metadata is not valid",
		}).Warnf("Failed to register instance %+v", inst)

		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorInstanceMetadataInvalid)
		return
	}

	// Get the instance ID
	iid := inst.ID
	if iid == "" {
		iid = inst.HostName
	}

	ttl := defaultDurationInt
	if inst.Lease != nil && inst.Lease.DurationInt > 0 {
		ttl = inst.Lease.DurationInt
	}

	catalog := routes.catalog(w, r)
	if catalog == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "catalog is nil",
		}).Errorf("Failed to register instance %+v", inst)

		return
	}

	si := &store.ServiceInstance{
		ID:          buildUniqueInstanceID(appid, iid),
		ServiceName: inst.Application,
		Endpoint:    &store.Endpoint{Type: "tcp", Value: fmt.Sprintf("%s:%v", inst.IPAddr, inst.Port.Value)},
		Status:      inst.Status,
		TTL:         time.Duration(ttl) * time.Second,
		Metadata:    inst.Metadata,
		Extension:   buildExtensionFromInstance(inst)}

	var sir *store.ServiceInstance

	if sir, err = catalog.Register(si); err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warnf("Failed to register instance %+v", inst)

		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorInstanceRegistrationFailed)
		return
	} else if sir == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "instance is nil",
		}).Warnf("Failed to register instance %+v", inst)

		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorNilObject)
		return
	} else if sir.ID == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "instance id is empty",
		}).Warnf("Failed to register instance %s", sir)

		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorInstanceIdentifierMissing)
		return
	}

	routes.logger.WithFields(log.Fields{
		"namespace": r.Env[env.Namespace],
	}).Infof("Instance %s registered", sir)

	r.Env[env.ServiceInstance] = sir
	w.WriteHeader(http.StatusNoContent)
}

func validateJSON(jsonString json.RawMessage) bool {
	var js interface{}
	return json.Unmarshal(jsonString, &js) == nil
}

func (routes *Routes) deregisterInstance(w rest.ResponseWriter, r *rest.Request) {
	appid := r.PathParam(RouteParamAppID)
	if appid == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "application id is required",
		}).Warn("Failed to deregister instance")

		i18n.Error(r, w, http.StatusBadRequest, i18n.EurekaErrorApplicationIdentifierMissing)
		return
	}

	iid := r.PathParam(RouteParamInstanceID)
	if iid == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "instance id is required",
		}).Warn("Failed to deregister instance")

		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorInstanceIdentifierMissing)
		return
	}

	catalog := routes.catalog(w, r)
	if catalog == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "catalog is nil",
		}).Errorf("Failed to deregister instance %s", iid)

		return
	}

	uid := buildUniqueInstanceID(appid, iid)
	si, err := catalog.Deregister(uid)
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warnf("Failed to deregister instance %s", uid)

		i18n.Error(r, w, http.StatusGone, i18n.ErrorInstanceNotFound)
		return
	}

	routes.logger.WithFields(log.Fields{
		"namespace": r.Env[env.Namespace],
	}).Infof("Instance id %s deregistered", uid)

	r.Env[env.ServiceInstance] = si
	w.WriteHeader(http.StatusOK)
}

func (routes *Routes) renewInstance(w rest.ResponseWriter, r *rest.Request) {
	appid := r.PathParam(RouteParamAppID)
	if appid == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "application id is required",
		}).Warn("Failed to renew instance")

		i18n.Error(r, w, http.StatusBadRequest, i18n.EurekaErrorApplicationIdentifierMissing)
		return
	}

	iid := r.PathParam(RouteParamInstanceID)
	if iid == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "instance id is required",
		}).Warn("Failed to renew instance")

		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorInstanceIdentifierMissing)
		return
	}
	uid := buildUniqueInstanceID(appid, iid)

	catalog := routes.catalog(w, r)
	if catalog == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "catalog is nil",
		}).Errorf("Failed to renew instance %s", iid)

		return
	}

	si, err := catalog.Renew(uid)
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warnf("Failed to renew instance %s", uid)

		i18n.Error(r, w, http.StatusGone, i18n.ErrorInstanceNotFound)
		return
	}

	r.Env[env.ServiceInstance] = si
	w.WriteHeader(http.StatusOK)
}

func (routes *Routes) getInstance(w rest.ResponseWriter, r *rest.Request) {
	appid := r.PathParam(RouteParamAppID)
	if appid == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "application id is required",
		}).Warn("Failed to query instancee")

		i18n.Error(r, w, http.StatusBadRequest, i18n.EurekaErrorApplicationIdentifierMissing)
		return
	}

	iid := r.PathParam(RouteParamInstanceID)
	if iid == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "instance id is required",
		}).Warn("Failed to query instance")

		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorInstanceIdentifierMissing)
		return
	}
	uid := buildUniqueInstanceID(appid, iid)

	catalog := routes.catalog(w, r)
	if catalog == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "catalog is nil",
		}).Errorf("Failed to renew instance %s", iid)

		return
	}

	si, err := catalog.Instance(uid)
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warnf("Failed to query instance %s", uid)

		i18n.Error(r, w, http.StatusNotFound, i18n.ErrorInstanceNotFound)
		return
	}

	r.Env[env.ServiceInstance] = si
	inst := buildInstanceFromRegistry(si)
	err = w.WriteJson(&InstanceWrapper{inst})
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warn("Failed to encode instance")

		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorEncoding)
		return
	}
}

func (routes *Routes) setStatus(w rest.ResponseWriter, r *rest.Request) {
	appid := r.PathParam(RouteParamAppID)
	if appid == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "application id is required",
		}).Warn("Failed to set instances status")

		i18n.Error(r, w, http.StatusBadRequest, i18n.EurekaErrorApplicationIdentifierMissing)
		return
	}

	iid := r.PathParam(RouteParamInstanceID)
	if iid == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "instance id is required",
		}).Warn("Failed to set instances status")

		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorInstanceIdentifierMissing)
		return
	}
	uid := buildUniqueInstanceID(appid, iid)

	status := r.URL.Query().Get("value")
	if status == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "status value is required",
		}).Warn("Failed to set instances status")

		i18n.Error(r, w, http.StatusBadRequest, i18n.EurekaErrorStatusMissing)
		return
	}

	catalog := routes.catalog(w, r)
	if catalog == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "catalog is nil",
		}).Errorf("Failed to set instance %s status", uid)

		return
	}

	si, err := catalog.Instance(uid)
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warnf("Failed to set instance %s status", uid)

		i18n.Error(r, w, http.StatusNotFound, i18n.ErrorInstanceNotFound)
		return
	}

	if si.ServiceName != appid {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "Application id does not match",
		}).Warnf("Failed to set instance %s status. service_name: %s", uid, si.ServiceName)

		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorInstanceNotFound)
		return
	}

	si, err = catalog.SetStatus(uid, status)
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warnf("Failed to set instance %s status", uid)

		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorInstanceNotFound)
		return
	}

	routes.logger.WithFields(log.Fields{
		"namespace": r.Env[env.Namespace],
	}).Infof("Instance %s status was changed. old: %s, new: %s", uid, si.Status, status)

	r.Env[env.ServiceInstance] = si
	w.WriteHeader(http.StatusOK)
}

func buildUniqueInstanceID(appid, iid string) string {
	return fmt.Sprintf("%s.%s", appid, iid)
}

func buildExtensionFromInstance(inst *Instance) map[string]interface{} {
	extension := map[string]interface{}{
		"HostName":   inst.HostName,
		"VipAddress": inst.VIPAddr,
		"IPAddr":     inst.IPAddr,
		"Port":       inst.Port,
		"CountryId":  inst.CountryID,
	}
	if inst.ID != "" {
		extension["ID"] = inst.ID
	}
	if inst.GroupName != "" {
		extension["GroupName"] = inst.GroupName
	}
	if inst.SecVIPAddr != "" {
		extension["SecVIPAddr"] = inst.SecVIPAddr
	}
	if inst.SecPort != nil {
		extension["SecPort"] = inst.SecPort
	}
	if inst.HomePage != "" {
		extension["HomePage"] = inst.HomePage
	}
	if inst.StatusPage != "" {
		extension["StatusPage"] = inst.StatusPage
	}
	if inst.HealthCheck != "" {
		extension["HealthCheck"] = inst.HealthCheck
	}
	if inst.Datacenter != nil {
		extension["Datacenter"] = inst.Datacenter
	}
	if inst.Lease != nil {
		extension["Lease"] = inst.Lease
	}
	return extension
}

func buildInstanceFromRegistry(si *store.ServiceInstance) *Instance {
	inst := buildDefaultInstance(si)

	if si.Extension != nil {
		inst.HostName = si.Extension["HostName"].(string)
		inst.IPAddr = si.Extension["VipAddress"].(string)
		inst.IPAddr = si.Extension["IPAddr"].(string)
		inst.Port = si.Extension["Port"].(*Port)
		inst.CountryID = si.Extension["CountryId"].(int)

		if value, ok := si.Extension["ID"]; ok {
			inst.ID = value.(string)
		}
		if value, ok := si.Extension["GroupName"]; ok {
			inst.GroupName = value.(string)
		}
		if value, ok := si.Extension["SecVIPAddr"]; ok {
			inst.SecVIPAddr = value.(string)
		}
		if value, ok := si.Extension["SecPort"]; ok {
			inst.SecPort = value.(*Port)
		}
		if value, ok := si.Extension["HomePage"]; ok {
			inst.HomePage = value.(string)
		}
		if value, ok := si.Extension["StatusPage"]; ok {
			inst.StatusPage = value.(string)
		}
		if value, ok := si.Extension["HealthCheck"]; ok {
			inst.HealthCheck = value.(string)
		}
		if value, ok := si.Extension["Datacenter"]; ok {
			inst.Datacenter = value.(*DatacenterInfo)
		}
		if value, ok := si.Extension["Lease"]; ok {
			li := value.(*LeaseInfo)
			inst.Lease.RenewalInt = li.RenewalInt
		}
		inst.Metadata = si.Metadata
	}

	return inst
}

func buildDefaultInstance(si *store.ServiceInstance) *Instance {
	inst := &Instance{
		Application: si.ServiceName,
		VIPAddr:     si.ServiceName,
		GroupName:   "UNKNOWN",
		Status:      si.Status,
		Datacenter: &DatacenterInfo{
			Class: "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo",
			Name:  "MyOwn",
		},
		Lease: &LeaseInfo{
			RegistrationTs: si.RegistrationTime.Unix(),
			DurationInt:    uint32(si.TTL / time.Second),
			LastRenewalTs:  si.LastRenewal.Unix(),
		},
		CountryID:     1,
		CordServer:    "false",
		ActionType:    "ADDED",
		OvrStatus:     "UNKNOWN",
		LastDirtyTs:   fmt.Sprintf("%d", si.RegistrationTime.Unix()),
		LastUpdatedTs: fmt.Sprintf("%d", si.RegistrationTime.Unix()),
	}

	if si.Endpoint != nil && len(si.Endpoint.Value) > 0 {
		pos := strings.LastIndex(si.Endpoint.Value, ":")
		if pos > -1 {
			inst.HostName = si.Endpoint.Value[:pos]
			inst.Port = &Port{Enabled: "true", Value: si.Endpoint.Value[pos+1:]}
		} else {
			inst.HostName = si.Endpoint.Value
		}
		inst.IPAddr = inst.HostName
	}

	return inst
}
