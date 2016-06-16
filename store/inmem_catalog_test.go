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
	"fmt"
	"math/rand"
	"testing"
	"time"

	"strings"

	"strconv"
	"sync"
	"sync/atomic"

	"github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
)

const (
	testMinTTL    = time.Duration(0) * time.Millisecond
	testMaxTTL    = time.Duration(10) * time.Minute
	testShortTTL  = time.Duration(100) * time.Millisecond
	testMediumTTL = time.Duration(3) * time.Second
)

func TestNewInMemoryCatalogNilConfiguration(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	assert.NotNil(t, catalog)

}

func TestNewInMemoryCatalogWithConfig(t *testing.T) {

	conf := createNewConfig(testShortTTL)
	catalog := newInMemoryCatalog(conf)

	assert.NotNil(t, catalog)

}

func TestEmptyCatalog(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	instances, err := catalog.List("Calc", nil)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceName, extractErrorCode(err))
	assert.Empty(t, instances)

}

func TestRegisterInstance(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)
	id, err := doRegister(catalog, instance)

	assert.NoError(t, err)
	assert.NotNil(t, id)
	assert.NotEmpty(t, id)

	instances, err := catalog.List("Calc", nil)

	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.EqualValues(t, true, assertContainsInstance(t, instances, instance))

}

func TestRegisterInstanceWithID(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	instance := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1:9080", Type: "tcp"},
		ID:          "some-nonsense-id",
	}

	registeredInstance, err := catalog.Register(instance)
	// We allow registration with ID for the replication
	assert.NoError(t, err)
	assert.NotNil(t, registeredInstance)
	assert.Equal(t, "some-nonsense-id", registeredInstance.ID)

}

func TestRegisterInstanceWithTTL(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	instance := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1:9080", Type: "tcp"},
		TTL:         time.Duration(15) * time.Second,
	}

	registeredInstance, err := catalog.Register(instance)

	assert.NoError(t, err)
	assert.NotNil(t, registeredInstance)
	assert.NotNil(t, registeredInstance.ID)
	assert.NotEmpty(t, registeredInstance.ID)
	assert.EqualValues(t, instance.TTL, registeredInstance.TTL)

}

func TestRegisterInstanceWithCatalogTTL(t *testing.T) {

	conf := createNewConfig(2 * DefaultConfig.DefaultTTL)
	catalog := newInMemoryCatalog(conf)

	instance := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1:9080", Type: "tcp"},
		TTL:         2 * DefaultConfig.DefaultTTL,
	}

	registeredInstance, err := catalog.Register(instance)

	assert.NoError(t, err)
	assert.NotNil(t, registeredInstance)
	assert.NotNil(t, registeredInstance.ID)
	assert.NotEmpty(t, registeredInstance.ID)
	assert.EqualValues(t, conf.defaultTTL, registeredInstance.TTL)

}

func TestRegisterInstanceWithoutTTL(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	instance := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1:9080", Type: "tcp"},
	}

	registeredInstance, err := catalog.Register(instance)

	assert.NoError(t, err)
	assert.NotNil(t, registeredInstance)
	assert.NotNil(t, registeredInstance.ID)
	assert.NotEmpty(t, registeredInstance.ID)
	assert.EqualValues(t, DefaultConfig.DefaultTTL, registeredInstance.TTL)

}

func TestRegisterInstanceWithTooLowTTL(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	instance := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1:9080", Type: "tcp"},
		TTL:         DefaultConfig.MinimumTTL / 2,
	}

	registeredInstance, err := catalog.Register(instance)

	assert.NoError(t, err)
	assert.NotNil(t, registeredInstance)
	assert.EqualValues(t, DefaultConfig.MinimumTTL, registeredInstance.TTL)

}

func TestRegisterInstanceWithTooHighTTL(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	instance := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1:9080", Type: "tcp"},
		TTL:         DefaultConfig.MaximumTTL * 2,
	}

	registeredInstance, err := catalog.Register(instance)

	assert.NoError(t, err)
	assert.NotNil(t, registeredInstance)
	assert.EqualValues(t, DefaultConfig.MaximumTTL, registeredInstance.TTL)

}

func TestRegisterInstanceSameServiceSameEndpointSameData(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	instance1 := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint: &Endpoint{
			Value: "192.168.0.1:9080",
			Type:  "tcp",
		},
		Status:   "UP",
		Metadata: []byte("metadata"),
	}
	instance2 := instance1.DeepClone()

	inst1, err1 := catalog.Register(instance1)
	assert.NoError(t, err1)
	assert.NotNil(t, inst1)

	inst2, err2 := catalog.Register(instance2)
	assert.NoError(t, err2)
	assert.NotNil(t, inst2)

	assertSameInstance(t, inst1, inst2)

	instances, err := catalog.List("Calc", nil)

	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assertContainsInstance(t, instances, inst1)
	assertContainsInstance(t, instances, inst2)

}

func TestRegisterInstanceSameServiceSameEndpointDifferentData(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	instance1 := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint: &Endpoint{
			Value: "192.168.0.1:9080",
			Type:  "tcp",
		},
		Status:   "UP",
		Metadata: []byte("metadata"),
	}
	instance2 := instance1.DeepClone()
	instance2.Status = "DOWN"
	instance2.Metadata = []byte("other-metadata")

	inst1, err1 := catalog.Register(instance1)
	assert.NoError(t, err1)
	assert.NotNil(t, inst1)

	inst2, err2 := catalog.Register(instance2)
	assert.NoError(t, err2)
	assert.NotNil(t, inst2)

	assert.NotEqual(t, inst1.Status, inst2.Status)
	assert.NotEqual(t, inst1.Metadata, inst2.Metadata)

	instances, err := catalog.List("Calc", nil)

	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assertContainsInstance(t, instances, inst2)

}

func TestRegisterInstanceSameServiceDifferentEndpoint(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	instance1 := newServiceInstance("Calc", "192.168.0.1", 9080)
	instance2 := instance1.DeepClone()
	instance2.Endpoint.Value = strings.Join([]string{"192.168.0.2", string(9080)}, ":")
	instance2.Endpoint.Type = "tcp"

	id1, _ := doRegister(catalog, instance1)
	id2, err := doRegister(catalog, instance2)

	assert.NoError(t, err)
	assert.NotEqual(t, id1, id2)

	instances, err := catalog.List("Calc", nil)

	assert.NoError(t, err)
	assert.Len(t, instances, 2)
	assertContainsInstance(t, instances, instance1)
	assertContainsInstance(t, instances, instance2)

}

func TestRegisterInstanceDifferentServiceSameEndpoint(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	instance1 := newServiceInstance("Calc1", "192.168.0.1", 9080)
	instance2 := instance1.DeepClone()
	instance2.ServiceName = "Calc2"

	id1, _ := doRegister(catalog, instance1)
	id2, err := doRegister(catalog, instance2)

	assert.NoError(t, err)
	assert.NotEqual(t, id1, id2)

	instances1, err := catalog.List("Calc1", nil)

	assert.NoError(t, err)
	assert.Len(t, instances1, 1)
	assertContainsInstance(t, instances1, instance1)

	instances2, err := catalog.List("Calc2", nil)

	assert.NoError(t, err)
	assert.Len(t, instances2, 1)
	assertContainsInstance(t, instances2, instance2)

}

func TestRegisterInstanceWithExtension(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	extension := map[string]interface{}{"key_str": "value1", "key_int": 7}
	instance1 := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint: &Endpoint{
			Value: "192.168.0.1:9080",
			Type:  "tcp",
		},
		Status:    "UP",
		Metadata:  []byte("metadata"),
		Extension: extension,
	}

	inst1, err1 := catalog.Register(instance1)
	assert.NoError(t, err1)
	assert.NotNil(t, inst1)

	instances1, err := catalog.List("Calc", nil)

	assert.NoError(t, err)
	assert.Len(t, instances1, 1)
	assertContainsInstance(t, instances1, inst1)
	assert.EqualValues(t, extension, inst1.Extension)
}

func TestListServices(t *testing.T) {

	cases := make(map[string]bool)
	cases["Calc1"] = true
	cases["Calc2"] = true
	cases["Calc3"] = true
	cases["Calc4"] = true

	catalog := newInMemoryCatalog(nil)

	for key := range cases {
		instance := newServiceInstance(key, "192.168.0.1", 9080)
		doRegister(catalog, instance)
	}

	services := catalog.ListServices(nil)
	for _, srv := range services {
		assert.NotNil(t, srv)
		assert.True(t, cases[srv.ServiceName])
	}
}

func TestOutOfServiceDoesNotExpire(t *testing.T) {
	conf := createNewConfig(testShortTTL)
	catalog := newInMemoryCatalog(conf)

	instance1 := &ServiceInstance{
		ServiceName: "Calc1",
		Endpoint: &Endpoint{
			Value: "192.168.0.1:9080",
			Type:  "tcp",
		},
		Status: "OUT_OF_SERVICE",
	}
	doRegister(catalog, instance1)

	instance2 := &ServiceInstance{
		ServiceName: "Calc1",
		Endpoint: &Endpoint{
			Value: "192.168.0.2:9080",
			Type:  "tcp",
		},
		Status: "STARTING",
	}
	doRegister(catalog, instance2)

	// Sleep past the ttl
	time.Sleep(testShortTTL * 2)

	instances, err := catalog.List("Calc1", nil)
	assert.NoError(t, err)

	// Should only have instance1.  Instance2 should have expired.
	assert.Len(t, instances, 1)
	assertContainsInstance(t, instances, instance1)

	// One more time just to be sure
	time.Sleep(testShortTTL * 2)

	instances, err = catalog.List("Calc1", nil)
	assert.NoError(t, err)

	// Should only have instance1.  Instance2 should have expired.
	assert.Len(t, instances, 1)
	assertContainsInstance(t, instances, instance1)
}

func TestReRegisterExpiredInstance(t *testing.T) {

	conf := createNewConfig(testShortTTL)
	catalog := newInMemoryCatalog(conf)

	instance1 := newServiceInstance("Calc", "192.168.0.1", 9080)

	id1, _ := doRegister(catalog, instance1)
	time.Sleep(testShortTTL * 2)

	instance2 := newServiceInstance("Calc", "192.168.0.1", 9080)
	id2, err := doRegister(catalog, instance2)

	assert.NoError(t, err)
	assert.NotNil(t, id2)
	assert.NotEmpty(t, id2)
	assert.Equal(t, id1, id2)

	instances, err := catalog.List("Calc", nil)

	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assertContainsInstance(t, instances, instance2)

}

func TestDeregisterInstance(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)

	id, _ := doRegister(catalog, instance)
	err := catalog.Deregister(id)

	assert.NoError(t, err)

	instances, err := catalog.List("Calc", nil)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceName, extractErrorCode(err))
	assert.Empty(t, instances)

}

func TestDeregisterInstanceMultipleServiceInstances(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	instance1 := newServiceInstance("Calc", "192.168.0.1", 9080)
	instance2 := instance1.DeepClone()
	instance2.Endpoint.Value = strings.Join([]string{"192.168.0.2", string(9080)}, ":")
	instance2.Endpoint.Type = "tcp"

	id1, _ := doRegister(catalog, instance1)
	doRegister(catalog, instance2)

	err := catalog.Deregister(id1)

	assert.NoError(t, err)

	instances, err := catalog.List("Calc", nil)

	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assertContainsInstance(t, instances, instance2)

}

func TestDeregisterInstanceNotRegistered(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	id := "service-ID"
	err := catalog.Deregister(id)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceInstance, extractErrorCode(err))

}

func TestDeregisterInstanceAlreadyDeregistered(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)

	id, _ := doRegister(catalog, instance)
	catalog.Deregister(id)

	err := catalog.Deregister(id)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceInstance, extractErrorCode(err))
}

func TestDeregisterInstanceAlreadyExpired(t *testing.T) {

	conf := createNewConfig(testShortTTL)
	catalog := newInMemoryCatalog(conf)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)

	id, _ := doRegister(catalog, instance)
	time.Sleep(testShortTTL * 2)

	err := catalog.Deregister(id)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceInstance, extractErrorCode(err))

}

func TestRenewInstance(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	instance := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1:9080", Type: "tcp"},
	}

	id, _ := doRegister(catalog, instance)
	err := catalog.Renew(id)

	assert.NoError(t, err)

}

func TestRenewInstanceNotRegistered(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	err := catalog.Renew("some-bogus-id")

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceInstance, extractErrorCode(err))

}

func TestRenewInstanceAlreadyDeregistered(t *testing.T) {

	catalog := newInMemoryCatalog(nil)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)

	id, _ := doRegister(catalog, instance)
	catalog.Deregister(id)
	err := catalog.Renew(id)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceInstance, extractErrorCode(err))

}

func TestRenewInstanceAlreadyExpired(t *testing.T) {

	conf := createNewConfig(testShortTTL)
	catalog := newInMemoryCatalog(conf)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)

	id, _ := doRegister(catalog, instance)
	time.Sleep(testShortTTL * 2)

	err := catalog.Renew(id)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceInstance, extractErrorCode(err))

}

func TestFindInstanceByID(t *testing.T) {
	catalog := newInMemoryCatalog(nil)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)
	id, _ := doRegister(catalog, instance)

	actual, _ := catalog.Instance(id)
	actual.LastRenewal = instance.LastRenewal
	actual.RegistrationTime = instance.RegistrationTime
	assert.Equal(t, instance, actual)
	assert.EqualValues(t, instance.ServiceName, actual.ServiceName)
	assert.EqualValues(t, instance.Endpoint.Value, actual.Endpoint.Value)
}

func TestSingleServiceQuota(t *testing.T) {
	var instanceID string
	namespaceCapacity := 10
	conf := &inmemConfig{defaultDefaultTTL, testMinTTL, testMaxTTL, namespaceCapacity}
	catalog := newInMemoryCatalog(conf)

	for i := 0; i < namespaceCapacity; i++ {
		instance := newServiceInstance("Calc", "192.168.0.1", uint32(9080+i))
		id, err := doRegister(catalog, instance)

		assert.NoError(t, err)
		assert.NotEmpty(t, id)
		instanceID = id
	}

	// register a new instance should fail
	instance := newServiceInstance("Calc", "192.168.0.1", 7080)
	id, err := doRegister(catalog, instance)
	assert.Error(t, err)
	assert.Empty(t, id)

	// add an existing instance and reattempt should succeed
	id, err = doRegister(catalog, instance)
	assert.Error(t, err)
	assert.Empty(t, id)

	// deregister instance and register a new one should succeed
	assert.NoError(t, catalog.Deregister(instanceID))
	id, err = doRegister(catalog, instance)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
}

func TestMultipleServicesQuota(t *testing.T) {
	var instanceID string
	namespaceCapacity := 10
	conf := &inmemConfig{defaultDefaultTTL, testMinTTL, testMaxTTL, namespaceCapacity}
	catalog := newInMemoryCatalog(conf)

	for i := 0; i < namespaceCapacity; i++ {
		instance := newServiceInstance(fmt.Sprintf("Calc_%d", i), "192.168.0.1", 9080)
		id, err := doRegister(catalog, instance)

		assert.NoError(t, err)
		assert.NotEmpty(t, id)
		instanceID = id
	}

	// register a new instance should fail
	instance := newServiceInstance("Calc", "192.168.0.1", 9080)
	id, err := doRegister(catalog, instance)
	assert.Error(t, err)
	assert.Empty(t, id)

	// add an existing instance and reattempt should succeed
	id, err = doRegister(catalog, instance)
	assert.Error(t, err)
	assert.Empty(t, id)

	// deregister instance and register a new one should succeed
	assert.NoError(t, catalog.Deregister(instanceID))
	id, err = doRegister(catalog, instance)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
}

func TestInstancesMetrics(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()

	conf := createNewConfig(testShortTTL)
	catalog := newInMemoryCatalog(conf)

	instancesCount := func() int64 {
		return metrics.DefaultRegistry.Get(instancesMetricName).(metrics.Counter).Count()
	}

	instance1 := newServiceInstance("Calc", "192.168.0.1", 9080)
	instance2 := newServiceInstance("Calc", "192.168.0.2", 9080)
	instance3 := newServiceInstance("Calc", "192.168.0.3", 9080)
	instance3.TTL = 4 * testShortTTL

	assert.EqualValues(t, 0, instancesCount())

	instance1, _ = catalog.Register(instance1)
	assert.EqualValues(t, 1, instancesCount())

	instance2, _ = catalog.Register(instance2)
	instance3, _ = catalog.Register(instance3)
	assert.EqualValues(t, 3, instancesCount())

	catalog.Deregister(instance1.ID)
	assert.EqualValues(t, 2, instancesCount())

	time.Sleep(testShortTTL * 2)
	assert.EqualValues(t, 1, instancesCount())
}

func TestAverageMetadataMetrics(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()
	catalog := newInMemoryCatalog(nil)

	averageMetadata := func() float64 {
		return metrics.DefaultRegistry.Get(metadataLengthMetricName).(metrics.Histogram).Mean()
	}

	instance1 := newServiceInstance("Calc", "192.168.0.1", 9080)
	instance2 := newServiceInstance("Calc", "192.168.0.2", 9080)
	instance3 := newServiceInstance("Calc", "192.168.0.3", 9080)

	instance1.Metadata = []byte("1234567890")                 // Length 10
	instance2.Metadata = []byte("abcdefghijklmnopqrstuvwxyz") // Length 26

	assert.EqualValues(t, 0, averageMetadata())

	catalog.Register(instance1)
	assert.EqualValues(t, 10, averageMetadata())

	catalog.Register(instance2)
	assert.EqualValues(t, 18, averageMetadata())

	// instance3, which doesn't have metadata, shouldn't affect the average
	catalog.Register(instance3)
	assert.EqualValues(t, 18, averageMetadata())
}

func TestAverageTagsMetrics(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()
	catalog := newInMemoryCatalog(nil)

	averageTags := func() float64 {
		return metrics.DefaultRegistry.Get(tagsLengthMetricName).(metrics.Histogram).Mean()
	}

	instance1 := newServiceInstance("Calc", "192.168.0.1", 9080)
	instance2 := newServiceInstance("Calc", "192.168.0.2", 9080)
	instance3 := newServiceInstance("Calc", "192.168.0.3", 9080)

	instance1.Tags = []string{"Tag1", "Tag2"}                 // 2 tag
	instance2.Tags = []string{"Tag1", "Tag2", "Tag3", "Tag4"} // 4 tags

	assert.EqualValues(t, 0, averageTags())

	catalog.Register(instance1)
	assert.EqualValues(t, 2, averageTags())

	catalog.Register(instance2)
	assert.EqualValues(t, 3, averageTags())

	// instance3, which doesn't have tags, shouldn't affect the average
	catalog.Register(instance3)
	assert.EqualValues(t, 3, averageTags())
}

func TestMetadataInstancesMetrics(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()
	catalog := newInMemoryCatalog(nil)

	metadataInstanceCount := func() int64 {
		return metrics.DefaultRegistry.Get(metadataInstancesMetricName).(metrics.Counter).Count()
	}

	instance1 := newServiceInstance("Calc", "192.168.0.1", 9080)
	instance2 := newServiceInstance("Calc", "192.168.0.2", 9080)

	assert.EqualValues(t, 0, metadataInstanceCount(), "Initial metadata count should be 0")

	instance1.Metadata = []byte("1234567890")
	instance1, _ = catalog.Register(instance1)
	assert.EqualValues(t, 1, metadataInstanceCount(), "Registering an instance with metadata should increase metadata count")

	instance1, _ = catalog.Register(instance1)
	assert.EqualValues(t, 1, metadataInstanceCount(), "Re-registering an instance with metadata should not affect metadata count")

	instance1.Metadata = []byte("12345678901234567890")
	instance1, _ = catalog.Register(instance1)
	assert.EqualValues(t, 1, metadataInstanceCount(), "Re-registering an instance with different metadata should not increase metadata count")

	instance1.Metadata = nil
	instance1, _ = catalog.Register(instance1)
	assert.EqualValues(t, 0, metadataInstanceCount(), "Re-registering an instance with removed metadata should decrease metadata count")

	instance1.Metadata = []byte("1234567890")
	instance1, _ = catalog.Register(instance1)
	assert.EqualValues(t, 1, metadataInstanceCount(), "Re-registering an instance with added metadata should increase metadata count")

	instance2, _ = catalog.Register(instance2)
	assert.EqualValues(t, 1, metadataInstanceCount(), "Registering an instance with no metadata should not affect metadata count")

	catalog.Deregister(instance2.ID)
	assert.EqualValues(t, 1, metadataInstanceCount(), "De-registering an instance with no metadata should not affect metadata count")

	catalog.Deregister(instance1.ID)
	assert.EqualValues(t, 0, metadataInstanceCount(), "De-registering an instance with metadata should decrease metadata count")
}

func TestTagsInstancesMetrics(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()
	catalog := newInMemoryCatalog(nil)

	tagsInstanceCount := func() int64 {
		return metrics.DefaultRegistry.Get(tagsInstancesMetricName).(metrics.Counter).Count()
	}

	instance1 := newServiceInstance("Calc", "192.168.0.1", 9080)
	instance2 := newServiceInstance("Calc", "192.168.0.2", 9080)

	assert.EqualValues(t, 0, tagsInstanceCount(), "Initial tags count should be 0")

	instance1.Tags = []string{"Tag1"}
	instance1, _ = catalog.Register(instance1)
	assert.EqualValues(t, 1, tagsInstanceCount(), "Registering an instance with tags should increase tags count")

	instance1, _ = catalog.Register(instance1)
	assert.EqualValues(t, 1, tagsInstanceCount(), "Re-registering an instance with tags should not affect tags count")

	instance1.Tags = []string{"Tag1", "Tag2"}
	instance1, _ = catalog.Register(instance1)
	assert.EqualValues(t, 1, tagsInstanceCount(), "Re-registering an instance with different tags should not increase tags count")

	instance1.Tags = nil
	instance1, _ = catalog.Register(instance1)
	assert.EqualValues(t, 0, tagsInstanceCount(), "Re-registering an instance with removed tags should decrease tags count")

	instance1.Tags = []string{"Tag1"}
	instance1, _ = catalog.Register(instance1)
	assert.EqualValues(t, 1, tagsInstanceCount(), "Re-registering an instance with added tags should increase tags count")

	instance2, _ = catalog.Register(instance2)
	assert.EqualValues(t, 1, tagsInstanceCount(), "Registering an instance with no tags should not affect tags count")

	catalog.Deregister(instance2.ID)
	assert.EqualValues(t, 1, tagsInstanceCount(), "De-registering an instance with no tags should not affect tags count")

	catalog.Deregister(instance1.ID)
	assert.EqualValues(t, 0, tagsInstanceCount(), "De-registering an instance with tags should decrease tags count")
}

func TestExpirationMetric(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()

	conf := createNewConfig(testShortTTL)
	catalog := newInMemoryCatalog(conf)

	expirationCount := func() int64 {
		return metrics.DefaultRegistry.Get(expirationMetricName).(metrics.Meter).Count()
	}
	assert.EqualValues(t, 0, expirationCount())

	catalog.Register(newServiceInstance("Calc", "192.168.0.1", 9080))
	catalog.Register(newServiceInstance("Calc", "192.168.0.2", 9080))
	catalog.Register(newServiceInstance("Calc", "192.168.0.3", 9080))
	assert.EqualValues(t, 0, expirationCount())

	time.Sleep(testShortTTL * 2)
	assert.EqualValues(t, 3, expirationCount())
}

func TestLifetimeMetric(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()

	conf := createNewConfig(testShortTTL)
	catalog := newInMemoryCatalog(conf)

	assertLifetime := func(d time.Duration) {
		const margin = 50 * time.Millisecond
		lifetime := time.Duration(metrics.DefaultRegistry.Get(lifetimeMetricName).(metrics.Histogram).Mean())

		assert.True(t, lifetime > d-margin, "Actual lifetime (%s) is greater than expected (%ds) by more than the allowed margin (%s)", lifetime, d, margin)
		assert.True(t, lifetime < d+margin, "Actual lifetime (%s) is lower than expected (%s) by more than the allowed margin (%s)", lifetime, d, margin)
	}

	assertLifetime(0)

	instance1 := newServiceInstance("Calc", "192.168.0.1", 9080)
	instance2 := newServiceInstance("Calc", "192.168.0.2", 9080)
	instance3 := newServiceInstance("Calc", "192.168.0.3", 9080)

	instance1.TTL = 1 * testShortTTL
	instance2.TTL = 3 * testShortTTL
	instance3.TTL = 100 * testShortTTL // but will be unregistered after 8 * testShortTTL

	instance1, _ = catalog.Register(instance1)
	instance2, _ = catalog.Register(instance2)
	instance3, _ = catalog.Register(instance3)
	assertLifetime(0)

	time.Sleep(testShortTTL * 2)
	assertLifetime(testShortTTL)

	time.Sleep(testShortTTL * 2)
	assertLifetime(2 * testShortTTL)

	time.Sleep(testShortTTL * 4)
	catalog.Deregister(instance3.ID)
	assertLifetime(4 * testShortTTL)
}

func TestMultiCatalogsMetrics(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()

	conf := createNewConfig(testShortTTL)
	catalog1 := newInMemoryCatalog(conf)
	catalog2 := newInMemoryCatalog(conf)

	instancesCount := func() int64 {
		return metrics.DefaultRegistry.Get(instancesMetricName).(metrics.Counter).Count()
	}

	catalog1.Register(newServiceInstance("Calc", "192.168.0.1", 9080))
	catalog2.Register(newServiceInstance("Calc", "192.168.0.2", 9080))

	assert.EqualValues(t, 2, instancesCount())
}

func TestInstanceExpiresDefaultTTL(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ttl := DefaultConfig.DefaultTTL
	catalog := newInMemoryCatalog(nil)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)
	instance, _ = catalog.Register(instance)

	done := time.After(ttl)
Outer:
	for {
		select {
		case <-done:
			break Outer
		default:
			{
				instances, err := catalog.List("Calc", nil)
				assert.NoError(t, err)
				assertContainsInstance(t, instances, instance)
				time.Sleep(1 * time.Second)
			}
		}
	}

	// Grace period
	time.Sleep(1 * time.Second)

	instances, err := catalog.List("Calc", nil)
	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceName, extractErrorCode(err))
	assert.Empty(t, instances)

}

func TestInstanceExpiresCatalogTTL(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ttl := time.Duration(5) * time.Second
	conf := createNewConfig(ttl)
	catalog := newInMemoryCatalog(conf)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)
	instance, _ = catalog.Register(instance)

	done := time.After(ttl)
Outer:
	for {
		select {
		case <-done:
			break Outer
		default:
			{
				instances, err := catalog.List("Calc", nil)
				assert.NoError(t, err)
				assertContainsInstance(t, instances, instance)
				time.Sleep(1 * time.Second)
			}
		}
	}

	// Grace period
	time.Sleep(1 * time.Second)

	instances, err := catalog.List("Calc", nil)
	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceName, extractErrorCode(err))
	assert.Empty(t, instances)

}

func TestInstanceExpireInstanceTTL(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ttl := time.Duration(5) * time.Second
	catalog := newInMemoryCatalog(nil)

	instance := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1:9080", Type: "tcp"},
		TTL:         ttl,
	}
	instance, _ = catalog.Register(instance)

	done := time.After(ttl)
Outer:
	for {
		select {
		case <-done:
			break Outer
		default:
			{
				instances, err := catalog.List("Calc", nil)
				assert.NoError(t, err)
				assertContainsInstance(t, instances, instance)
				time.Sleep(1 * time.Second)
			}
		}
	}

	// Grace period
	time.Sleep(1 * time.Second)

	instances, err := catalog.List("Calc", nil)
	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceName, extractErrorCode(err))
	assert.Empty(t, instances)

}

func TestRegisterInstancesSameServiceConcurrently(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	const numOfInstances = 10000

	ttl := time.Duration(30) * time.Second
	conf := createNewConfig(ttl)
	catalog := newInMemoryCatalog(conf)

	var wg sync.WaitGroup
	wg.Add(numOfInstances)

	for i := 0; i < numOfInstances; i++ {

		hostSuffix := strconv.Itoa(i)
		host := "192.168.0." + hostSuffix

		go func() {

			defer wg.Done()
			time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)

			instance := newServiceInstance("Calc", host, 9080)

			id, err := doRegister(catalog, instance)

			assert.NoError(t, err, "Error registering host %v: %v", host, err)
			assert.NotNil(t, id, "Nil ID for host %v ", host)
			assert.NotEmpty(t, id, "Empty ID for host %v", host)

		}()

	}

	wg.Wait()

	instances, err := catalog.List("Calc", nil)
	assert.NoError(t, err)
	assert.Len(t, instances, numOfInstances)

}

func TestRegisterInstancesConcurrently(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// Test assumes numOfInstances % numOfServices == 0
	const numOfServices = 50
	const numOfInstances = 10000

	ttl := time.Duration(30) * time.Second
	conf := createNewConfig(ttl)
	catalog := newInMemoryCatalog(conf)

	var wg sync.WaitGroup
	wg.Add(numOfInstances)

	for i := 0; i < numOfInstances; i++ {

		hostSuffix := strconv.Itoa(i)
		host := "192.168.0." + hostSuffix
		service := "Calc" + strconv.Itoa(i%numOfServices)

		go func() {

			defer wg.Done()
			time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)

			instance := newServiceInstance(service, host, 9080)

			id, err := doRegister(catalog, instance)

			assert.NoError(t, err, "Error registering host %v for service '%v': %v", host, service, err)
			assert.NotNil(t, id, "Nil ID for host %v on service '%v'", host, service)
			assert.NotEmpty(t, id, "Empty ID for host %v on service '%v'", host, service)

		}()

	}

	wg.Wait()

	for i := 0; i < numOfServices; i++ {

		service := "Calc" + strconv.Itoa(i)
		instances, err := catalog.List(service, nil)

		assert.NoError(t, err, "Error getting instances list of '%v': %v", service, err)
		assert.Len(t, instances, numOfInstances/numOfServices)

	}

}

func TestDeregisterInstancesSameServiceConcurrently(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	const numOfInstances = 10000

	ttl := time.Duration(30) * time.Second
	conf := createNewConfig(ttl)
	catalog := newInMemoryCatalog(conf)
	var ids [numOfInstances]string

	for i := 0; i < numOfInstances; i++ {

		hostSuffix := strconv.Itoa(i)
		host := "192.168.0." + hostSuffix
		instance := newServiceInstance("Calc", host, 9080)

		id, _ := doRegister(catalog, instance)
		ids[i] = id

	}

	var wg sync.WaitGroup
	wg.Add(numOfInstances)

	for i := 0; i < numOfInstances; i++ {

		id := ids[i]

		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)
			err := catalog.Deregister(id)
			assert.NoError(t, err, "Error deregistering instance ID %v: %v", id, err)
		}()

	}

	wg.Wait()

	instances, err := catalog.List("Calc", nil)
	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceName, extractErrorCode(err))
	assert.Empty(t, instances)

}

func TestDeregisterInstancesConcurrently(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// Test assumes numOfInstances % numOfServices == 0
	const numOfServices = 50
	const numOfInstances = 10000

	ttl := time.Duration(30) * time.Second
	conf := createNewConfig(ttl)
	catalog := newInMemoryCatalog(conf)
	var ids [numOfInstances]string

	for i := 0; i < numOfInstances; i++ {

		hostSuffix := strconv.Itoa(i)
		host := "192.168.0." + hostSuffix
		service := "Calc" + strconv.Itoa(i%numOfServices)
		instance := newServiceInstance(service, host, 9080)

		id, _ := doRegister(catalog, instance)
		ids[i] = id

	}

	var wg sync.WaitGroup
	wg.Add(numOfInstances)

	for i := 0; i < numOfInstances; i++ {

		id := ids[i]

		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)
			err := catalog.Deregister(id)
			assert.NoError(t, err, "Error deregistering instance ID %v: %v", id, err)
		}()

	}

	wg.Wait()

	for i := 0; i < numOfServices; i++ {

		service := "Calc" + strconv.Itoa(i)
		instances, err := catalog.List(service, nil)

		assert.Error(t, err, "Expected error getting instances list of non-existent service '%v'", service)
		assert.EqualValues(t, ErrorNoSuchServiceName, extractErrorCode(err), "Wrong error for service '%v': %v", service, err)
		assert.Empty(t, instances, "Non-empty instances list for service '%v': %v instances left", service, len(instances))

	}

}

func TestRenewInstancePreventExpiration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	conf := createNewConfig(testShortTTL)
	catalog := newInMemoryCatalog(conf)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)

	id, _ := doRegister(catalog, instance)

	done := make(chan bool)
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(testShortTTL / 2)
			err := catalog.Renew(id)
			assert.NoError(t, err)
		}
		done <- true
	}()

Outer:
	for {
		select {
		case <-done:
			break Outer
		default:
			{
				time.Sleep(testShortTTL / 2)
				instances, err := catalog.List("Calc", nil)
				assert.NoError(t, err)
				assert.Len(t, instances, 1)
				assertContainsInstance(t, instances, instance)
			}
		}
	}

}

func TestRenewInstancesConcurrently(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	if testing.Short() {
		t.SkipNow()
	}

	const numOfInstances = 10000
	const numOfIterations = 10
	const ttl = testMediumTTL

	conf := createNewConfig(ttl)
	catalog := newInMemoryCatalog(conf)

	var ids [numOfInstances]string
	for i := 0; i < numOfInstances; i++ {
		hostSuffix := strconv.Itoa(i)
		host := "192.168.0." + hostSuffix
		instance := newServiceInstance("Calc", host, 9080)
		id, _ := doRegister(catalog, instance)
		ids[i] = id
	}

	var wg sync.WaitGroup
	wg.Add(numOfInstances)

	done := make(chan struct{})
	for i := 0; i < numOfInstances; i++ {
		id := ids[i]
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					{
						duration := randPercentOfDuration(0.25, 0.5, ttl)
						time.Sleep(duration)
						err := catalog.Renew(id)
						assert.NoError(t, err, "Renewal of instance %v failed: %v", id, err)

					}
				}
			}
		}()
	}

	time.Sleep(ttl * numOfIterations)

	close(done)
	wg.Wait()

	instances, err := catalog.List("Calc", nil)
	assert.NoError(t, err)
	assert.Len(t, instances, numOfInstances, "Expected %v surviving instances, found %v", numOfInstances, len(instances))

}

func TestRenewAndExpireInstancesConcurrently(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	if testing.Short() {
		t.SkipNow()
	}

	const numOfInstances = 10000
	const numOfIterations = 10
	const expirationProbability = 0.2
	const ttl = testMediumTTL

	conf := createNewConfig(ttl)
	catalog := newInMemoryCatalog(conf)

	var ids [numOfInstances]string
	for i := 0; i < numOfInstances; i++ {
		hostSuffix := strconv.Itoa(i)
		host := "192.168.0." + hostSuffix
		instance := newServiceInstance("Calc", host, 9080)
		id, _ := doRegister(catalog, instance)
		ids[i] = id
	}

	var wg sync.WaitGroup
	wg.Add(numOfInstances)

	var expiredInstances uint32
	stopExpiration := make(chan struct{})
	done := make(chan struct{})
	for i := 0; i < numOfInstances; i++ {
		id := ids[i]
		go func() {
			defer wg.Done()
			expirationAllowed := true
			stopExpirationLocal := stopExpiration
			for {
				select {
				case <-done:
					return
				case <-stopExpirationLocal:
					{
						expirationAllowed = false
						stopExpirationLocal = nil
					}
				default:
					{
						if expirationAllowed && rand.Float32() <= expirationProbability {
							atomic.AddUint32(&expiredInstances, 1)
							return
						}
						duration := randPercentOfDuration(0.25, 0.5, ttl)
						time.Sleep(duration)
						err := catalog.Renew(id)
						assert.NoError(t, err, "Renewal of instance %v failed: %v", id, err)
					}
				}
			}
		}()
	}

	time.Sleep(ttl * (numOfIterations - 2))
	close(stopExpiration)
	time.Sleep(ttl * 2)
	close(done)
	wg.Wait()

	survivingInstances := numOfInstances - expiredInstances

	instances, err := catalog.List("Calc", nil)
	assert.NoError(t, err)
	assert.Len(t, instances, int(survivingInstances), "Expected %v surviving instances, found %v", survivingInstances, len(instances))

}

func newServiceInstance(name string, host string, port uint32) *ServiceInstance {
	return &ServiceInstance{
		ServiceName: name,
		Endpoint: &Endpoint{
			Type:  "tcp",
			Value: strings.Join([]string{host, fmt.Sprint(port)}, ":"),
		},
	}
}

func doRegister(catalog Catalog, si *ServiceInstance) (string, error) {
	registeredSI, err := catalog.Register(si)
	if err == nil {
		si.ID = registeredSI.ID
		si.TTL = registeredSI.TTL
	}
	return si.ID, err
}

func createNewConfig(defaultTTL time.Duration) *inmemConfig {
	return &inmemConfig{defaultTTL, testMinTTL, testMaxTTL, -1}
}

// This function is used instead of the assert.Contains because the instance
// is NOT exactly the same in the replicated catalogs due to the LastRenewal field
func assertContainsInstance(t *testing.T, instances []*ServiceInstance, si *ServiceInstance) bool {

	// A non-nil array contains the nil instance
	if si == nil {
		return true
	}

	for _, inst := range instances {
		if inst.ID == si.ID {
			return assertSameInstance(t, si, inst)
		}
	}

	return assert.Fail(t, "expected instance %v doesn't exist within %v", si, instances)

}

func assertSameInstance(t *testing.T, expected, actual *ServiceInstance) bool {
	same := true
	same = same && assert.EqualValues(t, expected.ID, actual.ID)
	same = same && assert.EqualValues(t, expected.ServiceName, actual.ServiceName)
	same = same && assert.EqualValues(t, expected.Endpoint, actual.Endpoint)
	same = same && assert.EqualValues(t, expected.Status, actual.Status)
	same = same && assert.EqualValues(t, expected.TTL, actual.TTL)
	same = same && assert.EqualValues(t, expected.Metadata, actual.Metadata)
	return same
}

func randPercent(low, high float64) float64 {
	if low < 0.0 || low >= 1.0 || high < 0.0 || high >= 1.0 || low > high {
		panic("invalid arguments to randPercent")
	}
	return low + ((high - low) * rand.Float64())
}

func randPercentOfDuration(low, high float64, duration time.Duration) time.Duration {
	percent := randPercent(low, high)
	return time.Duration(percent * float64(duration))
}

func extractErrorCode(err error) ErrorCode {
	return err.(*Error).Code
}
