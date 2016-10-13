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
	"sort"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/registry/client"
)

// RegistryListener is notified of changes to the registry catalog
type RegistryListener interface {
	CatalogChange([]client.ServiceInstance) error
}

type registry struct {
	pollInterval   time.Duration
	registryClient client.Discovery
	listeners      []RegistryListener
	ticker         *time.Ticker
	hashed         map[string][]*client.ServiceInstance
	lock           sync.RWMutex
}

// RegistryConfig options
type RegistryConfig struct {
	PollInterval   time.Duration
	RegistryClient client.Discovery
}

// RegistryMonitor Definition:
type RegistryMonitor interface {
	Monitor
	client.Discovery
	AddListener(listener RegistryListener)
}

// NewRegistry instantiates new instance
func NewRegistry(conf RegistryConfig) RegistryMonitor {
	return &registry{
		listeners:      []RegistryListener{},
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
	if !m.catalogsEqual(m.hashed, latestCatalog) {
		// Update cached copy of catalog
		m.lock.Lock()
		m.hashed = latestCatalog
		m.lock.Unlock()
		instances, _ := m.registryClient.ListInstances(client.InstanceFilter{})
		// Notify the listeners.
		m.lock.RLock()
		copyOfListeners := m.listeners
		m.lock.RUnlock()

		for _, listener := range copyOfListeners {
			arrayOfInstancesByValue := m.copyInstances(instances)
			if err = listener.CatalogChange(arrayOfInstancesByValue); err != nil {
				logrus.WithError(err).Warn("Registry listener failed")
			}
		}

	}
	return nil
}

func (m *registry) copyInstances(instances []*client.ServiceInstance) []client.ServiceInstance {
	arrayToReturn := make([]client.ServiceInstance, 0, len(instances))
	for i := range instances {
		arrayToReturn[i] = *instances[i]
	}
	return arrayToReturn
}

// catalogsEqual checks for pertinent differences between the given instances. We assume that all the instances have
// been presorted. Instances are compared by all values except for heartbeat and TTL.
func (m *registry) catalogsEqual(a, b map[string][]*client.ServiceInstance) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	equal := true
	if len(a) != len(b) {
		equal = false
	} else {
		for key, v := range a {
			if _, ok := b[key]; !ok {
				equal = false
				break
			}
			if len(b[key]) != len(v) {
				equal = false
			}
			for i := range v {
				if v[i].ID != b[key][i].ID || v[i].ServiceName != b[key][i].ServiceName || v[i].Status != b[key][i].Status ||
					!reflect.DeepEqual(v[i].Tags, b[key][i].Tags) || v[i].Endpoint.Type != b[key][i].Endpoint.Type ||
					v[i].Endpoint.Value != b[key][i].Endpoint.Value || !reflect.DeepEqual(v[i].Metadata, b[key][i].Metadata) {
					equal = false
					break
				}
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
func (m *registry) getLatestCatalog() (map[string][]*client.ServiceInstance, error) {
	mapToReturn := make(map[string][]*client.ServiceInstance)
	instances, err := m.registryClient.ListInstances(client.InstanceFilter{})
	if err != nil {
		return mapToReturn, err
	}

	for i := range instances {
		mapToReturn[instances[i].ServiceName] = append(mapToReturn[instances[i].ServiceName], instances[i])
	}
	for _, value := range mapToReturn {
		sort.Sort(ByID(value))
	}

	return mapToReturn, nil
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
type ByID []*client.ServiceInstance

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

func (m *registry) ListServices() ([]string, error) {
	keys := make([]string, 0, len(m.hashed))
	m.lock.RLock()
	for k := range m.hashed {
		keys = append(keys, k)
	}
	m.lock.RUnlock()
	return keys, nil
}

func (m *registry) ListInstances(filter client.InstanceFilter) ([]*client.ServiceInstance, error) {
	servicesToReturn := []*client.ServiceInstance{}
	count := 0
	m.lock.RLock()
	serviceInstances := m.hashed[filter.ServiceName]
	m.lock.RUnlock()
	for _, service := range serviceInstances {
		if service.ServiceName == filter.ServiceName {
			for _, tag := range filter.Tags {
				for _, serviceTag := range service.Tags {
					if tag == serviceTag {
						count++
					}
				}
			}
			if count == len(filter.Tags) {
				servicesToReturn = append(servicesToReturn, service)
			}
		}
	}
	return servicesToReturn, nil
}

func (m *registry) ListServiceInstances(serviceName string) ([]*client.ServiceInstance, error) {
	servicesToReturn := []*client.ServiceInstance{}
	m.lock.RLock()
	if instances, ok := m.hashed[serviceName]; ok {
		servicesToReturn = instances
	}
	m.lock.RUnlock()
	return servicesToReturn, nil
}

func (m *registry) AddListener(listener RegistryListener) {
	m.lock.Lock()
	copyOfListeners := m.listeners
	copyOfListeners = append(m.listeners, listener)
	m.listeners = copyOfListeners
	m.lock.Unlock()

}
