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

package monitor

import (
	"encoding/json"
	"reflect"
	"sort"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/registry/client"
	"github.com/amalgam8/sidecar/proxy/resources"
)

// RegistryListener is notified of changes to registry
type RegistryListener interface {
	CatalogChange(catalog resources.ServiceCatalog) error
}

type registry struct {
	pollInterval   time.Duration
	registryClient client.Discovery
	cachedCatalog  resources.ServiceCatalog
	listeners      []RegistryListener
	ticker         *time.Ticker
}

// RegistryConfig options
type RegistryConfig struct {
	PollInterval   time.Duration
	RegistryClient client.Discovery
	Listeners      []RegistryListener
}

// NewRegistry instantiates new instance
func NewRegistry(conf RegistryConfig) Monitor {
	return &registry{
		listeners:      conf.Listeners,
		registryClient: conf.RegistryClient,
		pollInterval:   conf.PollInterval,
	}
}

// Start monitoring registry
func (m *registry) Start() error {
	// Stop existing ticker if necessary
	if m.ticker != nil {
		if err := m.Stop(); err != nil {
			logrus.WithError(err).Error("Could not stop existing periodic poll")
			return err
		}
	}

	// Create new ticker
	m.ticker = time.NewTicker(m.pollInterval)

	// Do initial poll
	if err := m.poll(); err != nil {
		logrus.WithError(err).Error("Catalog check failed")
	}

	// Start periodic poll
	for range m.ticker.C {
		if err := m.poll(); err != nil {
			logrus.WithError(err).Error("Catalog check failed")
		}
	}

	return nil
}

// poll registry for changes in the catalog
func (m *registry) poll() error {
	// Get newest catalog from Registry
	latestCatalog, err := m.getLatestCatalog()
	if err != nil {
		logrus.WithError(err).Warn("Could not get latest catalog from registry")
		return err
	}

	// Check for differences
	if !m.catalogsEqual(m.cachedCatalog, latestCatalog) {
		// Update cached copy of catalog
		m.cachedCatalog = latestCatalog

		for _, listener := range m.listeners {
			if err = listener.CatalogChange(latestCatalog); err != nil {
				logrus.WithError(err).Warn("Registry listener failed")
			}
		}
	}

	return nil
}

// catalogsEqual
func (m *registry) catalogsEqual(a, b resources.ServiceCatalog) bool {
	equal := reflect.DeepEqual(a.Services, b.Services)
	logrus.WithFields(logrus.Fields{
		"a":     a,
		"b":     b,
		"equal": equal,
	}).Debug("Comparing catalogs")
	return equal
}

// getLatestCatalog
// FIXME: is this conversion still necessary?
func (m *registry) getLatestCatalog() (resources.ServiceCatalog, error) {
	catalog := resources.ServiceCatalog{}

	instances, err := m.registryClient.ListInstances(client.InstanceFilter{})
	if err != nil {
		return catalog, err
	}

	// Convert
	serviceMap := make(map[string]*resources.Service)
	for _, instance := range instances {
		if serviceMap[instance.ServiceName] == nil {
			serviceMap[instance.ServiceName] = &resources.Service{
				Name:      instance.ServiceName,
				Endpoints: []resources.Endpoint{},
			}
		}

		metadata := map[string]string{}
		err = json.Unmarshal(instance.Metadata, &metadata)

		endpoint := resources.Endpoint{
			Type:     instance.Endpoint.Type,
			Value:    instance.Endpoint.Value,
			Metadata: resources.MetaData{Version: metadata["version"]},
		}

		serviceMap[instance.ServiceName].Endpoints = append(serviceMap[instance.ServiceName].Endpoints, endpoint)
	}

	for _, service := range serviceMap {
		catalog.Services = append(catalog.Services, *service)
	}

	// Sort for comparisons (registry does not guarantee any ordering)
	sort.Sort(resources.ByService(catalog.Services))
	for _, service := range catalog.Services {
		sort.Sort(resources.ByEndpoint(service.Endpoints))
	}

	return catalog, nil
}

// Stop monitoring registry
func (m *registry) Stop() error {
	// Stop ticker if necessary
	if m.ticker != nil {
		m.ticker.Stop()
		m.ticker = nil
	}

	return nil
}
