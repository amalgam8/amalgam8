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
	cache          map[string][]*client.ServiceInstance
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
	instances, err := m.getRegistryInstances()
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
func (m *registry) compareToCache(catalog map[string][]*client.ServiceInstance) bool {
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

func (m *registry) getRegistryInstances() ([]*client.ServiceInstance, error) {
	return m.registryClient.ListInstances(client.InstanceFilter{})
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
	keys := make([]string, 0, len(m.cache))
	m.lock.RLock()
	for k := range m.cache {
		keys = append(keys, k)
	}
	m.lock.RUnlock()
	return keys, nil
}

func (m *registry) ListInstances(filter client.InstanceFilter) ([]*client.ServiceInstance, error) {
	servicesToReturn := []*client.ServiceInstance{}
	count := 0
	m.lock.RLock()
	serviceInstances := m.cache[filter.ServiceName]
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
	if instances, ok := m.cache[serviceName]; ok {
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

func instanceListAsValues(instances []*client.ServiceInstance) []client.ServiceInstance {
	values := make([]client.ServiceInstance, len(instances))

	for i, instance := range instances {
		values[i] = *instance
	}

	return values
}

func instanceListAsMap(instances []*client.ServiceInstance) map[string][]*client.ServiceInstance {
	// Assume 3 instances per service by average, for initial map capacity
	// TODO: can improve based on length of previously stored map
	numOfServices := len(instances) / 3

	m := make(map[string][]*client.ServiceInstance, numOfServices)
	for _, instance := range instances {
		m[instance.ServiceName] = append(m[instance.ServiceName], instance)
	}

	// Sort instances of each service
	for _, value := range m {
		sort.Sort(ByID(value))
	}

	return m
}
