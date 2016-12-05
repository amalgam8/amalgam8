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

	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/pkg/api/v1"
)

func TestParseEndpoint(t *testing.T) {
	cases := []struct {
		addr             v1.EndpointAddress
		port             v1.EndpointPort
		expectedEndpoint *api.ServiceEndpoint
		expectedError    bool
	}{
		{
			addr:             v1.EndpointAddress{IP: "10.0.1.1"},
			port:             v1.EndpointPort{Port: 53, Protocol: v1.ProtocolUDP},
			expectedEndpoint: &api.ServiceEndpoint{Type: "udp", Value: "10.0.1.1:53"},
		},
		{
			addr:             v1.EndpointAddress{IP: "10.0.1.1"},
			port:             v1.EndpointPort{Port: 5000, Protocol: v1.ProtocolTCP},
			expectedEndpoint: &api.ServiceEndpoint{Type: "tcp", Value: "10.0.1.1:5000"},
		},
		{
			addr:             v1.EndpointAddress{IP: "10.0.1.1"},
			port:             v1.EndpointPort{Port: 5000, Protocol: v1.ProtocolTCP, Name: "donald-duck"},
			expectedEndpoint: &api.ServiceEndpoint{Type: "tcp", Value: "10.0.1.1:5000"},
		},
		{
			addr:             v1.EndpointAddress{IP: "10.0.1.1"},
			port:             v1.EndpointPort{Port: 80, Protocol: v1.ProtocolTCP, Name: "http"},
			expectedEndpoint: &api.ServiceEndpoint{Type: "http", Value: "10.0.1.1:80"},
		},
		{
			addr:             v1.EndpointAddress{IP: "10.0.1.1"},
			port:             v1.EndpointPort{Port: 80, Protocol: v1.ProtocolTCP, Name: "HTTP"},
			expectedEndpoint: &api.ServiceEndpoint{Type: "http", Value: "10.0.1.1:80"},
		},
		{
			addr:             v1.EndpointAddress{IP: "10.0.1.1"},
			port:             v1.EndpointPort{Port: 943, Protocol: v1.ProtocolTCP, Name: "https"},
			expectedEndpoint: &api.ServiceEndpoint{Type: "https", Value: "10.0.1.1:943"},
		},
		{
			addr:             v1.EndpointAddress{IP: "10.0.1.1"},
			port:             v1.EndpointPort{Port: 943, Protocol: v1.ProtocolTCP, Name: "HTTPS"},
			expectedEndpoint: &api.ServiceEndpoint{Type: "https", Value: "10.0.1.1:943"},
		},
		{
			addr:          v1.EndpointAddress{IP: "10.0.1.1"},
			port:          v1.EndpointPort{Port: 80, Protocol: v1.Protocol("WTF")},
			expectedError: true,
		},
		{
			addr:          v1.EndpointAddress{IP: "10.0.1.1"},
			port:          v1.EndpointPort{Port: 80, Protocol: v1.Protocol("")},
			expectedError: true,
		},
	}

	for i, c := range cases {
		endpoint, err := buildEndpointFromAddress(c.addr, c.port)

		if c.expectedError {
			assert.Error(t, err, "Expected non-nil error for test-case %d", i)
		} else {
			assert.NoError(t, err, "Expected no error for test-case %d", i)
		}

		assert.Equal(t, c.expectedEndpoint, endpoint, "Wrong endpoint for test-case %d", i)
	}
}
