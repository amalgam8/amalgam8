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
	"github.com/amalgam8/amalgam8/pkg/auth"
)

type mockExternalRegistry struct {
	mockServiceInstances map[string]*ServiceInstance
}

// NewMockExternalRegistry is a mock version of the external registry
func NewMockExternalRegistry(mockServiceInstances map[string]*ServiceInstance) ExternalRegistry {
	db := &mockExternalRegistry{
		mockServiceInstances: mockServiceInstances,
	}

	return db
}

func (mer *mockExternalRegistry) ReadKeys(namespace auth.Namespace) ([]string, error) {
	var keys []string

	for key := range mer.mockServiceInstances {
		keys = append(keys, key)
	}

	return keys, nil
}

func (mer *mockExternalRegistry) ReadServiceInstanceByInstID(namespace auth.Namespace, instanceID string) (*ServiceInstance, error) {
	for _, instance := range mer.mockServiceInstances {
		if instanceID == instance.ID {
			return instance.DeepClone(), nil
		}
	}

	return &ServiceInstance{}, nil
}

func (mer *mockExternalRegistry) ListServiceInstancesByName(namespace auth.Namespace, name string) (map[string]*ServiceInstance, error) {
	siMap := make(map[string]*ServiceInstance)

	for _, instance := range mer.mockServiceInstances {
		if name == instance.ServiceName {
			dbKey := DBKey{InstanceID: instance.ID, Namespace: namespace.String()}
			siMap[dbKey.String()] = instance.DeepClone()
			continue
		}
	}
	return siMap, nil
}

func (mer *mockExternalRegistry) ListAllServiceInstances(namespace auth.Namespace) (map[string]ServiceInstanceMap, error) {
	// Sort the instances by service name
	serviceList := make(map[string][]*ServiceInstance)
	for _, instance := range mer.mockServiceInstances {
		serviceList[instance.ServiceName] = append(serviceList[instance.ServiceName], instance)
	}

	serviceInstances := make(map[string]ServiceInstanceMap)
	for sname, instances := range serviceList {
		siMap := make(ServiceInstanceMap)
		for _, inst := range instances {
			dbKey := DBKey{InstanceID: inst.ID, Namespace: inst.ServiceName}
			siMap[dbKey] = inst.DeepClone()
		}
		serviceInstances[sname] = siMap
	}

	return serviceInstances, nil
}

func (mer *mockExternalRegistry) InsertServiceInstance(namespace auth.Namespace, instance *ServiceInstance) error {
	si := instance.DeepClone()
	dbKey := DBKey{InstanceID: instance.ID, Namespace: namespace.String()}
	mer.mockServiceInstances[dbKey.String()] = si
	return nil
}

func (mer *mockExternalRegistry) DeleteServiceInstance(namespace auth.Namespace, instanceID string) (int, error) {
	var count int

	dbKey := DBKey{InstanceID: instanceID, Namespace: namespace.String()}
	// Don't need to loop to delete, but need a count of instances deleted
	for index := range mer.mockServiceInstances {
		if index == dbKey.String() {
			count++
			// Remove the instance from the slice
			delete(mer.mockServiceInstances, dbKey.String())
		}
	}
	return count, nil
}
