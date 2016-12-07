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

package api

import (
	"encoding/json"
	"time"
)

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
