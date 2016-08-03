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

package env

// The following are the variables exposed by the Registry to the wrapping middlewares
// as request.Env[variable].(type)
const (
	// Namespace set by the auth middleware based on the token in the request
	// Type: auth.Namespace
	Namespace = "NAMESPACE"

	// RequestID defines the unique id of the request, set by the trace middleware
	// Type: string
	RequestID = "REQUEST_ID"

	// StatusCode defines the HTTP status code, set by the rest.RecorderMiddleware
	// Type: int
	StatusCode = "STATUS_CODE"

	// BytesWritten defines the number of data bytes written to the response, set by the rest.RecorderMiddleware
	// Type: int64
	BytesWritten = "BYTES_WRITTEN"

	// StartTime defines the time when the execution of the request was started, set by the rest.TimerMiddleware
	// Type: *time.Time
	StartTime = "START_TIME"

	// ElapsedTime defines the elapsed time spent during the execution of the wrapped handler, set by the rest. TimerMiddleware
	// Type: *time.Duration
	ElapsedTime = "ELAPSED_TIME"

	// APIProtocol set by the protocol.APIHandler
	// Type: protocol.Type
	APIProtocol = "API_PROTOCOL"

	// APIOperation set by the protocol.APIHandler
	// Type: protocol.Operation
	APIOperation = "API_OPERATION"

	// ServiceInstance is the instance on which the current API operation operates upon, set by the corresponding HTTP handler.
	// Relevant only for operations targeted at a single service instance (register, deregister, heartbeat, ...).
	// Type: *store.ServiceInstance
	ServiceInstance = "SERVICE_INSTANCE"
)
