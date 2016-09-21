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
	"encoding/json"

	"github.com/amalgam8/amalgam8/pkg/auth"
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
	ComponentFlag string `json:"comp"`
	Namespace     string `json:"ns"`
	InstanceID    string `json:"iid"`
}

// Database key as JSON string
func (dbk *DBKey) String() string {
	dbk.ComponentFlag = "reg"
	dbKey, _ := json.Marshal(dbk)
	return string(dbKey[:])
}
