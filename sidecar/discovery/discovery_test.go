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

package discovery

import (
	"fmt"
	"testing"

	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/sidecar/proxy/envoy"
	"github.com/stretchr/testify/assert"
)

var shoppingCartInstances = []*api.ServiceInstance{
	{
		ServiceName: "shoppingCart",
		ID:          "1",
		Endpoint:    api.ServiceEndpoint{Type: "http", Value: "127.0.0.1:8080"},
	},
	{
		ServiceName: "shoppingCart",
		Tags:        []string{"first", "third"},
		ID:          "2",
		Endpoint:    api.ServiceEndpoint{Type: "tcp", Value: "127.0.0.5:5050"},
	},
	{
		Tags:        []string{"first", "second"},
		ServiceName: "shoppingCart",
		ID:          "3",
		Endpoint:    api.ServiceEndpoint{Type: "tcp", Value: "127.0.0.4:3050"},
	},
	{
		Tags:        []string{"second"},
		ServiceName: "shoppingCart",
		ID:          "8",
		Endpoint:    api.ServiceEndpoint{Type: "tcp", Value: "127.0.0.4:3050"},
	},
}

func TestTranslate(t *testing.T) {
	hosts := translate(shoppingCartInstances)
	fmt.Println(hosts)

	assert.Len(t, hosts, len(shoppingCartInstances))
}

func TestFilter(t *testing.T) {
	instances := filterInstances(shoppingCartInstances, []string{})
	assert.Len(t, instances, len(shoppingCartInstances))

	instances = filterInstances(shoppingCartInstances, []string{"first"})
	assert.Len(t, instances, 2)

	instances = filterInstances(shoppingCartInstances, []string{"first", "second"})
	assert.Len(t, instances, 1)

	instances = filterInstances(shoppingCartInstances, []string{"fourth"})
	assert.Len(t, instances, 0)
}

func TestBuildClusters(t *testing.T) {
	instances := []*api.ServiceInstance{
		{
			ServiceName: "service1",
			Tags:        []string{"tag1"},
			ID:          "1",
			Endpoint:    api.ServiceEndpoint{Type: "http", Value: "127.0.0.1:8080"},
		},
		{
			ServiceName: "service1",
			Tags:        []string{"tag1", "tag2"},
			ID:          "2",
			Endpoint:    api.ServiceEndpoint{Type: "tcp", Value: "127.0.0.5:5050"},
		},
		{
			ServiceName: "service1",
			ID:          "3",
			Endpoint:    api.ServiceEndpoint{Type: "tcp", Value: "127.0.0.4:3050"},
		},
		{
			Tags:        []string{"second"},
			ServiceName: "service2",
			ID:          "8",
			Endpoint:    api.ServiceEndpoint{Type: "tcp", Value: "127.0.0.4:3050"},
		},
	}

	rules := []api.Rule{
		{
			ID:          "abcdef",
			Destination: "service1",
			Route: &api.Route{
				Backends: []api.Backend{
					{
						Name:   "service1",
						Tags:   []string{"tag1"},
						Weight: 0.25,
					},
				},
			},
		},
		{
			ID:          "abcdef",
			Destination: "service1",
			Route: &api.Route{
				Backends: []api.Backend{
					{
						Name:   "service1",
						Tags:   []string{"tag1", "tag2"},
						Weight: 0.75,
						LbType: "random",
					},
				},
			},
		},
		{
			ID:          "abcdef",
			Destination: "service2",
			Route: &api.Route{
				Backends: []api.Backend{
					{
						Name:   "service2",
						Tags:   []string{},
						Weight: 1.00,
					},
				},
			},
		},
		{
			ID:          "abcdef",
			Destination: "service2",
			Actions:     []api.Action{},
		},
	}

	clusters := buildClusters(instances, rules, nil)

	assert.Len(t, clusters, 5)

	clusterName := envoy.BuildServiceKey("service1", []string{"tag1"})
	assert.Equal(t, envoy.Cluster{
		Name:             clusterName,
		ServiceName:      clusterName,
		Type:             "sds",
		LbType:           "round_robin",
		ConnectTimeoutMs: 1000,
		OutlierDetection: &envoy.OutlierDetection{
			MaxEjectionPercent: 100,
		},
		CircuitBreakers: &envoy.CircuitBreakers{},
	}, clusters[1])

	assert.Equal(t, envoy.BuildServiceKey("service1", []string{"tag1", "tag2"}), clusters[2].Name)
	assert.Equal(t, envoy.BuildServiceKey("service2", []string{}), clusters[3].Name)
}
