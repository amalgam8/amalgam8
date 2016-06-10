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

// APIDescriptor encapsulates information associated with an API call.
type APIDescriptor struct {

	// The path part of the URL.
	Path string

	// The HTTP method (e.g. "GET"). Valid values are defined in the MethodXXX constants in the http package.
	Method string

	// The protocol (e.g., Eureka).
	Protocol Type

	// The logic operation (e.g., RegisterInstance).
	Operation Operation

	// The handler registered for this API call.
	Handler rest.HandlerFunc
}

// AsRoute converts this APIDescriptor into a rest.Route object.
func (desc *APIDescriptor) AsRoute() *rest.Route {
	return &rest.Route{
		HttpMethod: desc.Method,
		PathExp:    desc.Path,
		Func:       desc.Handler,
	}
}
