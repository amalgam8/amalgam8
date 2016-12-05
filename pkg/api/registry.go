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

// Package api defines core types used in the registry
package api

// ServiceRegistry defines the interface used for registering service instances.
type ServiceRegistry interface {

	// Register adds a service instance, described by the given ServiceInstance structure, to the registry.
	Register(instance *ServiceInstance) (*ServiceInstance, error)

	// Deregister removes a registered service instance, identified by the given ID, from the registry.
	Deregister(id string) error

	// Renew sends a heartbeat for the service instance identified by the given ID.
	Renew(id string) error
}
