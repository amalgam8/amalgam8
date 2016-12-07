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
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/pkg/api"
)

// DiscoveryListener is notified of changes to the discovery catalog
type DiscoveryListener interface {
	CatalogChange([]api.ServiceInstance) error
}

// DiscoveryConfig holds configuration options for the discovery monitor
type DiscoveryConfig struct {
	Discovery    api.ServiceDiscovery
	Listeners    []DiscoveryListener
	PollInterval time.Duration
}

// DiscoveryMonitor interface.
type DiscoveryMonitor interface {
	Monitor
	Listeners() []DiscoveryListener
	SetListeners(listeners []DiscoveryListener)
}

type discoveryMonitor struct {
	discovery api.ServiceDiscovery

	ticker       *time.Ticker
	pollInterval time.Duration

	cache     map[string][]*api.ServiceInstance
	listeners []DiscoveryListener
}

// DefaultDiscoveryPollInterval is the default used for the discovery monitor's poll interval,
// if no other value is specified. Currently, all existing ServiceDiscovery adapters use caching
// with background polling, so the 1 second polling here is basically polling a local cache only.
// Note: this will be removed once the ServiceDiscovery interface exposes a Watch() mechanism.
const DefaultDiscoveryPollInterval = 1 * time.Second

// NewDiscoveryMonitor instantiates a new discovery monitor
func NewDiscoveryMonitor(conf DiscoveryConfig) DiscoveryMonitor {
	if conf.PollInterval == 0 {
		conf.PollInterval = DefaultDiscoveryPollInterval
	}

	return &discoveryMonitor{
		discovery:    conf.Discovery,
		listeners:    conf.Listeners,
		pollInterval: conf.PollInterval,
	}
}

// Not safe if monitor has started
func (m *discoveryMonitor) Listeners() []DiscoveryListener {
	return m.listeners
}

// Not safe if monitor has started
func (m *discoveryMonitor) SetListeners(listeners []DiscoveryListener) {
	m.listeners = listeners
}

// Start monitoring discovery
func (m *discoveryMonitor) Start() error {
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

// poll discovery for changes in the catalog
func (m *discoveryMonitor) poll() error {
	// Get newest catalog from discovery
	instances, err := m.discovery.ListInstances()
	if err != nil {
		logrus.WithError(err).Warn("Could not get latest catalog from discovery")
		return err
	}
	catalog := instanceListAsMap(instances)

	// Check for changes
	if m.compareToCache(catalog) {
		// Match, nothing else to do
		return nil
	}

	// Update cached catalog
	m.cache = catalog

	// Notify the listeners
	instancesByValue := instanceListAsValues(instances)
	for _, listener := range m.listeners {
		if err = listener.CatalogChange(instancesByValue); err != nil {
			logrus.WithError(err).Warn("Registry listener failed")
		}
	}

	return nil
}

// compareToCache compares the given catalog to the cached one, by comparing all instance attributes
// except for heartbeat and TTL. Instance list for each service is assumed to be presorted.
// Return 'true' if catalog match, and 'false' otherwise.
func (m *discoveryMonitor) compareToCache(catalog map[string][]*api.ServiceInstance) bool {
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

// Stop monitoring discovery
func (m *discoveryMonitor) Stop() error {
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
