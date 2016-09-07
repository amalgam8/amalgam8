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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/amalgam8/amalgam8/registry/api/env"
	"github.com/amalgam8/amalgam8/registry/api/middleware"
	"github.com/amalgam8/amalgam8/registry/store"
	"github.com/amalgam8/amalgam8/registry/utils/i18n"
	"github.com/amalgam8/amalgam8/registry/utils/reflection"
)

var instanceQueryValuesToFieldNames = make(map[string]string)

func init() {
	var si ServiceInstance
	instanceQueryValuesToFieldNames = si.GetJSONToFieldsMap()
}

func (routes *Routes) registerInstance(w rest.ResponseWriter, r *rest.Request) {
	var err error
	var req *InstanceRegistration
	req, err = routes.parseInstanceRegistrationRequest(w, r)
	if err != nil {
		return // error to client already set
	}

	catalog := routes.catalog(w, r)
	if catalog == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "catalog is nil",
		}).Errorf("Failed to register instance %+v", req)

		return
	}

	si := &store.ServiceInstance{
		ServiceName: req.ServiceName,
		Endpoint:    &store.Endpoint{Type: req.Endpoint.Type, Value: req.Endpoint.Value},
		Status:      strings.ToUpper(req.Status),
		TTL:         time.Duration(req.TTL) * time.Second,
		Metadata:    req.Metadata,
		Tags:        req.Tags}
	var sir *store.ServiceInstance

	if sir, err = catalog.Register(si); err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warnf("Failed to register instance %+v", req)

		i18n.Error(r, w, statusCodeFromError(err), i18n.ErrorInstanceRegistrationFailed)
		return
	} else if sir == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "instance is nil",
		}).Warnf("Failed to register instance %+v", req)

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

	r.Env[env.ServiceInstance] = sir
	routes.sendRegistrationResponse(w, r, sir)
}

func (routes *Routes) sendRegistrationResponse(w rest.ResponseWriter, r *rest.Request, sir *store.ServiceInstance) {
	routes.logger.WithFields(log.Fields{
		"namespace": r.Env[env.Namespace],
	}).Infof("Instance %s registered", sir)

	ttl := uint32(sir.TTL / time.Second)
	linksURL := r.BaseUrl()
	if middleware.IsUsingSecureConnection(r) { // request came in over a secure connection, continue using it
		linksURL.Scheme = "https"
	}

	links := BuildLinks(linksURL.String(), sir.ID)
	instance := &ServiceInstance{
		ID:    sir.ID,
		TTL:   ttl,
		Links: links,
	}

	w.Header().Set("Location", links.Self)
	w.WriteHeader(http.StatusCreated)

	if err := w.WriteJson(instance); err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warnf("Failed to write registration response for instance %s", sir)

		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorEncoding)
	}
}

func (routes *Routes) parseInstanceRegistrationRequest(w rest.ResponseWriter, r *rest.Request) (*InstanceRegistration, error) {
	var req InstanceRegistration
	var err error

	if err = r.DecodeJsonPayload(&req); err != nil {

		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warn("Failed to register instance")

		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorInstanceRegistrationFailed)
		return nil, err
	}

	if req.ServiceName == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "service name is required",
		}).Warnf("Failed to register instance %+v", req)

		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorServiceNameMissing)
		return nil, errors.New("Service name is required")
	}

	if req.Endpoint == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "Endpoint is required",
		}).Warnf("Failed to register instance %+v", req)

		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorInstanceEndpointMissing)
		return nil, errors.New("Endpoint is required")
	}

	if req.Endpoint.Type == "" || req.Endpoint.Value == "" {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "Endpoint type or value are missing or mismatched",
		}).Warnf("Failed to register instance %+v", req)

		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorInstanceEndpontMalformed)
		return nil, errors.New("Endpoint type or value are missing or mismatched")
	}

	switch req.Endpoint.Type {
	case EndpointTypeHTTP:
	case EndpointTypeHTTPS:
	case EndpointTypeUDP:
	case EndpointTypeTCP:
	case EndpointTypeUser:
	default:
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "Endpoint type is of invalid value",
		}).Warnf("Failed to register instance %+v", req)

		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorInstanceEndpointInvalidType)
		return nil, errors.New("Endpoint type is of invalid value")
	}

	metadataValid := true

	if req.Metadata != nil {
		metadataValid = validateJSON(req.Metadata)
	}

	if !metadataValid {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "Metadata is invalid",
		}).Warnf("Failed to register instance %+v", req)

		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorInstanceMetadataInvalid)
		return nil, errors.New("Metadata is invalid")
	}

	// If the status is not passed in, set it to UP
	if req.Status == "" {
		req.Status = store.Up
	}

	// Validate the status value
	switch strings.ToUpper(req.Status) {
	case store.Up:
	case store.Starting:
	case store.OutOfService:
	default:
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "Status field is not a valid value",
		}).Warnf("Failed to register instance %+v", req)

		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorInstanceStatusInvalid,
			map[string]interface{}{"Status": fmt.Sprintf("%s, %s, %s", store.Up, store.Starting, store.OutOfService)})
		return nil, errors.New("Status field is not a valid value")
	}

	return &req, nil
}

func validateJSON(jsonString json.RawMessage) bool {
	var js interface{}
	return json.Unmarshal(jsonString, &js) == nil
}

func (routes *Routes) deregisterInstance(w rest.ResponseWriter, r *rest.Request) {
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
		// error response set by routes.catalog()
		return
	}

	si, err := catalog.Deregister(iid)
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warnf("Failed to deregister instance %s", iid)

		i18n.Error(r, w, statusCodeFromError(err), i18n.ErrorInstanceDeletionFailed)
		return
	}

	routes.logger.WithFields(log.Fields{
		"namespace": r.Env[env.Namespace],
	}).Infof("Instance id %s deregistered", iid)

	r.Env[env.ServiceInstance] = si
	w.WriteHeader(http.StatusOK)
}

func (routes *Routes) renewInstance(w rest.ResponseWriter, r *rest.Request) {
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

		i18n.Error(r, w, statusCodeFromError(err), i18n.ErrorInstanceHeartbeatFailed)
		return
	}

	r.Env[env.ServiceInstance] = si
	w.WriteHeader(http.StatusOK)
}

func (routes *Routes) listInstances(w rest.ResponseWriter, r *rest.Request) {
	var fields []string

	fields, err := extractFields(r)
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warn("Failed to list instances")

		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorFilterBadFields)
		return
	}

	sc, err := newSelectCriteria(r)
	if err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Error("Failed to list instances")

		i18n.Error(r, w, http.StatusBadRequest, i18n.ErrorFilterSelectionCriteria)
		return
	}

	catalog := routes.catalog(w, r)
	if catalog == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "catalog is nil",
		}).Error("Failed to list instances")

		return
	}

	services := catalog.ListServices(nil)
	if services == nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     "services list is nil",
		}).Error("Failed to list instances")

		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorInstanceEnumeration)
		return
	}

	var insts = []*ServiceInstance{}
	for _, svc := range services {
		instances, err := catalog.List(svc.ServiceName, sc.instanceFilter)
		if err != nil {
			routes.logger.WithFields(log.Fields{
				"namespace": r.Env[env.Namespace],
				"error":     err,
			}).Errorf("Failed to list instances for service %s", svc.ServiceName)

			i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorInstanceEnumeration)
			return
		}

		for _, si := range instances {
			inst, err := copyInstanceWithFilter(svc.ServiceName, si, fields)
			if err != nil {
				routes.logger.WithFields(log.Fields{
					"namespace": r.Env[env.Namespace],
					"error":     err,
				}).Warn("Failed to list instances")

				i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorInstanceEnumeration)
				return
			}
			insts = append(insts, inst)
		}
	}

	if err = w.WriteJson(&InstancesList{Instances: insts}); err != nil {
		routes.logger.WithFields(log.Fields{
			"namespace": r.Env[env.Namespace],
			"error":     err,
		}).Warn("Failed to encode instances list response")

		i18n.Error(r, w, http.StatusInternalServerError, i18n.ErrorEncoding)
		return
	}

	routes.logger.WithFields(log.Fields{
		"namespace": r.Env[env.Namespace],
	}).Infof("Lookup instances (%d)", len(insts))
}

func statusCodeFromError(err error) int {
	if regerr, ok := err.(*store.Error); ok {
		switch regerr.Code {
		case store.ErrorBadRequest:
			return http.StatusBadRequest
		case store.ErrorNoSuchServiceName:
			return http.StatusNotFound
		case store.ErrorNoSuchServiceInstance:
			return http.StatusGone
		case store.ErrorNamespaceQuotaExceeded:
			return http.StatusForbidden
		case store.ErrorInternalServerError:
			return http.StatusInternalServerError
		default:
			return http.StatusInternalServerError
		}
	}
	return http.StatusInternalServerError
}

// Extract and validate filtering fields request. Note that an empty-string request is perfectly valid
func extractFields(r *rest.Request) ([]string, error) {
	if _, filteringRequested := r.URL.Query()["fields"]; !filteringRequested {
		return nil, nil
	}

	fieldsValue := r.URL.Query().Get("fields")
	if fieldsValue == "" {
		return []string{}, nil
	}

	fieldsSplit := strings.Split(fieldsValue, ",")
	fields := make([]string, len(fieldsSplit))
	for i, fld := range fieldsSplit {
		fldName, ok := instanceQueryValuesToFieldNames[fld]
		if !ok {
			return nil, fmt.Errorf("Field %s is not a valid field", fld)
		}

		fields[i] = fldName
	}

	return fields, nil
}

func copyInstanceWithFilter(sname string, si *store.ServiceInstance, fields []string) (*ServiceInstance, error) {
	inst := &ServiceInstance{
		ID:          si.ID,
		ServiceName: sname,
		Endpoint: &InstanceAddress{
			Type:  si.Endpoint.Type,
			Value: si.Endpoint.Value,
		},
		Status:        si.Status,
		Tags:          si.Tags,
		TTL:           uint32(si.TTL / time.Second),
		Metadata:      si.Metadata,
		LastHeartbeat: &si.LastRenewal,
	}

	if fields != nil {
		filteredInstance := &ServiceInstance{}
		err := reflection.FilterStructByFields(inst, filteredInstance, fields)
		if err != nil {
			return nil, err
		}
		// Add Endpoint because it should always be returned to the user
		filteredInstance.Endpoint = &InstanceAddress{Type: inst.Endpoint.Type, Value: inst.Endpoint.Value}
		filteredInstance.ID = inst.ID
		return filteredInstance, nil
	}

	return inst, nil
}
