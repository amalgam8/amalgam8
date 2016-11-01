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

package monitor

import (
	"testing"
	"time"

	"github.com/amalgam8/amalgam8/registry/client"
)

func TestCatalogComparison(t *testing.T) {
	r := registry{}

	cases := []struct {
		A, B  map[string][]*client.ServiceInstance
		Equal bool
	}{
		{
			A:     map[string][]*client.ServiceInstance{},
			B:     map[string][]*client.ServiceInstance{},
			Equal: true,
		},
		{ // TTL and heartbeat should be ignored when comparing
			A: map[string][]*client.ServiceInstance{
				"Service": []*client.ServiceInstance{
					{
						LastHeartbeat: time.Unix(0, 0),
						TTL:           1,
					},
				},
			},
			B: map[string][]*client.ServiceInstance{
				"Service": []*client.ServiceInstance{
					{
						LastHeartbeat: time.Unix(1, 0),
						TTL:           2,
					},
				},
			},
			Equal: true,
		},
		{
			A: map[string][]*client.ServiceInstance{
				"ServiceA": []*client.ServiceInstance{
					{
						ServiceName: "ServiceA",
					},
				},
			},
			B:     map[string][]*client.ServiceInstance{},
			Equal: false,
		},
		{
			A: map[string][]*client.ServiceInstance{
				"ServiceA": []*client.ServiceInstance{
					{
						ServiceName: "ServiceA",
					},
				},
			},
			B: map[string][]*client.ServiceInstance{
				"ServiceB": []*client.ServiceInstance{
					{
						ServiceName: "ServiceB",
					},
				},
			},
			Equal: false,
		},
	}
	for i, c := range cases {
		r.cache = c.A
		actual := r.compareToCache(c.B)
		if actual != c.Equal {
			t.Errorf("catalogsEqual(%v, %v): expected %v, got %v %d", c.A, c.B, c.Equal, actual, i)
		}
	}
}
