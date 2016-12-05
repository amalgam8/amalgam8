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

package store

import (
	"time"

	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/pkg/auth"
)

// DiscoveryFactory is a function which accepts a namespace and returns a discovery object for that namespace
type DiscoveryFactory func(namespace auth.Namespace) (api.ServiceDiscovery, error)

// NewDiscoveryAdapter creates a CatalogFactory from the given DiscoveryFactory
func NewDiscoveryAdapter(factory DiscoveryFactory) CatalogFactory {
	return &discoveryAdapterCatalogFactory{
		DiscoveryFactory: factory,
	}
}

type discoveryAdapterCatalogFactory struct {
	DiscoveryFactory
}

func (f *discoveryAdapterCatalogFactory) CreateCatalog(namespace auth.Namespace) (Catalog, error) {
	discovery, err := f.DiscoveryFactory(namespace)
	if err != nil {
		return nil, err
	}
	return newDiscoveryAdapterCatalog(discovery), nil
}

// discoveryAdapterCatalog provides a read-only catalog interface over the given discovery interface.
type discoveryAdapterCatalog struct {
	discovery api.ServiceDiscovery
}

func newDiscoveryAdapterCatalog(discovery api.ServiceDiscovery) *discoveryAdapterCatalog {
	return &discoveryAdapterCatalog{
		discovery: discovery,
	}
}

func (c *discoveryAdapterCatalog) Register(si *ServiceInstance) (*ServiceInstance, error) {
	return nil, NewError(ErrorBadRequest, "Read-only Catalog: API Not Supported", "Register")
}

func (c *discoveryAdapterCatalog) Deregister(instanceID string) (*ServiceInstance, error) {
	return nil, NewError(ErrorBadRequest, "Read-only Catalog: API Not Supported", "Deregister")
}

func (c *discoveryAdapterCatalog) Renew(instanceID string) (*ServiceInstance, error) {
	return nil, NewError(ErrorBadRequest, "Read-only Catalog: API Not Supported", "Renew")
}

func (c *discoveryAdapterCatalog) SetStatus(instanceID, status string) (*ServiceInstance, error) {
	return nil, NewError(ErrorBadRequest, "Read-only Catalog: API Not Supported", "SetStatus")
}

func (c *discoveryAdapterCatalog) List(serviceName string, predicate Predicate) ([]*ServiceInstance, error) {
	registryInstances, err := c.discovery.ListServiceInstances(serviceName)
	if err != nil {
		return nil, err
	}

	if len(registryInstances) == 0 {
		return nil, NewError(ErrorNoSuchServiceName, "no such service", serviceName)
	}

	instances := convertServiceInstances(registryInstances)
	if predicate != nil {
		instances = filterServiceInstances(instances, predicate)
	}

	return instances, nil
}

func (c *discoveryAdapterCatalog) ListServices(predicate Predicate) []*Service {
	registryInstances, err := c.discovery.ListInstances()
	if err != nil {
		return nil
	}

	instances := convertServiceInstances(registryInstances)
	if predicate != nil {
		instances = filterServiceInstances(instances, predicate)
	}

	serviceNames := make(map[string]struct{}, len(instances))
	for _, instance := range instances {
		serviceNames[instance.ServiceName] = struct{}{}
	}

	services := make([]*Service, 0, len(serviceNames))
	for serviceName := range serviceNames {
		services = append(services, &Service{ServiceName: serviceName})
	}

	return services
}

func (c *discoveryAdapterCatalog) Instance(instanceID string) (*ServiceInstance, error) {
	registryInstances, err := c.discovery.ListInstances()
	if err != nil {
		return nil, err
	}

	for _, instance := range registryInstances {
		if instance.ID == instanceID {
			return convertServiceInstance(instance), nil
		}
	}

	return nil, NewError(ErrorNoSuchServiceInstance, "no such service instance", instanceID)
}

// convert the instances structs to the store package own representation
func convertServiceInstances(instances []*api.ServiceInstance) []*ServiceInstance {
	converted := make([]*ServiceInstance, len(instances))
	for i, instance := range instances {
		converted[i] = convertServiceInstance(instance)
	}
	return converted
}

func convertServiceInstance(instance *api.ServiceInstance) *ServiceInstance {
	return &ServiceInstance{
		ID:          instance.ID,
		ServiceName: instance.ServiceName,
		Endpoint: &Endpoint{
			Type:  instance.Endpoint.Type,
			Value: instance.Endpoint.Value,
		},
		Status:      instance.Status,
		Metadata:    instance.Metadata,
		LastRenewal: instance.LastHeartbeat,
		TTL:         time.Duration(instance.TTL) * time.Second,
		Tags:        instance.Tags,
		Extension:   map[string]interface{}{},
	}
}

func filterServiceInstances(instances []*ServiceInstance, predicate Predicate) []*ServiceInstance {
	filtered := make([]*ServiceInstance, 0, len(instances))
	for _, instance := range instances {
		if predicate(instance) {
			filtered = append(filtered, instance)
		}
	}
	return filtered
}
