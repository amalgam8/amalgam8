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

package manager

import (
	"net/http"

	"time"

	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/controller/resources"
	"github.com/pborman/uuid"
)

// Manager client
type Manager interface {
	Create(id, token string, rules resources.TenantInfo) error
	Set(id string, rules resources.TenantInfo) error
	Get(id string) (resources.TenantEntry, error)
	Delete(id string) error

	SetVersion(id string, version resources.Version) error
	DeleteVersion(id, service string) error
	GetVersion(id, service string) (resources.Version, error)

	AddRules(id string, filters []resources.Rule) error
	ListRules(id string, filterIDs []string) ([]resources.Rule, error)
	//UpdateFilters(id string, filters []resources.Rule) error
	DeleteRules(id string, filterIDs []string) error
}

type manager struct {
	db database.Tenant
}

// Config options
type Config struct {
	Database database.Tenant
}

// NewManager creates Manager instance
func NewManager(conf Config) Manager {
	return &manager{
		db: conf.Database,
	}
}

func (m *manager) Create(id, token string, tenantInfo resources.TenantInfo) error {
	var err error

	entry := resources.TenantEntry{
		BasicEntry: resources.BasicEntry{
			ID: id,
		},
		TenantToken: token,
		ProxyConfig: resources.ProxyConfig{
			LoadBalance: tenantInfo.LoadBalance,
			Filters:     tenantInfo.Filters,
		},
		ServiceCatalog: resources.ServiceCatalog{
			Services:   []resources.Service{},
			LastUpdate: time.Now(),
		},
	}

	// Copy each element

	if entry.ProxyConfig.Filters.Rules == nil {
		entry.ProxyConfig.Filters.Rules = []resources.Rule{}
	}

	if entry.ProxyConfig.Filters.Versions == nil {
		entry.ProxyConfig.Filters.Versions = []resources.Version{}
	}

	// Set defaults if necessary
	if entry.ProxyConfig.LoadBalance == "" {
		entry.ProxyConfig.LoadBalance = "round_robin" // FIXME: common location for this?
	}

	for _, rule := range entry.ProxyConfig.Filters.Rules {
		if err = validateRule(rule); err != nil {
			return err
		}
	}

	rules := []resources.Rule{}
	for _, rule := range entry.ProxyConfig.Filters.Rules {
		if rule.DelayProbability == 0.0 && rule.AbortProbability == 0.0 {
			continue
		}
		rules = append(rules, rule)
	}
	entry.ProxyConfig.Filters.Rules = rules

	// Add to rules
	if err = m.db.Create(entry); err != nil {
		logrus.WithError(err).Error("Failed setting rules")
		return &ServiceUnavailableError{Reason: "Could not create entry", ErrorMessage: "FIXME", Err: err}
	}

	return nil
}

// Set database entry
func (m *manager) Set(id string, tenantInfo resources.TenantInfo) error {
	var err error

	// TODO: only read and set proxyconfig if necessary
	entry, err := m.db.Read(id)
	if err != nil {
		//handleDBError(w, req, err)
		return &DBError{Err: err}
	}

	if tenantInfo.LoadBalance != "" {
		entry.ProxyConfig.LoadBalance = tenantInfo.LoadBalance
	}

	if tenantInfo.Filters.Rules != nil {
		for _, rule := range tenantInfo.Filters.Rules {
			if err = validateRule(rule); err != nil {
				return err
			}
		}

		rules := []resources.Rule{}
		for _, rule := range tenantInfo.Filters.Rules {
			if rule.DelayProbability == 0.0 && rule.AbortProbability == 0.0 {
				continue
			}
			rules = append(rules, rule)
		}
		entry.ProxyConfig.Filters.Rules = rules
	}

	if tenantInfo.Filters.Versions != nil {
		//TODO validate fields
		entry.ProxyConfig.Filters.Versions = tenantInfo.Filters.Versions
	}

	if err = m.updateProxyConfig(entry); err != nil {
		logrus.WithFields(logrus.Fields{
			"err":       err,
			"tenant_id": id,
			//"request_id": reqID,
		}).Error("Error updating info for tenant ID")
		return &ServiceUnavailableError{Reason: "database update failed", ErrorMessage: "database_fail", Err: err}
	}

	return nil
}

// Get database entry
func (m *manager) Get(id string) (resources.TenantEntry, error) {
	entry, err := m.db.Read(id)
	if err != nil {
		return entry, &DBError{Err: err}
	}
	return entry, nil
}

// Delete database entry
func (m *manager) Delete(id string) error {
	if err := m.db.Delete(id); err != nil {
		return &DBError{Err: err}
	}
	return nil
}

func (m *manager) SetVersion(id string, newVersion resources.Version) error {
	entry, err := m.db.Read(id)
	if err != nil {
		return &DBError{Err: err}

	}

	updateIndex := -1
	for index, version := range entry.ProxyConfig.Filters.Versions {
		if version.Service == newVersion.Service {
			updateIndex = index
			break
		}
	}
	if updateIndex == -1 {
		entry.ProxyConfig.Filters.Versions = append(entry.ProxyConfig.Filters.Versions, newVersion)
	} else {
		entry.ProxyConfig.Filters.Versions[updateIndex] = newVersion
	}

	// Update the entry in the database
	if err = m.updateProxyConfig(entry); err != nil {

		logrus.WithFields(logrus.Fields{
			"err":       err,
			"tenant_id": id,
			//"request_id": reqID,
		}).Error("Error updating info for tenant ID")
		return &ServiceUnavailableError{Reason: "database update failed", ErrorMessage: "database_fail", Err: err}
	}

	return nil
}

func (m *manager) DeleteVersion(id, service string) error {

	entry, err := m.db.Read(id)
	if err != nil {
		return &DBError{Err: err}

	}

	updateIndex := -1
	for index, version := range entry.ProxyConfig.Filters.Versions {
		if version.Service == service {
			updateIndex = index
			break
		}
	}
	if updateIndex == -1 {
		logrus.Error(fmt.Sprintf("No registered service(s) for %v matching service name %v", id, service))
		return &RuleNotFoundError{Reason: "No registered service(s) matching name", ErrorMessage: "invalid_service"}
	}

	entry.ProxyConfig.Filters.Versions = append(entry.ProxyConfig.Filters.Versions[:updateIndex], entry.ProxyConfig.Filters.Versions[updateIndex+1:]...)

	// Update the entry in the database
	if err = m.updateProxyConfig(entry); err != nil {
		logrus.WithFields(logrus.Fields{
			"err":       err,
			"tenant_id": id,
			//"request_id": reqID,
		}).Error("Error updating info for tenant ID")
		return &ServiceUnavailableError{Reason: "database update failed", ErrorMessage: "database_fail", Err: err}
	}

	return nil
}

func (m *manager) GetVersion(id, service string) (resources.Version, error) {
	entry, err := m.db.Read(id)
	if err != nil {
		return resources.Version{}, &DBError{Err: err}

	}

	for _, version := range entry.ProxyConfig.Filters.Versions {
		if version.Service == service {
			return version, nil
		}
	}

	logrus.Error(fmt.Sprintf("No registered service(s) for %v matching service name %v", id, service))

	return resources.Version{}, &RuleNotFoundError{Reason: "No registered service(s) matching name", ErrorMessage: "invalid_service"}
}

func validateRule(rule resources.Rule) error {
	// Validate source
	if rule.Source == "" {
		return &InvalidRuleError{Reason: "invalid source", ErrorMessage: "invalid_source"}
	}

	// Validate destination
	if rule.Destination == "" {
		return &InvalidRuleError{Reason: "invalid destination", ErrorMessage: "invalid_destination"}
	}

	// Validate abort
	if rule.AbortProbability < 0.0 || rule.AbortProbability > 1.0 {
		return &InvalidRuleError{Reason: "invalid abort probability", ErrorMessage: "invalid_abort_probability"}
	}

	if rule.AbortProbability > 0.0 && (rule.ReturnCode < -1 || rule.ReturnCode == 0 || rule.ReturnCode >= 600) {
		return &InvalidRuleError{Reason: "invalid return code", ErrorMessage: "invalid_return_code"}
	}

	// Validate delay
	if rule.DelayProbability < 0.0 || rule.DelayProbability > 1.0 {
		return &InvalidRuleError{Reason: "invalid delay probability", ErrorMessage: "invalid_delay_probability"}
	}

	if rule.Delay < 0 || rule.Delay > 600 {
		return &InvalidRuleError{Reason: "invalid delay", ErrorMessage: "invalid_delay"}
	}

	if rule.DelayProbability > 0.0 && rule.Delay <= 0.0 {
		return &InvalidRuleError{Reason: "invalid delay", ErrorMessage: "invalid_delay"}
	}

	// Validate header
	if rule.Header == "" {
		return &InvalidRuleError{Reason: "invalid header", ErrorMessage: "invalid_header"}
	}

	// Validate header value
	if rule.Pattern == "" {
		return &InvalidRuleError{Reason: "invalid header pattern", ErrorMessage: "invalid_header_pattern"}
	}

	return nil
}

func (m *manager) updateProxyConfig(entry resources.TenantEntry) error {

	if err := m.updateTenant(entry); err != nil {
		if ce, ok := err.(*database.DBError); ok {
			if ce.StatusCode == http.StatusConflict {
				newerEntry, err := m.db.Read(entry.ID)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"err": err,
						"id":  entry.ID,
					}).Error("Failed to retrieve latest document during conflict resolution")
					return err
				}

				newerEntry.ProxyConfig = entry.ProxyConfig

				if err = m.updateTenant(entry); err != nil {
					logrus.WithFields(logrus.Fields{
						"err": err,
						"id":  entry.ID,
					}).Error("Failed to resolve document update conflict")
					return err
				}
				logrus.WithFields(logrus.Fields{
					"id": entry.ID,
				}).Debug("Succesfully resolved document update conflict")
				return nil
			}
			logrus.WithFields(logrus.Fields{
				"err": err,
				"id":  entry.ID,
			}).Error("Database error attempting to update proxy config")
			return err

		}
		logrus.WithFields(logrus.Fields{
			"err": err,
			"id":  entry.ID,
		}).Error("Failed attempting to update proxy config")
		return err
	}

	return nil
}

// InvalidRuleIndexError describes an error involving a rule at a index
type InvalidRuleIndexError struct {
	Index       int
	Description string
	Filter      resources.Rule
}

// InvalidRulesError list of rule errors
type InvalidRulesError []InvalidRuleIndexError

// Error description
func (e *InvalidRulesError) Error() string {
	return "manager: invalid filters"
}

// RulesNotFoundError describes 1..N rules that were not found
type RulesNotFoundError struct {
	IDs []string
}

// Error description
func (e *RulesNotFoundError) Error() string {
	return "manager: not found"
}

// AddRules to a tenant.
func (m *manager) AddRules(id string, filters []resources.Rule) error {

	// Validate filters
	invalidFilters := make([]InvalidRuleIndexError, 0, len(filters))
	for i, filter := range filters {
		if err := validateRule(filter); err != nil {
			filterErr := InvalidRuleIndexError{
				Index:       i,
				Description: "bad_filter",
				Filter:      filter,
			}
			invalidFilters = append(invalidFilters, filterErr)
		}
	}

	if len(invalidFilters) > 0 {
		err := InvalidRulesError(invalidFilters)
		return &err
	}

	entry, err := m.db.Read(id)
	if err != nil {
		return err
	}

	// Generate IDs
	for i := 0; i < len(filters); i++ {
		filters[i].ID = uuid.New()
	}

	// Add to the existing filters
	entry.ProxyConfig.Filters.Rules = append(entry.ProxyConfig.Filters.Rules, filters...)

	// Write the results
	if err = m.updateTenant(entry); err != nil {
		// TODO: handle database conflict errors by re-reading the document and re-attempting the operation?
		return err
	}

	return nil
}

// ListRules with the specified IDs in the indicated tenant. If no filter IDs are provided, all filters are listed.
func (m *manager) ListRules(id string, filterIDs []string) ([]resources.Rule, error) {
	entry, err := m.db.Read(id)
	if err != nil {
		return []resources.Rule{}, err
	}

	filters := make([]resources.Rule, 0, len(filterIDs))
	if len(filterIDs) == 0 {
		filters = entry.ProxyConfig.Filters.Rules
	} else {
		filterMap := make(map[string]resources.Rule)
		for _, filter := range entry.ProxyConfig.Filters.Rules {
			filterMap[filter.ID] = filter
		}

		missingIDs := make([]string, 0, len(filterIDs))
		for _, filterID := range filterIDs {
			filter, exists := filterMap[filterID]
			if exists {
				filters = append(filters, filter)
			} else {
				missingIDs = append(missingIDs, filterID)
			}
		}

		if len(missingIDs) > 0 {
			err := &RulesNotFoundError{
				IDs: missingIDs,
			}

			return []resources.Rule{}, err
		}
	}

	return filters, nil
}

// DeleteRules for a tenant. If no filter IDs are provided, all filters are deleted.
func (m *manager) DeleteRules(id string, filterIDs []string) error {
	entry, err := m.db.Read(id)
	if err != nil {
		return err
	}

	if len(filterIDs) == 0 {
		entry.ProxyConfig.Filters.Rules = []resources.Rule{}
	} else {
		filterMap := make(map[string]bool)
		for _, filterID := range filterIDs {
			filterMap[filterID] = false
		}

		filters := make([]resources.Rule, 0, len(entry.ProxyConfig.Filters.Rules))
		for _, filter := range entry.ProxyConfig.Filters.Rules {
			_, exists := filterMap[filter.ID]
			if exists {
				filterMap[filter.ID] = true
			} else {
				filters = append(filters, filter)
			}
		}

		missingIDs := make([]string, 0, len(filterIDs))
		for _, filterID := range filterIDs {
			deleted := filterMap[filterID]
			if !deleted {
				missingIDs = append(missingIDs, filterID)
			}
		}

		if len(missingIDs) > 0 {
			err := &RulesNotFoundError{
				IDs: missingIDs,
			}

			return err
		}

		entry.ProxyConfig.Filters.Rules = filters
	}

	// Write the results
	if err = m.updateTenant(entry); err != nil {
		// TODO: handle database conflict errors by re-reading the document and re-attempting the operation?
		return err
	}

	return nil
}

func (m *manager) updateTenant(tenant resources.TenantEntry) error {
	// Update last update time
	tenant.ServiceCatalog.LastUpdate = time.Now()
	return m.db.Update(tenant)
}
