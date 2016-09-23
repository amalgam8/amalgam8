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

const (
	registryRecordKey = "reg"
	namespaceKey      = "ns"
	instanceKey       = "inst"
	keySeparator      = ":"
)

// ExternalRegistry calls to manage the instances in an external store
type ExternalRegistry interface {
	ReadKeys(namespace auth.Namespace) ([]string, error)
	ReadServiceInstanceByInstID(namespace auth.Namespace, instanceID string) (*ServiceInstance, error)
	ListServiceInstancesByName(namespace auth.Namespace, name string) (map[string]*ServiceInstance, error)
	ListAllServiceInstances(namespace auth.Namespace) (map[string]ServiceInstanceMap, error)
	InsertServiceInstance(namespace auth.Namespace, instance *ServiceInstance) error
	DeleteServiceInstance(namespace auth.Namespace, instanceID string) (int, error)
}

// DBKey represents the service instance key
type DBKey struct {
	Namespace  string
	InstanceID string
}

// Database key as JSON string
func (dbk *DBKey) String() string {
	parts := []string{registryRecordKey, namespaceKey, dbk.Namespace, instanceKey, dbk.InstanceID}
	return strings.Join(parts, keySeparator)
}

// Return a DBKey from a string
func parseStringIntoDBKey(key string) (*DBKey, error) {
	// Validate the string is a DBKey
	// Construct the prefix: reg:ns:
	prefixParts := []string{registryRecordKey, namespaceKey, ""}
	prefix := strings.Join(prefixParts, keySeparator)
	// Construct the instance key with separators: :inst:
	instParts := []string{keySeparator, instanceKey, keySeparator}
	instKey := strings.Join(instParts, "")
	if !strings.HasPrefix(key, prefix) || !strings.Contains(key, instKey) {
		return nil, fmt.Errorf("Not a valid DBKey string")
	}

	// Remove the beginning reg:ns: key
	parsed := strings.SplitN(key, prefix, 2)
	// Split the namespace and instance id using :inst:
	parsed = strings.SplitN(parsed[1], instKey, 2)

	return &DBKey{Namespace: parsed[0], InstanceID: parsed[1]}, nil
}
