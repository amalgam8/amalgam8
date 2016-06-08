package api

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/checker"
	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/controller/middleware"
	"github.com/amalgam8/controller/proxyconfig"
	"github.com/amalgam8/controller/resources"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/cactus/go-statsd-client/statsd"
	"net/http"
)

// Tenant handles tenant API calls
type Tenant struct {
	statsd  statsd.Statter
	catalog checker.Checker
	rules   proxyconfig.Manager
}

// TenantConfig options
type TenantConfig struct {
	Statsd      statsd.Statter
	Checker     checker.Checker
	ProxyConfig proxyconfig.Manager
}

// TenantInfo JSON object for credentials and metadata of a tenant
type TenantInfo struct {
	ID                string                `json:"id"`
	Credentials       resources.Credentials `json:"credentials"`
	LoadBalance       string                `json:"load_balance"`
	Port              int                   `json:"port"`
	ReqTrackingHeader string                `json:"req_tracking_header"`
	Filters           resources.Filters     `json:"filters"`
}

// NewTenant creates struct
func NewTenant(conf TenantConfig) *Tenant {
	return &Tenant{
		statsd:  conf.Statsd,
		catalog: conf.Checker,
		rules:   conf.ProxyConfig,
	}
}

// Routes for tenant API calls
func (t *Tenant) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Post("/v1/tenants", ReportMetric(t.statsd, t.PostTenant, "tenants_create")),
		rest.Put("/v1/tenants/#id", ReportMetric(t.statsd, t.PutTenant, "tenants_update")),
		rest.Get("/v1/tenants/#id", ReportMetric(t.statsd, t.GetTenant, "tenants_read")),
		rest.Delete("/v1/tenants/#id", ReportMetric(t.statsd, t.DeleteTenant, "tenants_delete")),
		rest.Put("/v1/tenants/#id/versions/#service", ReportMetric(t.statsd, t.PutServiceVersions, "versions_update")),
		rest.Get("/v1/tenants/#id/versions/#service", ReportMetric(t.statsd, t.GetServiceVersions, "versions_read")),
		rest.Delete("/v1/tenants/#id/versions/#service", ReportMetric(t.statsd, t.DeleteServiceVersions, "versions_update")),
	}
}

// PostTenant initializes a tenant in the Controller
func (t *Tenant) PostTenant(w rest.ResponseWriter, req *rest.Request) error {
	var err error

	tenantConf := TenantInfo{}

	if err = req.DecodeJsonPayload(&tenantConf); err != nil {
		RestError(w, req, http.StatusBadRequest, "json_error")
		return err
	}

	// Validate input
	if tenantConf.ID == "" {
		RestError(w, req, http.StatusBadRequest, "error_invalid_input")
		return errors.New("special error")
	}

	// Copy each element
	proxyConf := resources.ProxyConfig{
		BasicEntry: resources.BasicEntry{
			ID: tenantConf.ID,
		},
		LoadBalance:       tenantConf.LoadBalance,
		Port:              tenantConf.Port,
		ReqTrackingHeader: tenantConf.ReqTrackingHeader,
		Filters:           tenantConf.Filters,
	}

	if proxyConf.Filters.Rules == nil {
		proxyConf.Filters.Rules = []resources.Rule{}
	}

	if proxyConf.Filters.Versions == nil {
		proxyConf.Filters.Versions = []resources.Version{}
	}

	// Set defaults if necessary
	if proxyConf.LoadBalance == "" {
		proxyConf.LoadBalance = "round_robin" // FIXME: common location for this?
	}

	if proxyConf.Port == 0 {
		proxyConf.Port = 6379 // FIXME
	}

	if proxyConf.ReqTrackingHeader == "" {
		proxyConf.ReqTrackingHeader = "X-Request-ID" // FIXME: common location for this?
	}

	if err = validateRules(w, req, proxyConf.Filters.Rules); err != nil {
		// RestError() called in validate function
		return err
	}

	rules := []resources.Rule{}
	for _, rule := range proxyConf.Filters.Rules {
		if rule.DelayProbability == 0.0 && rule.AbortProbability == 0.0 {
			continue
		}
		rules = append(rules, rule)
	}
	proxyConf.Filters.Rules = rules

	proxyConf.Credentials = resources.Credentials{
		Kafka:    tenantConf.Credentials.Kafka,
		Registry: tenantConf.Credentials.Registry,
	}

	// Ensure Registry credentials are provided
	if proxyConf.Credentials.Registry.URL == "" || proxyConf.Credentials.Registry.Token == "" {
		RestError(w, req, http.StatusBadRequest, "must provide Registry creds")
		return errors.New("must provide Registry creds")
	}

	mhCredValid := false

	if !proxyConf.Credentials.Kafka.SASL && len(proxyConf.Credentials.Kafka.Brokers) != 0 &&
		proxyConf.Credentials.Kafka.APIKey == "" &&
		proxyConf.Credentials.Kafka.AdminURL == "" &&
		proxyConf.Credentials.Kafka.RestURL == "" &&
		proxyConf.Credentials.Kafka.Password == "" &&
		proxyConf.Credentials.Kafka.User == "" {

		// local kafka case
		mhCredValid = true
	} else if proxyConf.Credentials.Kafka.SASL && proxyConf.Credentials.Kafka.APIKey != "" &&
		proxyConf.Credentials.Kafka.AdminURL != "" &&
		len(proxyConf.Credentials.Kafka.Brokers) != 0 &&
		proxyConf.Credentials.Kafka.RestURL != "" &&
		proxyConf.Credentials.Kafka.Password != "" &&
		proxyConf.Credentials.Kafka.User != "" {

		// Bluemix Message Hub case
		mhCredValid = true
	} else if !proxyConf.Credentials.Kafka.SASL && len(proxyConf.Credentials.Kafka.Brokers) == 0 &&
		proxyConf.Credentials.Kafka.APIKey == "" &&
		proxyConf.Credentials.Kafka.AdminURL == "" &&
		proxyConf.Credentials.Kafka.RestURL == "" &&
		proxyConf.Credentials.Kafka.Password == "" &&
		proxyConf.Credentials.Kafka.User == "" {

		// no kafka messaging used
		mhCredValid = true
	}

	if !mhCredValid {
		RestError(w, req, http.StatusBadRequest, "must provide all Kafka creds")
		return errors.New("bad Kafka credentials")
	}

	// TODO: perform a check to ensure that the SD and MH credentials actually work?

	// Add to rules
	if err = t.rules.Set(proxyConf); err != nil {
		logrus.WithError(err).Error("Failed setting rules")
		//TODO return 500 internal server error?
		RestError(w, req, http.StatusServiceUnavailable, "could not set rules")
		return err
	}

	// Register with catalog
	if err = t.catalog.Register(tenantConf.ID); err != nil {
		logrus.WithError(err).Error("Failed registering with catalog")
		if ce, ok := err.(*database.DBError); ok {
			if ce.StatusCode == http.StatusConflict {
				// FIXME if already present, creds and rules have just been overwritten
				RestError(w, req, http.StatusConflict, "already_exists")
				return err
			}
			RestError(w, req, http.StatusServiceUnavailable, "database_error")
			return err
		}
		RestError(w, req, http.StatusServiceUnavailable, "service_unavailable")
		return err
	}

	w.WriteHeader(http.StatusCreated)
	return nil
}

func validateRules(w rest.ResponseWriter, req *rest.Request, filters []resources.Rule) error {
	for _, filter := range filters {

		if filter.Destination == "" {
			RestError(w, req, http.StatusBadRequest, "invalid_destination")
			return errors.New("invalid destination")
		}

		if filter.AbortProbability < 0.0 || filter.AbortProbability > 1.0 {
			RestError(w, req, http.StatusBadRequest, "invalid_abort_probability")
			return errors.New("invalid abort probability")
		}

		if filter.ReturnCode < 0 || filter.ReturnCode >= 600 {
			RestError(w, req, http.StatusBadRequest, "invalid_return_code")
			return errors.New("invalid return code")
		}

		if filter.DelayProbability < 0.0 || filter.DelayProbability > 1.0 {
			RestError(w, req, http.StatusBadRequest, "invalid_delay_probability")
			return errors.New("invalid probability")
		}

		if filter.Delay < 0 || filter.Delay > 600 {
			RestError(w, req, http.StatusBadRequest, "invalid_delay")
			return errors.New("invalid duration")
		}

		if (filter.DelayProbability != 0.0 && filter.Delay == 0.0) || (filter.DelayProbability == 0.0 && filter.Delay != 0.0) {
			RestError(w, req, http.StatusBadRequest, "invalid_delay")
			return errors.New("invalid delay")
		}

		// if filter.Header == "" {
		// 	filter.Header = "X-Filter-Header"
		// }

		if filter.Pattern == "" {
			filter.Pattern = "*"
		}

	}

	return nil
}

// PutTenant updates credentials and/or metadata for a tenant
// TODO: if an update succeeds for one, but not the other we end up partially updating the state
func (t *Tenant) PutTenant(w rest.ResponseWriter, req *rest.Request) error {
	var err error

	id := req.PathParam("id")

	tenantConf := TenantInfo{}

	if err = req.DecodeJsonPayload(&tenantConf); err != nil {
		RestError(w, req, http.StatusBadRequest, "json_error")
		return err
	}

	// Only allow changes to registered tenants
	_, err = t.catalog.Get(id)
	if err != nil {
		RestError(w, req, http.StatusNotFound, "not_registered")
		return err
	}

	setRegistry := false
	setKafka := false

	if tenantConf.Credentials.Registry.URL != "" && tenantConf.Credentials.Registry.Token != "" {
		setRegistry = true
	} else if tenantConf.Credentials.Registry.URL != "" || tenantConf.Credentials.Registry.Token != "" {
		RestError(w, req, http.StatusBadRequest, "bad Registry credentials")
		return errors.New("bad Registry credentials")
	}

	if tenantConf.Credentials.Kafka.APIKey != "" &&
		tenantConf.Credentials.Kafka.AdminURL != "" &&
		len(tenantConf.Credentials.Kafka.Brokers) != 0 &&
		tenantConf.Credentials.Kafka.RestURL != "" &&
		tenantConf.Credentials.Kafka.Password != "" &&
		tenantConf.Credentials.Kafka.User != "" {
		setKafka = true
	} else if tenantConf.Credentials.Kafka.APIKey == "" &&
		tenantConf.Credentials.Kafka.AdminURL == "" &&
		len(tenantConf.Credentials.Kafka.Brokers) != 0 &&
		tenantConf.Credentials.Kafka.RestURL == "" &&
		tenantConf.Credentials.Kafka.User == "" &&
		tenantConf.Credentials.Kafka.Password == "" &&
		!tenantConf.Credentials.Kafka.SASL {
		setKafka = true
	} else if tenantConf.Credentials.Kafka.APIKey != "" ||
		tenantConf.Credentials.Kafka.AdminURL != "" ||
		len(tenantConf.Credentials.Kafka.Brokers) != 0 ||
		tenantConf.Credentials.Kafka.RestURL != "" ||
		tenantConf.Credentials.Kafka.Password != "" ||
		tenantConf.Credentials.Kafka.User != "" {
		RestError(w, req, http.StatusBadRequest, "")
		return errors.New("bad Kafka credentials")
	}

	// TODO: only read and set proxyconfig if necessary
	proxyConf, err := t.rules.Get(id)
	if err != nil {
		RestError(w, req, http.StatusServiceUnavailable, "get_proxy_conf_failed")
		return err
	}

	if setRegistry || setKafka {
		// TODO: perform a check to ensure that the Registry and Kafka credentials actually work?

		if setRegistry {
			proxyConf.Credentials.Registry = tenantConf.Credentials.Registry
		}

		if setKafka {
			proxyConf.Credentials.Kafka = tenantConf.Credentials.Kafka
		}
	}

	if tenantConf.LoadBalance != "" {
		proxyConf.LoadBalance = tenantConf.LoadBalance
	}

	if tenantConf.Port > 0 {
		proxyConf.Port = tenantConf.Port
	}

	if tenantConf.ReqTrackingHeader != "" {
		proxyConf.ReqTrackingHeader = tenantConf.ReqTrackingHeader
	}

	if tenantConf.Filters.Rules != nil {
		if err = validateRules(w, req, tenantConf.Filters.Rules); err != nil {
			return err
		}

		rules := []resources.Rule{}
		for _, rule := range tenantConf.Filters.Rules {
			if rule.DelayProbability == 0.0 && rule.AbortProbability == 0.0 {
				continue
			}
			rules = append(rules, rule)
		}
		proxyConf.Filters.Rules = rules
	}

	if tenantConf.Filters.Versions != nil {
		//TODO validate fields
		proxyConf.Filters.Versions = tenantConf.Filters.Versions
	}

	if err = t.rules.Set(proxyConf); err != nil {
		RestError(w, req, http.StatusServiceUnavailable, "set_proxy_conf_failed")
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

// GetTenant returns credentials and metadata for a tenant
func (t *Tenant) GetTenant(w rest.ResponseWriter, req *rest.Request) error {
	// validate auth header
	// if this tenant has orphans, CSB will know that the token is invalid

	id := req.PathParam("id")

	_, err := t.catalog.Get(id)
	if err != nil {
		if ce, ok := err.(*database.DBError); ok {
			if ce.StatusCode == http.StatusNotFound {
				RestError(w, req, http.StatusNotFound, "no matching id")
				return err
			}
			RestError(w, req, http.StatusServiceUnavailable, "rules_database_error")
			return err
		}
		RestError(w, req, http.StatusServiceUnavailable, "get_rules_failed")
		return err
	}

	proxyConfig, err := t.rules.Get(id)
	if err != nil {
		if ce, ok := err.(*database.DBError); ok {
			if ce.StatusCode == http.StatusNotFound {
				RestError(w, req, http.StatusNotFound, "no matching id")
				return err
			}
			RestError(w, req, http.StatusServiceUnavailable, "rules_database_error")
			return err
		}
		RestError(w, req, http.StatusServiceUnavailable, "get_rules_failed")
		return err
	}

	tenantConf := TenantInfo{
		ID:                id,
		Credentials:       proxyConfig.Credentials,
		LoadBalance:       proxyConfig.LoadBalance,
		Port:              proxyConfig.Port,
		ReqTrackingHeader: proxyConfig.ReqTrackingHeader,
		Filters:           proxyConfig.Filters,
	}

	w.WriteHeader(http.StatusOK)
	w.WriteJson(&tenantConf)
	return nil
}

// GetServiceVersions returns versioning info for a service of a tenant
func (t *Tenant) GetServiceVersions(w rest.ResponseWriter, req *rest.Request) error {
	reqID := req.Header.Get(middleware.RequestIDHeader)

	tenantID := req.PathParam("id")
	service := req.PathParam("service")

	proxyConfig, err := t.rules.Get(tenantID)
	if err != nil {
		RestError(w, req, http.StatusServiceUnavailable, "database_fail")
		return err
	}

	var respJSON *resources.Version
	for _, versions := range proxyConfig.Filters.Versions {
		if versions.Service == service {
			respJSON = &versions
			break
		}
	}

	if respJSON == nil {
		RestError(w, req, http.StatusNotFound, "invalid_service")
		return fmt.Errorf("No registered service(s) for %v matching service name %v", tenantID, service)
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

	tenantID := req.PathParam("id")
	service := req.PathParam("service")

	proxyConfig, err := t.rules.Get(tenantID)
	if err != nil {
		RestError(w, req, http.StatusServiceUnavailable, "database_fail")
		return err
	}

	newVersion := resources.Version{}
	err = req.DecodeJsonPayload(&newVersion)
	if err != nil {
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

	updateIndex := -1
	for index, version := range proxyConfig.Filters.Versions {
		if version.Service == service {
			updateIndex = index
			break
		}
	}
	if updateIndex == -1 {
		proxyConfig.Filters.Versions = append(proxyConfig.Filters.Versions, newVersion)
	} else {
		proxyConfig.Filters.Versions[updateIndex] = newVersion
	}

	// Update the entry in the database
	err = t.rules.Set(proxyConfig)
	if err != nil {
		if _, ok := err.(*database.DBError); ok {
			logrus.WithFields(logrus.Fields{
				"err":        err,
				"tenant_id":  tenantID,
				"request_id": reqID,
			}).Error("Error updating info for tenant ID")
			RestError(w, req, http.StatusServiceUnavailable, "database_fail")

		} else {
			logrus.WithFields(logrus.Fields{
				"err":        err,
				"tenant_id":  tenantID,
				"request_id": reqID,
			}).Error("Error updating info for tenant ID")
			RestError(w, req, http.StatusServiceUnavailable, "database_error")
		}

		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

// DeleteServiceVersions deletes versioning info for a service of a tenant
func (t *Tenant) DeleteServiceVersions(w rest.ResponseWriter, req *rest.Request) error {
	reqID := req.Header.Get(middleware.RequestIDHeader)

	tenantID := req.PathParam("id")
	service := req.PathParam("service")

	proxyConfig, err := t.rules.Get(tenantID)
	if err != nil {
		RestError(w, req, http.StatusServiceUnavailable, "database_fail")
		return err
	}

	updateIndex := -1
	for index, version := range proxyConfig.Filters.Versions {
		if version.Service == service {
			updateIndex = index
			break
		}
	}
	if updateIndex == -1 {
		RestError(w, req, http.StatusNotFound, "invalid_service")
		return fmt.Errorf("No registered service(s) for %v matching service name %v", tenantID, service)
	} else {
		proxyConfig.Filters.Versions = append(proxyConfig.Filters.Versions[:updateIndex], proxyConfig.Filters.Versions[updateIndex+1:]...)
	}

	// Update the entry in the database
	err = t.rules.Set(proxyConfig)
	if err != nil {
		if _, ok := err.(*database.DBError); ok {
			logrus.WithFields(logrus.Fields{
				"err":        err,
				"tenant_id":  tenantID,
				"request_id": reqID,
			}).Error("Error updating info for tenant ID")
			RestError(w, req, http.StatusServiceUnavailable, "database_fail")

		} else {
			logrus.WithFields(logrus.Fields{
				"err":        err,
				"tenant_id":  tenantID,
				"request_id": reqID,
			}).Error("Error updating info for tenant ID")
			RestError(w, req, http.StatusServiceUnavailable, "database_error")
		}

		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

// DeleteTenant removes tenant from Controller
func (t *Tenant) DeleteTenant(w rest.ResponseWriter, req *rest.Request) error {
	var err error

	id := req.PathParam("id")

	// Deregister from catalog
	if err = t.catalog.Deregister(id); err != nil {
		logrus.WithError(err).Error("Could not deregister tenant")
		if ce, ok := err.(*database.DBError); ok {
			if ce.StatusCode == http.StatusNotFound {
				RestError(w, req, http.StatusNotFound, "no matching registered id")
				return err
			}
			RestError(w, req, http.StatusServiceUnavailable, "database_error")
			return err
		}
		RestError(w, req, http.StatusServiceUnavailable, "error_deregister_failed")
		return err
	}

	// Delete from rules
	if err = t.rules.Delete(id); err != nil {
		logrus.WithError(err).Warn("Rule deletion failed, document orphaned")
		// TODO do anything else here
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
