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
	"fmt"
	"strings"

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

func (mer *mockExternalRegistry) ListServiceInstancesByKey(namespace auth.Namespace, key string) (map[string]*ServiceInstance, error) {
	instanceID := strings.SplitN(key, ".", 2)[0]
	serviceName := strings.SplitN(key, ".", 2)[1]
	siMap := make(map[string]*ServiceInstance)

	for _, instance := range mer.mockServiceInstances {
		if instanceID == "*" && serviceName == "*" {
			siMap[fmt.Sprintf("%s.%s", instance.ID, instance.ServiceName)] = instance.DeepClone()
			continue
		}
		if instanceID == "*" {
			if serviceName == instance.ServiceName {
				siMap[fmt.Sprintf("%s.%s", instance.ID, instance.ServiceName)] = instance.DeepClone()
			}
			continue
		}
		if serviceName == "*" {
			if instanceID == instance.ID {
				siMap[fmt.Sprintf("%s.%s", instance.ID, instance.ServiceName)] = instance.DeepClone()
			}
			continue
		}
		if instanceID == instance.ID && serviceName == instance.ServiceName {
			siMap[fmt.Sprintf("%s.%s", instance.ID, instance.ServiceName)] = instance.DeepClone()
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
			dbKey := DBKey{inst.ID, inst.ServiceName}
			siMap[dbKey] = inst.DeepClone()
		}
		serviceInstances[sname] = siMap
	}

	return serviceInstances, nil
}

func (mer *mockExternalRegistry) InsertServiceInstance(namespace auth.Namespace, instance *ServiceInstance) error {
	si := instance.DeepClone()
	mer.mockServiceInstances[fmt.Sprintf("%s.%s", instance.ID, instance.ServiceName)] = si
	return nil
}

func (mer *mockExternalRegistry) DeleteServiceInstance(namespace auth.Namespace, key string) (int, error) {
	var count int

	// Don't need to loop to delete, but need a count of instances deleted
	for index := range mer.mockServiceInstances {
		if index == key {
			count++
			// Remove the instance from the slice
			delete(mer.mockServiceInstances, key)
		}
	}
	return count, nil
}
