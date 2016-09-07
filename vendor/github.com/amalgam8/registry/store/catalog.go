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

// Predicate for filtering returned instances
type Predicate func(si *ServiceInstance) bool

// Catalog for managing instances within a registry namespace
type Catalog interface {
	Register(si *ServiceInstance) (*ServiceInstance, error)
	Deregister(instanceID string) (*ServiceInstance, error)
	Renew(instanceID string) (*ServiceInstance, error)
	SetStatus(instanceID, status string) (*ServiceInstance, error)

	Instance(instanceID string) (*ServiceInstance, error)
	List(serviceName string, predicate Predicate) ([]*ServiceInstance, error)
	ListServices(predicate Predicate) []*Service
}
