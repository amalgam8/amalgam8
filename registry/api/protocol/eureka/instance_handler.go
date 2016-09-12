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

	"github.com/amalgam8/amalgam8/registry/api/env"
	"github.com/amalgam8/amalgam8/registry/store"
	"github.com/amalgam8/amalgam8/registry/utils/i18n"
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

	if reg.Inst == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "instance is required",
		}).Warn("Failed to register instance")

		i18n.Error(r, w, http.StatusBadRequest, i18n.EurekaErrorRequiredFieldsMissing)
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

	if appid != inst.Application {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "application name mismatch",
		}).Warnf("Failed to register instance %+v", inst)

		i18n.Error(r, w, http.StatusBadRequest, i18n.EurekaErrorApplicationMismatch)
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
	// In the old eureka client versions (1.1.x) the instance ID is NOT explicitly set in the request data,
	// but it is part of the DatacenterInfo class.
	iid := inst.ID
	if iid == "" {
		iid, err = getInstanceID(inst)
		if err != nil {
			routes.logger.WithFields(log.Fields{
				"namespace": r.Env[env.Namespace],
				"error":     err,
			}).Warnf("Failed to register instance %+v", inst)

			i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorInstanceRegistrationFailed)
			return
		}
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

		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorNamespaceNotFound)
		return
	}

	ext, err := buildExtensionFromInstance(inst)
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Errorf("Failed to register instance %+v", inst)

		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorEncoding)
		return
	}

	si := &store.ServiceInstance{
		ID:          iid,
		ServiceName: inst.Application,
		Endpoint:    &store.Endpoint{Type: "tcp", Value: fmt.Sprintf("%s:%v", inst.IPAddr, inst.Port.Value)},
		Status:      inst.Status,
		TTL:         time.Duration(ttl) * time.Second,
		Metadata:    inst.Metadata,
		Extension:   ext}

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

	si, err := catalog.Deregister(iid)
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warnf("Failed to deregister instance %s", iid)

		i18n.Error(r, w, http.StatusGone, i18n.ErrorInstanceNotFound)
		return
	}

	routes.logger.WithFields(log.Fields{
		"namespace": r.Env[env.Namespace],
	}).Infof("Instance id %s deregistered", iid)

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

	catalog := routes.catalog(w, r)
	if catalog == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "catalog is nil",
		}).Errorf("Failed to renew instance %s", iid)

		return
	}

	si, err := catalog.Renew(iid)
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warnf("Failed to renew instance %s", iid)

		i18n.Error(r, w, http.StatusGone, i18n.ErrorInstanceNotFound)
		return
	}

	r.Env[env.ServiceInstance] = si
	w.WriteHeader(http.StatusOK)
}

func (routes *Routes) getInstance(w rest.ResponseWriter, r *rest.Request) {
	iid := r.PathParam(RouteParamInstanceID)
	if iid == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "instance id is required",
		}).Warn("Failed to query instance")

		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorInstanceIdentifierMissing)
		return
	}

	catalog := routes.catalog(w, r)
	if catalog == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "catalog is nil",
		}).Errorf("Failed to renew instance %s", iid)

		return
	}

	si, err := catalog.Instance(iid)
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warnf("Failed to query instance %s", iid)

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
		}).Errorf("Failed to set instance %s status", iid)

		return
	}

	si, err := catalog.Instance(iid)
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warnf("Failed to set instance %s status", iid)

		i18n.Error(r, w, http.StatusNotFound, i18n.ErrorInstanceNotFound)
		return
	}

	if si.ServiceName != appid {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "Application id does not match",
		}).Warnf("Failed to set instance %s status. service_name: %s", iid, si.ServiceName)

		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorInstanceNotFound)
		return
	}

	si, err = catalog.SetStatus(iid, status)
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warnf("Failed to set instance %s status", iid)

		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorInstanceNotFound)
		return
	}

	routes.logger.WithFields(log.Fields{
		"namespace": r.Env[env.Namespace],
	}).Infof("Instance %s status was changed. old: %s, new: %s", iid, si.Status, status)

	r.Env[env.ServiceInstance] = si
	w.WriteHeader(http.StatusOK)
}

func buildExtensionFromInstance(inst *Instance) (map[string]interface{}, error) {
	copyInst := *inst

	// Clear all fields that we want to override later
	copyInst.Status = ""
	copyInst.OvrStatus = ""
	copyInst.ActionType = ""
	copyInst.Metadata = nil
	copyInst.LastUpdatedTs = nil
	copyInst.LastDirtyTs = nil
	if copyInst.Lease != nil {
		copyInst.Lease.RegistrationTs = 0
		copyInst.Lease.LastRenewalTs = 0
	}

	ext, err := json.Marshal(copyInst)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{extEureka: string(ext), extVIP: copyInst.VIPAddr}, nil
}

func buildInstanceFromRegistry(si *store.ServiceInstance) *Instance {
	inst := buildDefaultInstance(si)

	if si.Extension != nil {
		if ext, ok := si.Extension[extEureka]; ok {
			json.Unmarshal([]byte(ext.(string)), &inst)
		}
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
		Metadata:      si.Metadata,
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
