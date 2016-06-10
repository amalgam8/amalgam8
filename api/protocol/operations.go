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

package protocol

import (
	"github.com/ant0ine/go-json-rest/rest"
)

// Operation represents an operation exposed by the Service Discovery API.
type Operation string

// The following are the current API operations exposed by Service Discovery.
//
// While most operations have implementations in both API protocols (Amalgam8 / Eureka),
// some are unique to a certain protocol - e.g., SetInstanceStatus is currently unique to Eureka).
//
// Also, several sub-operations may be mapped to the same Operation value, e.g. Amalgam8's ListInstances
// as well as Eureka's ListVips are both mapped to ListInstances.
const (
	RegisterInstance     Operation = "Register"
	DeregisterInstance             = "Deregister"
	RenewInstance                  = "Renew"
	ListServices                   = "ListServices"
	ListServiceInstances           = "ListServiceInstances"
	ListInstances                  = "ListInstances"
	SetInstanceStatus              = "SetStatus"
	GetInstance                    = "GetInfo"
)

// String returns a string representation of this Operation value.
func (op Operation) String() string {
	return string(op)
}

// Keys used in the HTTP request context (r.Env) to store API information
const (
	ProtocolKey  = "APIProtocol"
	OperationKey = "APIOperation"
)

// APIHandler returns a wrapper HandlerFunc that injects API information into the HTTP request's context (r.Env),
// before calling the provided HandlerFunc.
// The given protocol is injected as the ProtocolKey, and the given operation as the OperationKey.
func APIHandler(handler rest.HandlerFunc, protocol Type, operation Operation) rest.HandlerFunc {
	return func(w rest.ResponseWriter, r *rest.Request) {
		r.Env[ProtocolKey] = protocol
		r.Env[OperationKey] = operation
		handler(w, r)
	}
}
