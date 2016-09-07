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
	"time"
)

// Registered instance status related constants
const (
	Starting     = "STARTING"
	Up           = "UP"
	OutOfService = "OUT_OF_SERVICE"
	All          = "ALL" // ALL is only a valid status for the query string param and not for the register
)

// ServiceInstance represents a runtime instance of a service.
type ServiceInstance struct {
	ID               string
	ServiceName      string
	Endpoint         *Endpoint
	Status           string
	Metadata         []byte
	RegistrationTime time.Time
	LastRenewal      time.Time
	TTL              time.Duration
	Tags             []string
	Extension        map[string]interface{}
}

// String output the structure
func (si *ServiceInstance) String() string {
	return fmt.Sprintf("id: %s, service_name: %s, endpoint: %s, status: %s, registrationTime: %v, lastRenewal: %v, ttl: %d, tags: %v",
		si.ID, si.ServiceName, si.Endpoint, si.Status, si.RegistrationTime, si.LastRenewal, si.TTL, si.Tags)
}

// DeepClone creates a deep copy of the receiver
func (si *ServiceInstance) DeepClone() *ServiceInstance {
	cloned := *si
	cloned.Endpoint = si.Endpoint.DeepClone()
	if si.Metadata == nil || len(si.Metadata) == 0 {
		cloned.Metadata = nil
	} else {
		cloned.Metadata = make([]byte, len(si.Metadata))
		copy(cloned.Metadata, si.Metadata)
	}
	if len(si.Extension) == 0 {
		cloned.Extension = nil
	} else {
		cloned.Extension = make(map[string]interface{}, len(si.Extension))
		for k, v := range si.Extension {
			cloned.Extension[k] = v
		}
	}
	return &cloned
}

// ServiceInstanceMap is a map of ServiceInstances keyed by instance id and service name
type ServiceInstanceMap map[DBKey]*ServiceInstance
