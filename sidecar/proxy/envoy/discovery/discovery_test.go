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
