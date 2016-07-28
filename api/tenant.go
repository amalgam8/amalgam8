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

package api

import (
	"errors"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/manager"
	"github.com/amalgam8/controller/metrics"
	"github.com/amalgam8/controller/middleware"
	"github.com/amalgam8/controller/resources"
	"github.com/ant0ine/go-json-rest/rest"
)

// Tenant handles tenant API calls
type Tenant struct {
	reporter metrics.Reporter
	manager  manager.Manager
}

// TenantConfig options
type TenantConfig struct {
	Reporter metrics.Reporter
	Manager  manager.Manager
}

// NewTenant instantiates instance of the API
func NewTenant(conf TenantConfig) *Tenant {
	return &Tenant{
		reporter: conf.Reporter,
		manager:  conf.Manager,
	}
}

// Routes for tenant API calls
func (t *Tenant) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Post("/v1/tenants", reportMetric(t.reporter, t.PostTenant, "tenants_create")),
		rest.Put("/v1/tenants", reportMetric(t.reporter, t.PutTenant, "tenants_update")),
		rest.Get("/v1/tenants", reportMetric(t.reporter, t.GetTenant, "tenants_read")),
		rest.Delete("/v1/tenants", reportMetric(t.reporter, t.DeleteTenant, "tenants_delete")),

		rest.Put("/v1/versions/#service", reportMetric(t.reporter, t.PutServiceVersions, "versions_update")),
		rest.Get("/v1/versions/#service", reportMetric(t.reporter, t.GetServiceVersions, "versions_read")),
		rest.Delete("/v1/versions/#service", reportMetric(t.reporter, t.DeleteServiceVersions, "versions_delete")),

		//rest.Get("/v1/rules", reportMetric(t.reporter, t.GetRules, "rules_read")),
		//rest.Delete("/v1/rules", reportMetric(t.reporter, t.DeleteRules, "rules_delete")),
		//rest.Post("/v1/rules", reportMetric(t.reporter, t.PostRules, "rules_create")),
		//rest.Put("/v1/rules", reportMetric(t.reporter, t.PutRules, "rules_update")),
	}
}

func handleRuleError(w rest.ResponseWriter, req *rest.Request, err error) {
	if invalidRulesErr, ok := err.(*manager.InvalidRulesError); ok {
		// TODO: in the case of a PUT, we may want to indicate the invalid filters by ID instead of index
		restErrors := make([]TranslatableError, 0, len(*invalidRulesErr))
		for _, filterErr := range *invalidRulesErr {
			restErrors = append(restErrors, TranslatableError{
				Index: filterErr.Index,
				Error: "bad_filter",
			})
		}

		WriteRestErrors(w, req, restErrors, http.StatusBadRequest)
		return
	} else if rulesNotFoundErr, ok := err.(*manager.RulesNotFoundError); ok {
		restErrors := make([]TranslatableError, 0, len(rulesNotFoundErr.IDs))
		for _, ID := range rulesNotFoundErr.IDs {
			restErrors = append(restErrors, TranslatableError{
				ID:    ID,
				Error: "not_found",
			})
		}

		WriteRestErrors(w, req, restErrors, http.StatusNotFound)
		return
	}

	// TODO: handle 409s, 500s, 502s, 503s

	RestError(w, req, http.StatusInternalServerError, err.Error()) //"unknown_error")
}

// PostRules creates filters in bulk
func (t *Tenant) PostRules(w rest.ResponseWriter, req *rest.Request) error {
	id := GetTenantID(req)
	if id == "" {
		RestError(w, req, http.StatusBadRequest, "no_tenant_id")
		return errors.New("no tenant id")
	}

	filtersJSON := struct {
		Filters []resources.Rule `json:"rules"`
	}{}

	err := req.DecodeJsonPayload(&filtersJSON)
	if err != nil {
		RestError(w, req, http.StatusBadRequest, "json_error")
		return err
	}

	if err = t.manager.AddRules(id, filtersJSON.Filters); err != nil {
		handleRuleError(w, req, err)
		return err
	}

	w.WriteHeader(http.StatusCreated)
	w.WriteJson(&filtersJSON)

	return nil
}

// PutRules updates filters in bulk
func (t *Tenant) PutRules(w rest.ResponseWriter, req *rest.Request) error {
	w.WriteHeader(http.StatusNotImplemented)
	return nil
}

// GetRules reads filters in bulk
func (t *Tenant) GetRules(w rest.ResponseWriter, req *rest.Request) error {
	id := GetTenantID(req)
	if id == "" {
		RestError(w, req, http.StatusBadRequest, "no_tenant_id")
		return errors.New("no tenant id")
	}
	ruleIDs := getQueryIDs("id", req)

	rules, err := t.manager.ListRules(id, ruleIDs)
	if err != nil {
		handleRuleError(w, req, err)
		return err
	}

	respJSON := struct {
		Rules []resources.Rule `json:"rules"`
	}{
		Rules: rules,
	}

	w.WriteHeader(http.StatusOK)
	w.WriteJson(&respJSON)

	return nil
}

// DeleteRules deletes filters in bulk
func (t *Tenant) DeleteRules(w rest.ResponseWriter, req *rest.Request) error {
	id := GetTenantID(req)
	if id == "" {
		RestError(w, req, http.StatusBadRequest, "no_tenant_id")
		return errors.New("no tenant id")
	}
	ruleIDs := getQueryIDs("id", req)

	if err := t.manager.DeleteRules(id, ruleIDs); err != nil {
		handleRuleError(w, req, err)
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

// PostTenant initializes a tenant in the Controller
func (t *Tenant) PostTenant(w rest.ResponseWriter, req *rest.Request) error {
	var err error

	tenantID := GetTenantID(req)
	if tenantID == "" {
		RestError(w, req, http.StatusBadRequest, "error_invalid_input")
		return errors.New("special error")
	}

	tenantInfo := resources.TenantInfo{}

	tenantToken := req.Header.Get(middleware.AuthHeader)

	if err = req.DecodeJsonPayload(&tenantInfo); err != nil {
		RestError(w, req, http.StatusBadRequest, "json_error")
		return err
	}

	// Validate input
	if tenantID == "" {
		RestError(w, req, http.StatusBadRequest, "error_invalid_input")
		return errors.New("special error")
	}

	if err = t.manager.Create(tenantID, tenantToken, tenantInfo); err != nil {
		processError(w, req, err)
		return err
	}

	w.WriteHeader(http.StatusCreated)
	return nil
}

// PutTenant updates credentials and/or metadata for a tenant
func (t *Tenant) PutTenant(w rest.ResponseWriter, req *rest.Request) error {
	var err error

	tenantID := GetTenantID(req)
	if tenantID == "" {
		RestError(w, req, http.StatusBadRequest, "error_invalid_input")
		return errors.New("special error")
	}

	tenantInfo := resources.TenantInfo{}

	if err = req.DecodeJsonPayload(&tenantInfo); err != nil {
		RestError(w, req, http.StatusBadRequest, "json_error")
		return err
	}

	if err = t.manager.Set(tenantID, tenantInfo); err != nil {
		processError(w, req, err)
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

// GetTenant returns credentials and metadata for a tenant
func (t *Tenant) GetTenant(w rest.ResponseWriter, req *rest.Request) error {
	// validate auth header
	// if this tenant has orphans, CSB will know that the token is invalid

	tenantID := GetTenantID(req)
	if tenantID == "" {
		RestError(w, req, http.StatusBadRequest, "error_invalid_input")
		return errors.New("special error")
	}

	entry, err := t.manager.Get(tenantID)
	if err != nil {
		processError(w, req, err)
		return err
	}

	tenantInfo := resources.TenantInfo{
		Credentials: entry.ProxyConfig.Credentials,
		LoadBalance: entry.ProxyConfig.LoadBalance,
		Filters:     entry.ProxyConfig.Filters,
	}

	w.WriteHeader(http.StatusOK)
	w.WriteJson(&tenantInfo)
	return nil
}

// GetServiceVersions returns versioning info for a service of a tenant
func (t *Tenant) GetServiceVersions(w rest.ResponseWriter, req *rest.Request) error {
	reqID := req.Header.Get(middleware.RequestIDHeader)

	tenantID := GetTenantID(req)
	if tenantID == "" {
		RestError(w, req, http.StatusBadRequest, "error_invalid_input")
		return errors.New("special error")
	}
	service := req.PathParam("service")

	respJSON, err := t.manager.GetVersion(tenantID, service)
	if err != nil {
		processError(w, req, err)
		return err
	}

	err = w.WriteJson(respJSON)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":        err,
			"tenant_id":  tenantID,
			"service":    service,
			"request_id": reqID,
		}).Warn("Could not write JSON response for getting version information")
		return err
	}
	return nil
}

// PutServiceVersions adds versioning info for a service of a tenant
func (t *Tenant) PutServiceVersions(w rest.ResponseWriter, req *rest.Request) error {
	reqID := req.Header.Get(middleware.RequestIDHeader)

	tenantID := GetTenantID(req)
	if tenantID == "" {
		RestError(w, req, http.StatusBadRequest, "error_invalid_input")
		return errors.New("special error")
	}
	service := req.PathParam("service")

	newVersion := resources.Version{}
	if err := req.DecodeJsonPayload(&newVersion); err != nil {
		logrus.WithFields(logrus.Fields{
			"tenant_id":  tenantID,
			"request_id": reqID,
			"service":    service,
			"err":        err,
		}).Error("Could not parse JSON")
		RestError(w, req, http.StatusBadRequest, "json_error")
		return err
	}
	newVersion.Service = service

	if err := t.manager.SetVersion(tenantID, newVersion); err != nil {
		processError(w, req, err)
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

// DeleteServiceVersions deletes versioning info for a service of a tenant
func (t *Tenant) DeleteServiceVersions(w rest.ResponseWriter, req *rest.Request) error {
	//reqID := req.Header.Get(middleware.RequestIDHeader)

	tenantID := GetTenantID(req)
	if tenantID == "" {
		RestError(w, req, http.StatusBadRequest, "error_invalid_input")
		return errors.New("special error")
	}
	service := req.PathParam("service")

	if err := t.manager.DeleteVersion(tenantID, service); err != nil {
		if err != nil {
			processError(w, req, err)
			return err
		}
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

// DeleteTenant removes tenant from Controller
func (t *Tenant) DeleteTenant(w rest.ResponseWriter, req *rest.Request) error {
	var err error

	tenantID := GetTenantID(req)
	if tenantID == "" {
		RestError(w, req, http.StatusBadRequest, "error_invalid_input")
		return errors.New("special error")
	}

	// Delete from rules
	if err = t.manager.Delete(tenantID); err != nil {
		processError(w, req, err)
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func processError(w rest.ResponseWriter, req *rest.Request, err error) {
	if err != nil {
		tenantID := GetTenantID(req)
		requestID := req.Header.Get(middleware.RequestIDHeader)

		log := logrus.WithFields(logrus.Fields{
			"err":        err,
			"tenant_id":  tenantID,
			"request_id": requestID,
		})
		if e, ok := err.(*manager.InvalidRuleError); ok {
			log.Error("Bad request")
			RestError(w, req, http.StatusBadRequest, e.ErrorMessage)
		} else if e, ok := err.(*manager.DBError); ok {
			log.Error("Database error occured")
			handleDBReadError(w, req, e.Err)
		} else if e, ok := err.(*manager.ServiceUnavailableError); ok {
			log.Error("Service unavailable")
			RestError(w, req, http.StatusServiceUnavailable, e.ErrorMessage)
		} else if e, ok := err.(*manager.RuleNotFoundError); ok {
			log.Error("Rule ID not found")
			RestError(w, req, http.StatusNotFound, e.ErrorMessage)
		} else {
			log.Error("Unknown availability error occured")
			RestError(w, req, http.StatusServiceUnavailable, "unknown_availability_error")
		}
	}
}
