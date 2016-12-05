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

package client

import (
	"time"

	"sync"

	"github.com/amalgam8/amalgam8/pkg/api"
)

// Make sure we implement the ServiceDiscovery and ServiceRegistry interface.
var _ api.ServiceDiscovery = (*Cache)(nil)
var _ api.ServiceRegistry = (*Cache)(nil)

// CacheConfig stores the configurable attributes of the caching client.
type CacheConfig struct {
	Config

	// PollInterval is the time interval in which the caching client refreshes its local cache
	PollInterval time.Duration
}

// Cache implements the ServiceDiscovery and ServiceRegistry interfaces using a local cache.
// The cache is refreshed periodically using the non-caching, REST API-based Amalagma8 Registry client.
// The ServiceRegistry calls are pass
type Cache struct {
	Client
	cache map[string][]*api.ServiceInstance
	mutex sync.RWMutex
}

// NewCache constructs a new Caching Client using the given configuration.
func NewCache(config CacheConfig) (*Cache, error) {
	cl, err := New(config.Config)
	if err != nil {
		return nil, err
	}

	c := &Cache{
		Client: *cl,
		cache:  make(map[string][]*api.ServiceInstance),
	}

	go c.maintain(config.PollInterval)
	return c, nil
}

// ListServices queries for the list of services for which instances are currently registered.
func (c *Cache) ListServices() ([]string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	services := make([]string, 0, len(c.cache))
	for service := range c.cache {
		services = append(services, service)
	}

	return services, nil
}

// ListInstances queries for the list of service instances currently registered.
func (c *Cache) ListInstances() ([]*api.ServiceInstance, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	instances := make([]*api.ServiceInstance, 0, len(c.cache)*3)
	for _, service := range c.cache {
		instances = append(instances, service...)
	}

	return instances, nil
}

// ListServiceInstances queries for the list of service instances currently registered for the given service.
func (c *Cache) ListServiceInstances(serviceName string) ([]*api.ServiceInstance, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	service := c.cache[serviceName]
	instances := make([]*api.ServiceInstance, 0, len(service))
	instances = append(instances, service...)

	return instances, nil
}

func (c *Cache) maintain(pollInterval time.Duration) {
	go c.refresh()
	for range time.Tick(pollInterval) {
		go c.refresh()
	}
}

func (c *Cache) refresh() {
	instanceList, err := c.Client.ListInstances()
	if err != nil {
		return
	}

	instanceMap := make(map[string][]*api.ServiceInstance)
	for _, instance := range instanceList {
		instanceMap[instance.ServiceName] = append(instanceMap[instance.ServiceName], instance)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache = instanceMap
}
