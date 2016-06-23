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

package proxyconfig

import (
	"net/http"

	"errors"

	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/controller/notification"
	"github.com/amalgam8/controller/resources"
	"github.com/pborman/uuid"
)

// Manager client
type Manager interface {
	Set(rules resources.ProxyConfig) error
	Get(id string) (resources.ProxyConfig, error)
	Delete(id string) error

	AddFilters(id string, filters []resources.Rule) error
	ListFilters(id string, filterIDs []string) ([]resources.Rule, error)
	//UpdateFilters(id string, filters []resources.Rule) error
	DeleteFilters(id string, filterIDs []string) error
}

type InvalidFilterError struct {
	Index       int
	Description string
	Filter      resources.Rule
}

type InvalidFiltersError []InvalidFilterError

func (e *InvalidFiltersError) Error() string {
	return "proxyconfig: invalid filters"
}

type FiltersNotFoundError struct {
	IDs []string
}

func (e *FiltersNotFoundError) Error() string {
	return "proxyconfig: not found"
}

type manager struct {
	db            database.Rules
	producerCache notification.TenantProducerCache
}

// Config options
type Config struct {
	Database      database.Rules
	ProducerCache notification.TenantProducerCache
}

// NewManager creates Manager instance
func NewManager(conf Config) Manager {
	return &manager{
		db:            conf.Database,
		producerCache: conf.ProducerCache,
	}
}

func validateFilter(filter resources.Rule) error {
	if filter.Destination == "" {
		return errors.New("invalid destination")
	}

	if filter.AbortProbability < 0.0 || filter.AbortProbability > 1.0 {
		return errors.New("invalid abort probability")
	}

	if filter.ReturnCode < 0 || filter.ReturnCode >= 600 {
		return errors.New("invalid return code")
	}

	if filter.DelayProbability < 0.0 || filter.DelayProbability > 1.0 {
		return errors.New("invalid probability")
	}

	if filter.Delay < 0 || filter.Delay > 600 {
		return errors.New("invalid duration")
	}

	if (filter.DelayProbability != 0.0 && filter.Delay == 0.0) || (filter.DelayProbability == 0.0 && filter.Delay != 0.0) {
		return errors.New("invalid delay")
	}

	return nil
}

// AddFilters to a tenant.
func (m *manager) AddFilters(id string, filters []resources.Rule) error {
	// TODO: ensure we are operating on a non-orphan

	// Validate filters
	invalidFilters := make([]InvalidFilterError, 0, len(filters))
	for i, filter := range filters {
		if err := validateFilter(filter); err != nil {
			filterErr := InvalidFilterError{
				Index:       i,
				Description: "bad_filter",
				Filter:      filter,
			}
			invalidFilters = append(invalidFilters, filterErr)
		}
	}

	if len(invalidFilters) > 0 {
		err := InvalidFiltersError(invalidFilters)
		return &err
	}

	conf, err := m.db.Read(id)
	if err != nil {
		return err
	}

	// Generate IDs
	for i := 0; i < len(filters); i++ {
		filters[i].ID = uuid.New()
	}

	// Add to the existing filters
	conf.Filters.Rules = append(conf.Filters.Rules, filters...)

	// Write the results
	if err = m.db.Update(conf); err != nil {
		// TODO: handle database conflict errors by re-reading the document and re-attempting the operation?
		return err
	}

	// Notify of changes
	if err = m.producerCache.SendEvent(id, conf.Credentials.Kafka); err != nil {
		return err
	}

	return nil
}

// ListFilters with the specified IDs in the indicated tenant. If no filter IDs are provided, all filters are listed.
func (m *manager) ListFilters(id string, filterIDs []string) ([]resources.Rule, error) {
	// TODO: make sure we aren't operating on an orphan?

	conf, err := m.db.Read(id)
	if err != nil {
		return []resources.Rule{}, err
	}

	filters := make([]resources.Rule, 0, len(filterIDs))
	if len(filterIDs) == 0 {
		filters = conf.Filters.Rules
	} else {
		filterMap := make(map[string]resources.Rule)
		for _, filter := range conf.Filters.Rules {
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
			err := &FiltersNotFoundError{
				IDs: missingIDs,
			}

			return []resources.Rule{}, err
		}
	}

	return filters, nil
}

// DeleteFilters for a tenant. If no filter IDs are provided, all filters are deleted.
func (m *manager) DeleteFilters(id string, filterIDs []string) error {
	conf, err := m.db.Read(id)
	if err != nil {
		return err
	}

	if len(filterIDs) == 0 {
		conf.Filters.Rules = []resources.Rule{}
	} else {
		filterMap := make(map[string]bool)
		for _, filterID := range filterIDs {
			filterMap[filterID] = false
		}

		filters := make([]resources.Rule, 0, len(conf.Filters.Rules))
		for _, filter := range conf.Filters.Rules {
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
			err := &FiltersNotFoundError{
				IDs: missingIDs,
			}

			return err
		}

		conf.Filters.Rules = filters
	}

	// Write the results
	if err = m.db.Update(conf); err != nil {
		// TODO: handle database conflict errors by re-reading the document and re-attempting the operation?
		return err
	}

	// Notify of changes
	if err = m.producerCache.SendEvent(id, conf.Credentials.Kafka); err != nil {
		return err
	}

	return nil
}

// Set database entry
func (m *manager) Set(rules resources.ProxyConfig) error {
	var err error
	if err := m.validate(rules); err != nil {
		return err
	}

	if rules.Rev == "" {
		err = m.db.Create(rules)
	} else {
		err = m.db.Update(rules)
	}

	if err != nil {
		if ce, ok := err.(*database.DBError); ok {
			if ce.StatusCode == http.StatusConflict {
				// There is an old orphan entry in the database, delete it and create a new entry
				oldRules, err := m.db.Read(rules.ID)
				if err != nil {
					return err
				}

				rules.Rev = oldRules.Rev

				if err = m.db.Update(rules); err != nil {
					return err
				}
			} else {
				return err
			}

		} else {
			return err
		}
	}

	// Send Kafka event
	if err = m.producerCache.SendEvent(rules.ID, rules.Credentials.Kafka); err != nil {
		return err
	}

	return nil
}

// Get database entry
func (m *manager) Get(id string) (resources.ProxyConfig, error) {
	return m.db.Read(id)
}

// Delete database entry
func (m *manager) Delete(id string) error {
	return m.db.Delete(id)
}

func (m *manager) validate(config resources.ProxyConfig) error {
	return nil
}
