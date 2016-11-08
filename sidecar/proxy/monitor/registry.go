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
	"github.com/amalgam8/amalgam8/registry/api"
)

// RegistryListener is notified of changes to the registry catalog
type RegistryListener interface {
	CatalogChange([]api.ServiceInstance) error
}

type registryMonitor struct {
	pollInterval   time.Duration
	registryClient api.ServiceDiscovery
	listeners      []RegistryListener
	ticker         *time.Ticker
	cache          map[string][]*api.ServiceInstance
	lock           sync.RWMutex
}

// RegistryConfig options
type RegistryConfig struct {
	PollInterval   time.Duration
	RegistryClient api.ServiceDiscovery
}

// RegistryMonitor Definition:
type RegistryMonitor interface {
	Monitor
	api.ServiceDiscovery
	AddListener(listener RegistryListener)
}

// NewRegistry instantiates new instance
func NewRegistry(conf RegistryConfig) RegistryMonitor {
	return &registryMonitor{
		listeners:      []RegistryListener{},
		registryClient: conf.RegistryClient,
		pollInterval:   conf.PollInterval,
	}
}

// Start monitoring registry
func (m *registryMonitor) Start() error {
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
func (m *registryMonitor) poll() error {
	// Get newest catalog from registry
	instances, err := m.registryClient.ListInstances()
	if err != nil {
		logrus.WithError(err).Warn("Could not get latest catalog from registry")
		return err
	}
	catalog := instanceListAsMap(instances)

	// Check for changes
	if m.compareToCache(catalog) {
		// Match, nothing else to do
		return nil
	}

	// Update cached catalog
	m.lock.Lock()
	m.cache = catalog
	listeners := m.listeners
	m.lock.Unlock()

	// Notify the listeners
	instancesByValue := instanceListAsValues(instances)
	for _, listener := range listeners {
		if err = listener.CatalogChange(instancesByValue); err != nil {
			logrus.WithError(err).Warn("Registry listener failed")
		}
	}

	return nil
}

// compareToCache compares the given catalog to the cached one, by comparing all instance attributes
// except for heartbeat and TTL. Instance list for each service is assumed to be presorted.
// Return 'true' if catalog match, and 'false' otherwise.
func (m *registryMonitor) compareToCache(catalog map[string][]*api.ServiceInstance) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if len(catalog) != len(m.cache) {
		return false
	}

	for name, instances := range catalog {
		cachedInstances, ok := m.cache[name]
		if !ok {
			return false
		}

		if len(instances) != len(cachedInstances) {
			return false
		}

		for i, instance := range instances {
			cachedInstance := cachedInstances[i]
			if instance.ID != cachedInstance.ID ||
				instance.ServiceName != cachedInstance.ServiceName ||
				instance.Status != cachedInstance.Status ||
				instance.Endpoint.Type != cachedInstance.Endpoint.Type ||
				instance.Endpoint.Value != cachedInstance.Endpoint.Value ||
				!reflect.DeepEqual(instance.Tags, cachedInstance.Tags) ||
				!reflect.DeepEqual(instance.Metadata, cachedInstance.Metadata) {
				return false
			}
		}
	}

	return true
}

// Stop monitoring registry
func (m *registryMonitor) Stop() error {
	// Stop ticker if necessary
	if m.ticker != nil {
		m.ticker.Stop()
		m.ticker = nil
	}

	return nil
}

// ByID sorts by ID
type ByID []*api.ServiceInstance

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

func (m *registryMonitor) ListServices() ([]string, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	keys := make([]string, 0, len(m.cache))
	for k := range m.cache {
		keys = append(keys, k)
	}

	return keys, nil
}

func (m *registryMonitor) ListInstances() ([]*api.ServiceInstance, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	// Assume 3 instances per service by average, for initial slice capacity
	// TODO: can improve based on total number of instances stored at the cache
	instances := make([]*api.ServiceInstance, 0, len(m.cache)*3)
	for _, service := range m.cache {
		instances = append(instances, service...)
	}

	return instances, nil
}

func (m *registryMonitor) ListServiceInstances(serviceName string) ([]*api.ServiceInstance, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	service := m.cache[serviceName]
	instances := make([]*api.ServiceInstance, 0, len(service))
	instances = append(instances, service...)

	return instances, nil
}

func (m *registryMonitor) AddListener(listener RegistryListener) {
	m.lock.Lock()
	copyOfListeners := m.listeners
	copyOfListeners = append(m.listeners, listener)
	m.listeners = copyOfListeners
	m.lock.Unlock()

}

func instanceListAsValues(instances []*api.ServiceInstance) []api.ServiceInstance {
	values := make([]api.ServiceInstance, len(instances))

	for i, instance := range instances {
		values[i] = *instance
	}

	return values
}

func instanceListAsMap(instances []*api.ServiceInstance) map[string][]*api.ServiceInstance {
	// Assume 3 instances per service by average, for initial map capacity
	// TODO: can improve based on length of previously stored map
	numOfServices := len(instances) / 3

	m := make(map[string][]*api.ServiceInstance, numOfServices)
	for _, instance := range instances {
		m[instance.ServiceName] = append(m[instance.ServiceName], instance)
	}

	// Sort instances of each service
	for _, value := range m {
		sort.Sort(ByID(value))
	}

	return m
}
