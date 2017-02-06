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

package envoy

import (
	"encoding/json"
	"testing"

	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestSanitizeRules(t *testing.T) {
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
					{
						Name: "service1",
						Tags: []string{"tag2", "tag1"},
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
						Tags: []string{"tag1"},
					},
				},
			},
		},
	}

	SanitizeRules(rules)

	assert.InEpsilon(t, 0.25, rules[0].Route.Backends[0].Weight, 0.01)
	assert.Equal(t, "service1", rules[0].Route.Backends[0].Name)
	assert.InEpsilon(t, 0.75, rules[0].Route.Backends[1].Weight, 0.01)
	assert.Equal(t, "service1", rules[0].Route.Backends[1].Name)
	assert.Len(t, rules[0].Route.Backends[1].Tags, 2)
	assert.Equal(t, "tag1", rules[0].Route.Backends[1].Tags[0])
	assert.Equal(t, "tag2", rules[0].Route.Backends[1].Tags[1])
	assert.InEpsilon(t, 1.00, rules[1].Route.Backends[0].Weight, 0.01)
	assert.Equal(t, "service2", rules[1].Route.Backends[0].Name)
}

func TestFS(t *testing.T) {
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

	instances := []api.ServiceInstance{
		{
			ServiceName: "service1",
			Endpoint: api.ServiceEndpoint{
				Type:  "tcp",
				Value: "10.0.0.1:80",
			},
			Tags: []string{},
		},
		{
			ServiceName: "service1",
			Endpoint: api.ServiceEndpoint{
				Type:  "tcp",
				Value: "10.0.0.2:80",
			},
			Tags: []string{"tag1"},
		},
		{
			ServiceName: "service1",
			Endpoint: api.ServiceEndpoint{
				Type:  "tcp",
				Value: "10.0.0.3:80",
			},
			Tags: []string{"tag2"},
		},
		{
			ServiceName: "service1",
			Endpoint: api.ServiceEndpoint{
				Type:  "tcp",
				Value: "10.0.0.4:80",
			},
			Tags: []string{"tag1", "tag2"},
		},
		{
			ServiceName: "service2",
			Endpoint: api.ServiceEndpoint{
				Type:  "https",
				Value: "10.0.0.5:80",
			},
		},
	}

	SanitizeRules(rules)
	rules = AddDefaultRouteRules(rules, instances)

	//err := buildFS(rules)
	//assert.NoError(t, err)
}

// TODO move this test to Discovery
//func TestBuildClusters(t *testing.T) {
//	rules := []api.Rule{
//		{
//			ID:          "abcdef",
//			Destination: "service1",
//			Route: &api.Route{
//				Backends: []api.Backend{
//					{
//						Name:   "service1",
//						Tags:   []string{"tag1"},
//						Weight: 0.25,
//					},
//				},
//			},
//		},
//		{
//			ID:          "abcdef",
//			Destination: "service1",
//			Route: &api.Route{
//				Backends: []api.Backend{
//					{
//						Name:   "service1",
//						Tags:   []string{"tag1", "tag2"},
//						Weight: 0.75,
//					},
//				},
//			},
//		},
//		{
//			ID:          "abcdef",
//			Destination: "service2",
//			Route: &api.Route{
//				Backends: []api.Backend{
//					{
//						Name:   "service2",
//						Tags:   []string{},
//						Weight: 1.00,
//					},
//				},
//			},
//		},
//		{
//			ID:          "abcdef",
//			Destination: "service2",
//			Actions:     []api.Action{},
//		},
//	}
//
//	clusters := buildClusters(rules)
//
//	assert.Len(t, clusters, 3)
//
//	clusterName := BuildServiceKey("service1", []string{"tag1"})
//	assert.Equal(t, Cluster{
//		Name:             clusterName,
//		ServiceName:      clusterName,
//		Type:             "sds",
//		LbType:           "round_robin",
//		ConnectTimeoutMs: 1000,
//		OutlierDetection: &OutlierDetection{
//			MaxEjectionPercent: 100,
//		},
//		CircuitBreaker: &CircuitBreaker{},
//	}, clusters[0])
//
//	assert.Equal(t, BuildServiceKey("service1", []string{"tag1", "tag2"}), clusters[1].Name)
//	assert.Equal(t, BuildServiceKey("service2", []string{}), clusters[2].Name)
//}

func TestBuildServiceKey(t *testing.T) {
	type TestCase struct {
		Service  string
		Tags     []string
		Expected string
	}

	testCases := []TestCase{
		{
			Service:  "serviceX",
			Tags:     []string{},
			Expected: "serviceX",
		},
		{
			Service:  "serviceX",
			Tags:     []string{"A=1"},
			Expected: "serviceX:A=1",
		},
		{
			Service:  "serviceX",
			Tags:     []string{"A=1", "B=2"},
			Expected: "serviceX:A=1,B=2",
		},
		{
			Service:  "serviceX",
			Tags:     []string{"A=1", "B=2", "C=3"},
			Expected: "serviceX:A=1,B=2,C=3",
		},
		{
			Service:  "serviceX",
			Tags:     []string{"B=2", "C=3", "A=1"},
			Expected: "serviceX:A=1,B=2,C=3",
		},
	}

	for _, testCase := range testCases {
		actual := BuildServiceKey(testCase.Service, testCase.Tags)
		assert.Equal(t, testCase.Expected, actual)
	}
}

func TestParseServiceKey(t *testing.T) {
	type TestCase struct {
		Service    string
		Tags       []string
		ServiceKey string
	}

	testCases := []TestCase{
		{
			Service:    "serviceX",
			Tags:       []string{},
			ServiceKey: "serviceX",
		},
		{
			Service:    "serviceX",
			Tags:       []string{"A=1"},
			ServiceKey: "serviceX:A=1",
		},
		{
			Service:    "serviceX",
			Tags:       []string{"A=1", "B=2"},
			ServiceKey: "serviceX:A=1,B=2",
		},
		{
			Service:    "serviceX",
			Tags:       []string{"A=1", "B=2", "C=3"},
			ServiceKey: "serviceX:A=1,B=2,C=3",
		},
	}

	for _, testCase := range testCases {
		serviceName, tags := ParseServiceKey(testCase.ServiceKey)
		assert.Equal(t, testCase.Service, serviceName)
		assert.Equal(t, testCase.Tags, tags)
	}
}

// Ensure that parse(build(s)) == s
func TestBuildParseServiceKey(t *testing.T) {
	type TestCase struct {
		Service string
		Tags    []string
	}

	testCases := []TestCase{
		{
			Service: "service1",
			Tags:    []string{},
		},
		{
			Service: "service2",
			Tags:    []string{"A"},
		},
		{
			Service: "service3",
			Tags:    []string{"A", "B", "C"},
		},
		{
			Service: "ser__vice4_",
			Tags:    []string{"A_", "_B", "_C_"},
		},
		{
			Service: "_service5__",
			Tags:    []string{"_A", "B", "C"},
		},
		{
			Service: "",
			Tags:    []string{},
		},
	}

	for _, testCase := range testCases {
		s := BuildServiceKey(testCase.Service, testCase.Tags)
		service, tags := ParseServiceKey(s)
		assert.Equal(t, testCase.Service, service)
		assert.Equal(t, testCase.Tags, tags)
	}
}

func TestConvert2(t *testing.T) {
	instances := []api.ServiceInstance{
		{
			ServiceName: "service1",
			Endpoint: api.ServiceEndpoint{
				Type:  "tcp",
				Value: "10.0.0.1:80",
			},
			Tags: []string{},
		},
		{
			ServiceName: "service1",
			Endpoint: api.ServiceEndpoint{
				Type:  "tcp",
				Value: "10.0.0.2:80",
			},
			Tags: []string{"tag1"},
		},
		{
			ServiceName: "service1",
			Endpoint: api.ServiceEndpoint{
				Type:  "tcp",
				Value: "10.0.0.3:80",
			},
			Tags: []string{"tag2"},
		},
		{
			ServiceName: "service1",
			Endpoint: api.ServiceEndpoint{
				Type:  "tcp",
				Value: "10.0.0.4:80",
			},
			Tags: []string{"tag1", "tag2"},
		},
		{
			ServiceName: "service2",
			Endpoint: api.ServiceEndpoint{
				Type:  "https",
				Value: "10.0.0.5:80",
			},
		},
	}

	rules := []api.Rule{
		{
			ID:          "abcdef",
			Destination: "service1",
			Route: &api.Route{
				Backends: []api.Backend{
					{
						Name: "service1",
						Tags: []string{"tag1"},
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
						Name: "service1",
						Tags: []string{"tag1", "tag2"},
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

	SanitizeRules(rules)
	rules = AddDefaultRouteRules(rules, instances)

	//configRoot, err := generateConfig(rules, instances, "gateway")
	//assert.NoError(t, err)

	//data, err := json.MarshalIndent(configRoot, "", "  ")
	//assert.NoError(t, err)
}

func TestBookInfo(t *testing.T) {
	ruleBytes := []byte(`[
    {
      "id": "ad95f5d6-fa7b-448d-8c27-928e40b37202",
      "priority": 2,
      "destination": "reviews",
      "match": {
        "headers": {
          "Cookie": "^(.*?;)?(user=jason)(;.*)?$"
        }
      },
      "route": {
        "backends": [
          {
            "tags": [
              "v2"
            ]
          }
        ]
      }
    },
    {
      "id": "e31da124-8394-4b12-9abf-ebdc7db679a9",
      "priority": 1,
      "destination": "details",
      "route": {
        "backends": [
          {
            "tags": [
              "v1"
            ]
          }
        ]
      }
    },
    {
      "id": "ab823eb5-e56c-485c-901f-0f29adfa8e4f",
      "priority": 1,
      "destination": "productpage",
      "route": {
        "backends": [
          {
            "tags": [
              "v1"
            ]
          }
        ]
      }
    },
    {
      "id": "03b97f82-40c5-4c51-8bf9-b1057a73019b",
      "priority": 1,
      "destination": "ratings",
      "route": {
        "backends": [
          {
            "tags": [
              "v1"
            ]
          }
        ]
      }
    },
    {
      "id": "c67226e2-8506-4e75-9e47-84d9d24f0326",
      "priority": 1,
      "destination": "reviews",
      "route": {
        "backends": [
          {
            "tags": [
              "v1"
            ]
          }
        ]
      }
    },
{
      "id": "c2a22912-9479-4e0b-839b-ffe76bb0c579",
      "priority": 10,
      "destination": "ratings",
      "match": {
        "headers": {
          "Cookie": "^(.*?;)?(user=jason)(;.*)?$"
        },
        "source": {
          "name": "reviews",
          "tags": [
            "v2"
          ]
        }
      },
      "actions": [
        {
          "action": "delay",
          "duration": 7,
          "probability": 1,
          "tags": [
            "v1"
          ]
        }
      ]
    }
  ]
`)

	instanceBytes := []byte(`[
    {
      "id": "74d2a394184327f5",
      "service_name": "productpage",
      "endpoint": {
        "type": "http",
        "value": "172.17.0.6:9080"
      },
      "ttl": 60,
      "status": "UP",
      "last_heartbeat": "2016-11-18T17:02:32.822819186Z",
      "tags": [
        "v1"
      ]
    },
    {
      "id": "26b250bc98d8a74c",
      "service_name": "ratings",
      "endpoint": {
        "type": "http",
        "value": "172.17.0.11:9080"
      },
      "ttl": 60,
      "status": "UP",
      "last_heartbeat": "2016-11-18T17:02:33.784740831Z",
      "tags": [
        "v1"
      ]
    },
    {
      "id": "9f7a75cdbbf492c7",
      "service_name": "details",
      "endpoint": {
        "type": "http",
        "value": "172.17.0.7:9080"
      },
      "ttl": 60,
      "status": "UP",
      "last_heartbeat": "2016-11-18T17:02:32.986290003Z",
      "tags": [
        "v1"
      ]
    },
    {
      "id": "05f853b7b4ab8b37",
      "service_name": "reviews",
      "endpoint": {
        "type": "http",
        "value": "172.17.0.10:9080"
      },
      "ttl": 60,
      "status": "UP",
      "last_heartbeat": "2016-11-18T17:02:33.559542468Z",
      "tags": [
        "v3"
      ]
    },
    {
      "id": "a4a740e9af065016",
      "service_name": "reviews",
      "endpoint": {
        "type": "http",
        "value": "172.17.0.8:9080"
      },
      "ttl": 60,
      "status": "UP",
      "last_heartbeat": "2016-11-18T17:02:33.18906562Z",
      "tags": [
        "v1"
      ]
    },
    {
      "id": "5f940f0ddee732bb",
      "service_name": "reviews",
      "endpoint": {
        "type": "http",
        "value": "172.17.0.9:9080"
      },
      "ttl": 60,
      "status": "UP",
      "last_heartbeat": "2016-11-18T17:02:33.349101984Z",
      "tags": [
        "v2"
      ]
    }
  ]`)
	var ruleList []api.Rule
	err := json.Unmarshal(ruleBytes, &ruleList)
	assert.NoError(t, err)

	var instances []api.ServiceInstance
	err = json.Unmarshal(instanceBytes, &instances)
	assert.NoError(t, err)

	//configRoot, err := generateConfig(ruleList, instances, "ratings")
	//assert.NoError(t, err)
	//
	//data, err := json.MarshalIndent(configRoot, "", "  ")
	//assert.NoError(t, err)
	//
	//fmt.Println(string(data))
}

func TestFaults(t *testing.T) {
	ruleBytes := []byte(`[{
      "id": "c2a22912-9479-4e0b-839b-ffe76bb0c510",
      "priority": 10,
      "destination": "ratings",
      "match": {
        "headers": {
          "Cookie": "^(.*?;)?(user=jason)(;.*)?$"
        },
        "source": {
          "name": "reviews",
          "tags": [
            "v2"
          ]
        }
      },
      "actions": [
        {
          "action": "delay",
          "duration": 7,
          "probability": 1,
          "tags": [
            "v1"
          ]
        }
      ]
    },
    {
      "id": "c2a22912-9479-4e0b-839b-ffe76bb0c579",
      "priority": 10,
      "destination": "ratings",
      "match": {
        "headers": {
          "Cookie": "^(.*?;)?(user=jason)(;.*)?$"
        },
        "source": {
          "name": "reviews",
          "tags": [
            "v1"
          ]
        }
      },
      "actions": [
        {
          "action": "delay",
          "duration": 7,
          "probability": 1,
          "tags": [
            "v1"
          ]
        }
      ]
    },
    {
      "id": "c2a22912-9479-4e0b-839b-ffe76bb0c579",
      "priority": 10,
      "destination": "ratings",
      "match": {
        "headers": {
          "test": "myval"
        },
        "source": {
          "name": "reviews",
          "tags": [
            "v1"
          ]
        }
      },
      "actions": [
        {
          "action": "abort",
          "return_code": 418,
          "probability": 0.8,
          "tags": [
  	    "v1"
          ]
        }
      ]
    },
    {
      "id": "c67226e2-8506-4e75-9e47-84d9d24f0326",
      "priority": 1,
      "destination": "reviews",
      "route": {
        "backends": [
          {
            "tags": [
              "v1"
            ]
          }
        ]
      }
    }]`)

	var ruleList []api.Rule
	err := json.Unmarshal(ruleBytes, &ruleList)
	assert.NoError(t, err)

	faults := buildFaults(ruleList, "reviews", []string{"v1"})

	assert.Len(t, faults, 3)
	for _, fault := range faults {
		assert.Equal(t, fault.Type, "decoder")
		switch fault.Name {
		case "fault":
			conf, ok := fault.Config.(*FilterFaultConfig)
			assert.True(t, ok)
			if conf.Abort != nil {
				assert.Equal(t, conf.Abort.Percent, 80)
				assert.Equal(t, conf.Abort.HTTPStatus, 418)
				assert.Len(t, conf.Headers, 1)
				assert.Equal(t, conf.Headers[0].Name, "test")
				assert.Equal(t, conf.Headers[0].Value, "myval")
			} else if conf.Delay != nil {
				assert.Equal(t, conf.Delay.Type, "fixed")
				assert.Equal(t, conf.Delay.Percent, 100)
				assert.Equal(t, conf.Delay.Duration, 7000)
				assert.Len(t, conf.Headers, 1)
				assert.Equal(t, conf.Headers[0].Name, "Cookie")
				assert.Equal(t, conf.Headers[0].Value, "^(.*?;)?(user=jason)(;.*)?$")
			} else {
				t.Fail()
			}
		case "router":
			conf, ok := fault.Config.(FilterRouterConfig)
			assert.True(t, ok)
			assert.False(t, conf.DynamicStats)
		default:
			t.Fail()
		}
	}

}
