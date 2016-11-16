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

package kubernetes

import (
	"testing"

	"github.com/amalgam8/amalgam8/registry/api"
	"github.com/stretchr/testify/assert"
)

func TestParseEndpoint(t *testing.T) {
	cases := []struct {
		addr             EndpointAddress
		port             EndpointPort
		expectedEndpoint *api.ServiceEndpoint
		expectedError    bool
	}{
		{
			addr:             EndpointAddress{IP: "10.0.1.1"},
			port:             EndpointPort{Port: 53, Protocol: Protocol("UDP")},
			expectedEndpoint: &api.ServiceEndpoint{Type: "udp", Value: "10.0.1.1:53"},
		},
		{
			addr:             EndpointAddress{IP: "10.0.1.1"},
			port:             EndpointPort{Port: 5000, Protocol: Protocol("TCP")},
			expectedEndpoint: &api.ServiceEndpoint{Type: "tcp", Value: "10.0.1.1:5000"},
		},
		{
			addr:             EndpointAddress{IP: "10.0.1.1"},
			port:             EndpointPort{Port: 5000, Protocol: Protocol("TCP"), Name: "donald-duck"},
			expectedEndpoint: &api.ServiceEndpoint{Type: "tcp", Value: "10.0.1.1:5000"},
		},
		{
			addr:             EndpointAddress{IP: "10.0.1.1"},
			port:             EndpointPort{Port: 80, Protocol: Protocol("TCP"), Name: "http"},
			expectedEndpoint: &api.ServiceEndpoint{Type: "http", Value: "10.0.1.1:80"},
		},
		{
			addr:             EndpointAddress{IP: "10.0.1.1"},
			port:             EndpointPort{Port: 80, Protocol: Protocol("TCP"), Name: "HTTP"},
			expectedEndpoint: &api.ServiceEndpoint{Type: "http", Value: "10.0.1.1:80"},
		},
		{
			addr:             EndpointAddress{IP: "10.0.1.1"},
			port:             EndpointPort{Port: 943, Protocol: Protocol("TCP"), Name: "https"},
			expectedEndpoint: &api.ServiceEndpoint{Type: "https", Value: "10.0.1.1:943"},
		},
		{
			addr:             EndpointAddress{IP: "10.0.1.1"},
			port:             EndpointPort{Port: 943, Protocol: Protocol("TCP"), Name: "HTTPS"},
			expectedEndpoint: &api.ServiceEndpoint{Type: "https", Value: "10.0.1.1:943"},
		},
		{
			addr:          EndpointAddress{IP: "10.0.1.1"},
			port:          EndpointPort{Port: 80, Protocol: Protocol("WTF")},
			expectedError: true,
		},
		{
			addr:          EndpointAddress{IP: "10.0.1.1"},
			port:          EndpointPort{Port: 80, Protocol: Protocol("")},
			expectedError: true,
		},
	}

	for i, c := range cases {
		endpoint, err := parseEndpoint(c.addr, c.port)

		if c.expectedError {
			assert.Error(t, err, "Expected non-nil error for test-case %d", i)
		} else {
			assert.NoError(t, err, "Expected no error for test-case %d", i)
		}

		assert.Equal(t, c.expectedEndpoint, endpoint, "Wrong endpoint for test-case %d", i)
	}
}
