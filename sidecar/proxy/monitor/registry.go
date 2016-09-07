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
	"reflect"
	"time"

	"sort"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/registry/client"
)

// RegistryListener is notified of changes to the registry catalog
type RegistryListener interface {
	CatalogChange([]client.ServiceInstance) error
}

type registry struct {
	pollInterval   time.Duration
	registryClient client.Discovery
	cached         []client.ServiceInstance
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
	// Get newest catalog from registry
	latestCatalog, err := m.getLatestCatalog()
	if err != nil {
		logrus.WithError(err).Warn("Could not get latest catalog from registry")
		return err
	}

	// Check for changes
	if !m.catalogsEqual(m.cached, latestCatalog) {
		// Update cached copy of catalog
		m.cached = latestCatalog

		// Notify the listeners.
		for _, listener := range m.listeners {
			if err = listener.CatalogChange(latestCatalog); err != nil {
				logrus.WithError(err).Warn("Registry listener failed")
			}
		}
	}

	return nil
}

// catalogsEqual checks for pertinent differences between the given instances. We assume that all the instances have
// been presorted. Instances are compared by all values except for heartbeat and TTL.
func (m *registry) catalogsEqual(a, b []client.ServiceInstance) bool {
	equal := true
	if len(a) != len(b) {
		equal = false
	} else {
		for i := range a {
			if a[i].ID != b[i].ID || a[i].ServiceName != b[i].ServiceName || a[i].Status != b[i].Status ||
				!reflect.DeepEqual(a[i].Tags, b[i].Tags) || a[i].Endpoint.Type != b[i].Endpoint.Type ||
				a[i].Endpoint.Value != b[i].Endpoint.Value || !reflect.DeepEqual(a[i].Metadata, b[i].Metadata) {
				equal = false
				break
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"a":     a,
		"b":     b,
		"equal": equal,
	}).Debug("Comparing service instances")

	return equal
}

// getLatestCatalog
func (m *registry) getLatestCatalog() ([]client.ServiceInstance, error) {
	instances, err := m.registryClient.ListInstances(client.InstanceFilter{})
	if err != nil {
		return []client.ServiceInstance{}, err
	}

	// Dereference the instances.
	deref := make([]client.ServiceInstance, len(instances))
	for i := range instances {
		deref[i] = *instances[i]
	}

	sort.Sort(ByID(deref))

	return deref, nil
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

// ByID sorts by ID
type ByID []client.ServiceInstance

// Len of the array
func (a ByID) Len() int {
	return len(a)
}

// Swap i and j
func (a ByID) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Less i and j
func (a ByID) Less(i, j int) bool {
	return a[i].ID < a[j].ID
}
