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
	"encoding/json"
	"time"
)

// ServiceInstance holds information about a service instance registered with Amalgam8 Service Registry.
//
// It is used both as an input to registration calls, as well as the output of discovery calls.
// Depending on the context, some of the fields may be mandatory, optional, or ignored.
type ServiceInstance struct {

	// ID is the unique ID assigned to this service instance by Amalgam8 Service Registry.
	// This field is ignored for registration, and is mandatory for discovery.
	ID string `json:"id,omitempty"`

	// ServiceName is the name of the service being provided by this service instance.
	// This field is mandatory both for registration and discovery.
	ServiceName string `json:"service_name"`

	// Endpoint is the network endpoint of this service instance.
	// This field is mandatory both for registration and discovery.
	Endpoint ServiceEndpoint `json:"endpoint"`

	// Status is a string representing the status of the service instance.
	// Valid values are "STARTING", "UP", or "OUT_OF_SERVICE".
	// This field is optional both for registration and discovery.
	Status string `json:"status,omitempty"`

	// Tags is a set of arbitrary tags attached to this service instance.
	// This field is optional both for registration and discovery.
	Tags []string `json:"tags,omitempty"`

	// Metadata is a marshaled JSON value associated with this service instance, in encoded-form.
	// Any arbitrary JSON value is valid, including numbers, strings, arrays and objects.
	// This field is optional both for registration and discovery.
	Metadata json.RawMessage `json:"metadata,omitempty"`

	// TTL is the time-to-live associated with this service instance, specified in seconds.
	// This field is optional for registration, and is mandatory for discovery.
	TTL int `json:"ttl,omitempty"`

	// LastHeartbeat is the timestamp in which heartbeat has been last received for this service instance.
	// This field is ignored for registration, and is mandatory for discovery.
	LastHeartbeat time.Time `json:"last_heartbeat,omitempty"`
}
