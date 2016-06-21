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
	"fmt"
	"net/url"
)

// ServiceEndpoint describes a network endpoint of service instance.
type ServiceEndpoint struct {

	// Type is the endpoint's type. Valid values are "http", "tcp", or "user".
	Type string `json:"type"`

	// Value is the endpoint's value according to its type,
	// e.g. "172.135.10.1:8080" or "http://myapp.ng.bluemix.net/api/v1".
	Value string `json:"value"`
}

// NewHTTPEndpoint creates a new HTTP(S) network endpoint with the specified URL.
// The specified URL is assumed to have an "http" or "https" scheme.
func NewHTTPEndpoint(httpURL url.URL) ServiceEndpoint {
	return ServiceEndpoint{
		Type:  "http",
		Value: httpURL.String(),
	}
}

// NewTCPEndpoint creates a new TCP network endpoint with the specified host and optional port.
func NewTCPEndpoint(host string, port int) ServiceEndpoint {
	var value string
	if port > 0 {
		value = fmt.Sprintf("%s:%d", host, port)
	} else {
		value = host
	}
	return ServiceEndpoint{
		Type:  "tcp",
		Value: value,
	}
}

// NewCustomEndpoint creates a new network endpoint of a custom type.
// The value may be any arbitrary description of a network endpoint which makes sense in the context of the application.
func NewCustomEndpoint(endpoint string) ServiceEndpoint {
	return ServiceEndpoint{
		Type:  "user",
		Value: endpoint,
	}
}
