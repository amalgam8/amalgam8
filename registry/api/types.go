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

import (
	"encoding/json"
	"time"
)

// ServiceDiscovery defines the interface used for discovering service instances.
type ServiceDiscovery interface {

	// ListServices queries for the list of services for which instances are currently registered.
	ListServices() ([]string, error)

	// ListInstances queries for the list of service instances currently registered.
	ListInstances() ([]*ServiceInstance, error)

	// ListServiceInstances queries for the list of service instances currently registered for the given service.
	ListServiceInstances(serviceName string) ([]*ServiceInstance, error)
}

// ServiceRegistry defines the interface used for registering service instances.
type ServiceRegistry interface {

	// Register adds a service instance, described by the given ServiceInstance structure, to the registry.
	Register(instance *ServiceInstance) (*ServiceInstance, error)

	// Deregister removes a registered service instance, identified by the given ID, from the registry.
	Deregister(id string) error

	// Renew sends a heartbeat for the service instance identified by the given ID.
	Renew(id string) error
}

// ServiceInstance describes an instance of a service.
type ServiceInstance struct {

	// ID is the unique ID assigned to this service instance.
	ID string `json:"id,omitempty"`

	// ServiceName is the name of the service being provided by this service instance.
	ServiceName string `json:"service_name,omitempty"`

	// Endpoint is the network endpoint of this service instance.
	Endpoint ServiceEndpoint `json:"endpoint,omitempty"`

	// Status is a string representing the status of the service instance.
	Status string `json:"status,omitempty"`

	// Tags is a set of arbitrary tags attached to this service instance.
	Tags []string `json:"tags,omitempty"`

	// Metadata is a marshaled JSON value (object, string, ...) associated with this service instance, in encoded-form.
	Metadata json.RawMessage `json:"metadata,omitempty"`

	// TTL is the time-to-live associated with this service instance, specified in seconds.
	TTL int `json:"ttl,omitempty"`

	// LastHeartbeat is the timestamp in which heartbeat has been last received for this service instance.
	LastHeartbeat time.Time `json:"last_heartbeat,omitempty"`
}

// ServiceEndpoint describes a network endpoint of a service.
type ServiceEndpoint struct {

	// Type is the endpoint's type, normally a protocol name, like "http", "https", "tcp", or "udp".
	Type string `json:"type"`

	// Value is the endpoint's value according to its type,
	// e.g. "172.135.10.1:8080" or "http://myapp.ng.bluemix.net/api/v1".
	Value string `json:"value"`
}
