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

package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/utils/database"
)

func TestRedisRegistryInsertReadDelete(t *testing.T) {
	db := database.NewMockDB()
	redisreg := NewRedisRegistry(db)

	var ns auth.Namespace
	ns = "namespace"

	si := &ServiceInstance{
		ID:          "inst-id",
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1", Type: "tcp"},
		Status:      "UP",
	}

	err := redisreg.InsertServiceInstance(ns, si)

	assert.NoError(t, err)

	// Try an instance id that doesn't exist
	actualSi, err := redisreg.ReadServiceInstanceByInstID(ns, "noexist")
	assert.NoError(t, err)
	assert.Nil(t, actualSi)

	actualSi, err = redisreg.ReadServiceInstanceByInstID(ns, "inst-id")
	var defaultTime time.Time
	actualSi.RegistrationTime = defaultTime
	actualSi.LastRenewal = defaultTime

	assert.NoError(t, err)
	assert.Equal(t, *si, *actualSi)

	// Try to delete an instance that doesn't exist
	del, err := redisreg.DeleteServiceInstance(ns, "inst-id000")
	assert.NoError(t, err)
	assert.Equal(t, 0, del)

	del, err = redisreg.DeleteServiceInstance(ns, "inst-id")
	assert.NoError(t, err)
	assert.Equal(t, 1, del)
}

func TestRedisRegistryReadKeys(t *testing.T) {
	db := database.NewMockDB()
	redisreg := NewRedisRegistry(db)

	var ns, ns2 auth.Namespace
	ns = "namespace"
	ns2 = "namespace2"

	si := &ServiceInstance{
		ID:          "inst-id",
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1", Type: "tcp"},
		Status:      "UP",
	}

	err := redisreg.InsertServiceInstance(ns, si)
	assert.NoError(t, err)

	keys, err := redisreg.ReadKeys(ns)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(keys))

	si2 := &ServiceInstance{
		ID:          "inst-id2",
		ServiceName: "Calc2",
		Endpoint:    &Endpoint{Value: "192.168.0.2", Type: "tcp"},
		Status:      "UP",
	}

	err = redisreg.InsertServiceInstance(ns, si2)
	assert.NoError(t, err)

	keys, err = redisreg.ReadKeys(ns)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(keys))

	si3 := &ServiceInstance{
		ID:          "inst-id3",
		ServiceName: "Calc3",
		Endpoint:    &Endpoint{Value: "192.168.0.3", Type: "tcp"},
		Status:      "UP",
	}

	err = redisreg.InsertServiceInstance(ns2, si3)
	assert.NoError(t, err)

	keys, err = redisreg.ReadKeys(ns)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(keys))

	keys, err = redisreg.ReadKeys(ns2)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(keys))
}

func TestRedisRegistryListServiceInstancesByName(t *testing.T) {
	db := database.NewMockDB()
	redisreg := NewRedisRegistry(db)

	var ns auth.Namespace
	ns = "namespace"

	key1 := DBKey{Namespace: ns.String(), InstanceID: "inst-id"}
	si := &ServiceInstance{
		ID:          "inst-id",
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1", Type: "tcp"},
		Status:      "UP",
	}

	err := redisreg.InsertServiceInstance(ns, si)
	assert.NoError(t, err)

	key2 := DBKey{Namespace: ns.String(), InstanceID: "inst-id2"}
	si2 := &ServiceInstance{
		ID:          "inst-id2",
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.2", Type: "tcp"},
		Status:      "UP",
	}

	err = redisreg.InsertServiceInstance(ns, si2)
	assert.NoError(t, err)

	si3 := &ServiceInstance{
		ID:          "inst-id3",
		ServiceName: "Calc2",
		Endpoint:    &Endpoint{Value: "192.168.0.3", Type: "tcp"},
		Status:      "UP",
	}

	err = redisreg.InsertServiceInstance(ns, si3)
	assert.NoError(t, err)

	serviceList, err := redisreg.ListServiceInstancesByName(ns, "Calc")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(serviceList))
	for _, service := range serviceList {
		var defaultTime time.Time
		service.RegistrationTime = defaultTime
		service.LastRenewal = defaultTime
	}
	assert.Equal(t, *si, *serviceList[key1.String()])
	assert.Equal(t, *si2, *serviceList[key2.String()])
}

func TestRedisRegistryListAllServiceInstances(t *testing.T) {
	db := database.NewMockDB()
	redisreg := NewRedisRegistry(db)

	var ns, ns2 auth.Namespace
	ns = "namespace"
	ns2 = "namespace"

	key1 := DBKey{Namespace: ns.String(), InstanceID: "inst-id"}
	si := &ServiceInstance{
		ID:          "inst-id",
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1", Type: "tcp"},
		Status:      "UP",
	}

	err := redisreg.InsertServiceInstance(ns, si)
	assert.NoError(t, err)

	key2 := DBKey{Namespace: ns.String(), InstanceID: "inst-id2"}
	si2 := &ServiceInstance{
		ID:          "inst-id2",
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.2", Type: "tcp"},
		Status:      "UP",
	}

	err = redisreg.InsertServiceInstance(ns, si2)
	assert.NoError(t, err)

	si3 := &ServiceInstance{
		ID:          "inst-id3",
		ServiceName: "Calc2",
		Endpoint:    &Endpoint{Value: "192.168.0.3", Type: "tcp"},
		Status:      "UP",
	}

	err = redisreg.InsertServiceInstance(ns, si3)
	assert.NoError(t, err)

	si4 := &ServiceInstance{
		ID:          "inst-id4",
		ServiceName: "Calc-ns2",
		Endpoint:    &Endpoint{Value: "192.168.0.4", Type: "tcp"},
		Status:      "UP",
	}

	err = redisreg.InsertServiceInstance(ns2, si4)
	assert.NoError(t, err)

	serviceInstances, err := redisreg.ListAllServiceInstances(ns)
	assert.NoError(t, err)

	for _, serviceMap := range serviceInstances {
		for _, service := range serviceMap {
			var defaultTime time.Time
			service.RegistrationTime = defaultTime
			service.LastRenewal = defaultTime
		}
	}

	assert.Equal(t, 3, len(serviceInstances))
	assert.Equal(t, 2, len(serviceInstances["Calc"]))
	assert.Equal(t, 1, len(serviceInstances["Calc2"]))
	assert.Equal(t, *si, *serviceInstances["Calc"][key1])
	assert.Equal(t, *si2, *serviceInstances["Calc"][key2])
}
