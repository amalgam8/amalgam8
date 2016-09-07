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
	"strings"
	"testing"
	"time"

	"strconv"
	"sync"
	"sync/atomic"

	"github.com/amalgam8/registry/utils/logging"
	"github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
)

const (
	defaultTTL = time.Duration(30) * time.Second
	minimumTTL = time.Duration(10) * time.Second
	maximumTTL = time.Duration(10) * time.Minute
)

func createExternalCatalog(conf *externalConfig, db ExternalRegistry) *externalCatalog {
	counterFactory := func() metrics.Counter { return metrics.NewCounter() }
	meterFactory := func() metrics.Meter { return metrics.NewMeter() }
	histogramFactory := func() metrics.Histogram { return metrics.NewHistogram(metrics.NewExpDecaySample(1028, 0.015)) }

	catalog := &externalCatalog{
		conf:      conf,
		logger:    logging.GetLogger(module),
		namespace: "test",
		db:        db,

		instancesMetric:         metrics.GetOrRegister(instancesMetricName, counterFactory).(metrics.Counter),
		expirationMetric:        metrics.GetOrRegister(expirationMetricName, meterFactory).(metrics.Meter),
		lifetimeMetric:          metrics.GetOrRegister(lifetimeMetricName, histogramFactory).(metrics.Histogram),
		metadataLengthMetric:    metrics.GetOrRegister(metadataLengthMetricName, histogramFactory).(metrics.Histogram),
		metadataInstancesMetric: metrics.GetOrRegister(metadataInstancesMetricName, counterFactory).(metrics.Counter),
		tagsLengthMetric:        metrics.GetOrRegister(tagsLengthMetricName, histogramFactory).(metrics.Histogram),
		tagsInstancesMetric:     metrics.GetOrRegister(tagsInstancesMetricName, counterFactory).(metrics.Counter),
	}

	return catalog
}

func setupCatalogForTest() *externalCatalog {
	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{defaultTTL, minimumTTL, maximumTTL, 50, "redis", "testaddress", "testpassword", nil}
	return createExternalCatalog(conf, db)
}

func TestNewExternalCatalogNotRedis(t *testing.T) {
	conf := createNewExternalConfig(testShortTTL)
	conf.store = "test"
	catalog, err := newExternalCatalog(conf, "test", nil)

	if assert.Error(t, err, "An error was expected") {
		expectedError := fmt.Errorf("External store test is not supported")
		assert.Equal(t, expectedError, err)
	}
	assert.Equal(t, nil, catalog)
}

func TestNewExternalCatalogNilConfig(t *testing.T) {
	catalog, err := newExternalCatalog(nil, "test", nil)

	if assert.Error(t, err, "An error was expected") {
		expectedError := fmt.Errorf("Config cannot be nil")
		assert.Equal(t, expectedError, err)
	}
	assert.Equal(t, nil, catalog)
}

func TestExternalRegisterValidateParams(t *testing.T) {
	conf := createNewExternalConfig(testShortTTL)
	catalog := createExternalCatalog(conf, nil)

	string33 := "123456789012345678901234567890123"
	string65 := "12345678901234567890123456789012345678901234567890123456789012345"
	string1025 := "12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345"

	// Error when no Service Name
	si := newServiceInstance("", "192.168.0.1", 9080)

	_, err := catalog.Register(si)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorBadRequest, extractErrorCode(err))
	assert.EqualValues(t, "Empty service name", err.(*Error).Message)

	// Error when Service Name is too long (>64 bytes)
	si = newServiceInstance(string65, "192.168.0.1", 9080)

	_, err = catalog.Register(si)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorBadRequest, extractErrorCode(err))
	assert.EqualValues(t, "Service name length too long", err.(*Error).Message)

	// Error when Endpoint value too long (>64 bytes)
	si = &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: string65, Type: "tcp"},
	}

	_, err = catalog.Register(si)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorBadRequest, extractErrorCode(err))
	assert.EqualValues(t, "Endpoint value length too long", err.(*Error).Message)

	// Error when status too long (>32 bytes)
	si = &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1", Type: "tcp"},
		Status:      string33,
	}

	_, err = catalog.Register(si)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorBadRequest, extractErrorCode(err))
	assert.EqualValues(t, "Status length too long", err.(*Error).Message)

	// Error when metadata too long (>1024 bytes)
	si = &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1", Type: "tcp"},
		Metadata:    []byte(string1025),
	}

	_, err = catalog.Register(si)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorBadRequest, extractErrorCode(err))
	assert.EqualValues(t, "Metadata length too long", err.(*Error).Message)
}

func TestExternalRegisterInstanceWithID(t *testing.T) {
	catalog := setupCatalogForTest()

	si := &ServiceInstance{
		ID:          "inst1",
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1", Type: "tcp"},
	}
	registeredInstance, err := catalog.Register(si)

	// We allow registration with ID for the replication
	assert.NoError(t, err)
	assert.NotNil(t, registeredInstance)
	assert.Equal(t, "inst1", registeredInstance.ID)
}

func TestExternalRegisterInstanceWithTTL(t *testing.T) {
	catalog := setupCatalogForTest()

	si := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1:9080", Type: "tcp"},
		TTL:         time.Duration(15) * time.Second,
	}

	registeredInstance, err := catalog.Register(si)

	assert.NoError(t, err)
	assert.NotNil(t, registeredInstance)
	assert.NotNil(t, registeredInstance.ID)
	assert.NotEmpty(t, registeredInstance.ID)
	assert.EqualValues(t, si.TTL, registeredInstance.TTL)

}

func TestRedisRegisterInstanceWithCatalogTTL(t *testing.T) {
	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{2 * DefaultConfig.DefaultTTL, minimumTTL, maximumTTL, 50, "redis", "testaddress", "testpassword", nil}
	catalog := createExternalCatalog(conf, db)

	si := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1:9080", Type: "tcp"},
		TTL:         2 * DefaultConfig.DefaultTTL,
	}

	registeredInstance, err := catalog.Register(si)

	assert.NoError(t, err)
	assert.NotNil(t, registeredInstance)
	assert.NotNil(t, registeredInstance.ID)
	assert.NotEmpty(t, registeredInstance.ID)
	assert.EqualValues(t, conf.defaultTTL, registeredInstance.TTL)
}

func TestExternalRegisterInstanceWithoutTTL(t *testing.T) {
	catalog := setupCatalogForTest()

	si := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1:9080", Type: "tcp"},
	}

	registeredInstance, err := catalog.Register(si)

	assert.NoError(t, err)
	assert.NotNil(t, registeredInstance)
	assert.NotNil(t, registeredInstance.ID)
	assert.NotEmpty(t, registeredInstance.ID)
	assert.EqualValues(t, defaultTTL, registeredInstance.TTL)
}

func TestExternalRegisterInstanceWithTooLowTTL(t *testing.T) {
	catalog := setupCatalogForTest()

	si := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1:9080", Type: "tcp"},
		TTL:         minimumTTL / 2,
	}

	registeredInstance, err := catalog.Register(si)

	assert.NoError(t, err)
	assert.NotNil(t, registeredInstance)
	assert.EqualValues(t, minimumTTL, registeredInstance.TTL)
}

func TestExternalRegisterInstanceWithTooHighTTL(t *testing.T) {
	catalog := setupCatalogForTest()

	si := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1:9080", Type: "tcp"},
		TTL:         maximumTTL * 2,
	}

	registeredInstance, err := catalog.Register(si)

	assert.NoError(t, err)
	assert.NotNil(t, registeredInstance)
	assert.EqualValues(t, maximumTTL, registeredInstance.TTL)
}

func TestExternalRegisterInstanceSameServiceSameEndpointSameData(t *testing.T) {
	catalog := setupCatalogForTest()

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

func TestExternalRegisterInstanceSameServiceSameEndpointDifferentData(t *testing.T) {
	catalog := setupCatalogForTest()

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
	instance2.Status = "OUT_OF_SERVICE"
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

func TestExternalRegisterInstanceSameServiceDifferentEndpoint(t *testing.T) {
	catalog := setupCatalogForTest()

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

func TestExternalRegisterInstanceDifferentServiceSameEndpoint(t *testing.T) {
	catalog := setupCatalogForTest()

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

func TestExternalRegisterInstanceWithExtension(t *testing.T) {
	catalog := setupCatalogForTest()

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

func TestExternalListServices(t *testing.T) {
	cases := make(map[string]bool)
	cases["Calc1"] = true
	cases["Calc2"] = true
	cases["Calc3"] = true
	cases["Calc4"] = true

	catalog := setupCatalogForTest()

	for key := range cases {
		instance := newServiceInstance(key, "192.168.0.1", 9080)
		doRegister(catalog, instance)
	}

	services := catalog.ListServices(nil)

	assert.Len(t, services, 4)
	for _, srv := range services {
		assert.NotNil(t, srv)
		assert.True(t, cases[srv.ServiceName])
	}
}

func TestExternalOutOfServiceDoesNotExpire(t *testing.T) {
	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{testShortTTL, testShortTTL, maximumTTL, 50, "redis", "testaddress", "testpassword", nil}
	catalog := createExternalCatalog(conf, db)

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

func TestExternalReRegisterExpiredInstance(t *testing.T) {
	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{testShortTTL, testShortTTL, maximumTTL, 50, "redis", "testaddress", "testpassword", nil}
	catalog := createExternalCatalog(conf, db)

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

func TestExternalDeregisterInstance(t *testing.T) {
	catalog := setupCatalogForTest()

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)

	id, _ := doRegister(catalog, instance)
	dInstance, err := catalog.Deregister(id)

	assert.NoError(t, err)
	assert.NotNil(t, dInstance)
	assertSameInstance(t, instance, dInstance)

	instances, err := catalog.List("Calc", nil)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceName, extractErrorCode(err))
	assert.Empty(t, instances)

}

func TestExternalDeregisterInstanceMultipleServiceInstances(t *testing.T) {
	catalog := setupCatalogForTest()

	instance1 := newServiceInstance("Calc", "192.168.0.1", 9080)
	instance2 := instance1.DeepClone()
	instance2.Endpoint.Value = strings.Join([]string{"192.168.0.2", string(9080)}, ":")
	instance2.Endpoint.Type = "tcp"

	id1, _ := doRegister(catalog, instance1)
	doRegister(catalog, instance2)

	dInstance, err := catalog.Deregister(id1)

	assert.NoError(t, err)
	assert.NotNil(t, dInstance)
	assertSameInstance(t, instance1, dInstance)

	instances, err := catalog.List("Calc", nil)

	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assertContainsInstance(t, instances, instance2)

}

func TestExternalDeregisterInstanceNotRegistered(t *testing.T) {
	catalog := setupCatalogForTest()

	id := "service-ID"
	dInstance, err := catalog.Deregister(id)

	assert.Error(t, err)
	assert.Nil(t, dInstance)
	assert.EqualValues(t, ErrorNoSuchServiceInstance, extractErrorCode(err))

}

func TestExternalDeregisterInstanceAlreadyDeregistered(t *testing.T) {
	catalog := setupCatalogForTest()

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)

	id, _ := doRegister(catalog, instance)
	catalog.Deregister(id)

	dInstance, err := catalog.Deregister(id)

	assert.Error(t, err)
	assert.Nil(t, dInstance)
	assert.EqualValues(t, ErrorNoSuchServiceInstance, extractErrorCode(err))
}

func TestExternalDeregisterInstanceAlreadyExpired(t *testing.T) {
	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{testShortTTL, testShortTTL, maximumTTL, 50, "redis", "testaddress", "testpassword", nil}
	catalog := createExternalCatalog(conf, db)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)

	id, _ := doRegister(catalog, instance)
	time.Sleep(testShortTTL * 2)

	dInstance, err := catalog.Deregister(id)

	assert.Error(t, err)
	assert.Nil(t, dInstance)
	assert.EqualValues(t, ErrorNoSuchServiceInstance, extractErrorCode(err))

}

func TestExternalRenewInstance(t *testing.T) {
	catalog := setupCatalogForTest()

	instance := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1:9080", Type: "tcp"},
	}

	id, _ := doRegister(catalog, instance)
	rInstance, err := catalog.Renew(id)

	assert.NoError(t, err)
	assert.NotNil(t, rInstance)
	assertSameInstance(t, instance, rInstance)
}

func TestExternalRenewInstanceNotRegistered(t *testing.T) {
	catalog := setupCatalogForTest()

	_, err := catalog.Renew("some-bogus-id")

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceInstance, extractErrorCode(err))

}

func TestExternalRenewInstanceAlreadyDeregistered(t *testing.T) {
	catalog := setupCatalogForTest()

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)

	id, _ := doRegister(catalog, instance)
	catalog.Deregister(id)
	_, err := catalog.Renew(id)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceInstance, extractErrorCode(err))

}

func TestExternalRenewInstanceAlreadyExpired(t *testing.T) {
	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{testShortTTL, testShortTTL, maximumTTL, 50, "redis", "testaddress", "testpassword", nil}
	catalog := createExternalCatalog(conf, db)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)

	id, _ := doRegister(catalog, instance)
	time.Sleep(testShortTTL * 2)

	_, err := catalog.Renew(id)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceInstance, extractErrorCode(err))

}

func TestExternalFindInstanceByID(t *testing.T) {
	catalog := setupCatalogForTest()

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)
	id, _ := doRegister(catalog, instance)

	actual, _ := catalog.Instance(id)
	actual.LastRenewal = instance.LastRenewal
	actual.RegistrationTime = instance.RegistrationTime
	assert.Equal(t, instance, actual)
	assert.EqualValues(t, instance.ServiceName, actual.ServiceName)
	assert.EqualValues(t, instance.Endpoint.Value, actual.Endpoint.Value)
}

func TestExternalSingleServiceQuota(t *testing.T) {
	var instanceID string
	namespaceCapacity := 10
	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{testShortTTL, testShortTTL, maximumTTL, namespaceCapacity, "redis", "testaddress", "testpassword", nil}
	catalog := createExternalCatalog(conf, db)

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
	dInstance, err := catalog.Deregister(instanceID)
	assert.NoError(t, err)
	assert.NotNil(t, dInstance)
	id, err = doRegister(catalog, instance)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
}

func TestExternalMultipleServicesQuota(t *testing.T) {
	var instanceID string
	namespaceCapacity := 10
	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{testShortTTL, testShortTTL, maximumTTL, namespaceCapacity, "redis", "testaddress", "testpassword", nil}
	catalog := createExternalCatalog(conf, db)

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
	dInstance, err := catalog.Deregister(instanceID)
	assert.NoError(t, err)
	assert.NotNil(t, dInstance)
	id, err = doRegister(catalog, instance)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
}

func TestExternalInstancesMetrics(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()

	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{testShortTTL, testShortTTL, maximumTTL, 10, "redis", "testaddress", "testpassword", nil}
	catalog := createExternalCatalog(conf, db)

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

func TestExternalAverageMetadataMetrics(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()
	catalog := setupCatalogForTest()

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

func TestExternalAverageTagsMetrics(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()
	catalog := setupCatalogForTest()

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

func TestExternalMetadataInstancesMetrics(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()
	catalog := setupCatalogForTest()

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

func TestExternalTagsInstancesMetrics(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()
	catalog := setupCatalogForTest()

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

func TestExternalExpirationMetric(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()

	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{testShortTTL, testShortTTL, maximumTTL, 10, "redis", "testaddress", "testpassword", nil}
	catalog := createExternalCatalog(conf, db)

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

func TestExternalLifetimeMetric(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()

	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{testShortTTL, testShortTTL, maximumTTL, 10, "redis", "testaddress", "testpassword", nil}
	catalog := createExternalCatalog(conf, db)

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

func TestExternalMultiCatalogsMetrics(t *testing.T) {
	metrics.DefaultRegistry.UnregisterAll()

	serviceInstances := make(map[string]*ServiceInstance)
	serviceInstances2 := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)
	db2 := NewMockExternalRegistry(serviceInstances2)

	conf := &externalConfig{testShortTTL, testShortTTL, maximumTTL, 10, "redis", "testaddress", "testpassword", nil}
	catalog1 := createExternalCatalog(conf, db)
	catalog2 := createExternalCatalog(conf, db2)

	instancesCount := func() int64 {
		return metrics.DefaultRegistry.Get(instancesMetricName).(metrics.Counter).Count()
	}

	catalog1.Register(newServiceInstance("Calc", "192.168.0.1", 9080))
	catalog2.Register(newServiceInstance("Calc", "192.168.0.2", 9080))

	assert.EqualValues(t, 2, instancesCount())
}

func TestExternalInstanceExpiresDefaultTTL(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ttl := DefaultConfig.DefaultTTL
	catalog := setupCatalogForTest()

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

func TestExternalInstanceExpiresCatalogTTL(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ttl := time.Duration(5) * time.Second
	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{ttl, testShortTTL, maximumTTL, 10, "redis", "testaddress", "testpassword", nil}
	catalog := createExternalCatalog(conf, db)

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

func TestExternalInstanceExpireInstanceTTL(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ttl := time.Duration(5) * time.Second
	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{DefaultConfig.DefaultTTL, ttl, maximumTTL, 10, "redis", "testaddress", "testpassword", nil}
	catalog := createExternalCatalog(conf, db)

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

func TestExternalRegisterInstancesSameServiceConcurrently(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	const numOfInstances = 10000

	ttl := time.Duration(30) * time.Second
	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{ttl, testShortTTL, maximumTTL, -1, "redis", "testaddress", "testpassword", nil}
	catalog := createExternalCatalog(conf, db)

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

func TestExternalRegisterInstancesConcurrently(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// Test assumes numOfInstances % numOfServices == 0
	const numOfServices = 50
	const numOfInstances = 10000

	ttl := time.Duration(30) * time.Second
	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{ttl, testShortTTL, maximumTTL, -1, "redis", "testaddress", "testpassword", nil}
	catalog := createExternalCatalog(conf, db)

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

func TestExternalDeregisterInstancesSameServiceConcurrently(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	const numOfInstances = 10000

	ttl := time.Duration(30) * time.Second
	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{ttl, testShortTTL, maximumTTL, -1, "redis", "testaddress", "testpassword", nil}
	catalog := createExternalCatalog(conf, db)
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
			_, err := catalog.Deregister(id)
			assert.NoError(t, err, "Error deregistering instance ID %v: %v", id, err)
		}()

	}

	wg.Wait()

	instances, err := catalog.List("Calc", nil)
	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceName, extractErrorCode(err))
	assert.Empty(t, instances)

}

func TestExternalDeregisterInstancesConcurrently(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// Test assumes numOfInstances % numOfServices == 0
	const numOfServices = 50
	const numOfInstances = 10000

	ttl := time.Duration(30) * time.Second
	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{ttl, testShortTTL, maximumTTL, -1, "redis", "testaddress", "testpassword", nil}
	catalog := createExternalCatalog(conf, db)
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
			_, err := catalog.Deregister(id)
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

func TestExternalRenewInstancePreventExpiration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{defaultTTL, testShortTTL, maximumTTL, -1, "redis", "testaddress", "testpassword", nil}
	catalog := createExternalCatalog(conf, db)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)

	id, _ := doRegister(catalog, instance)

	done := make(chan bool)
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(testShortTTL / 2)
			_, err := catalog.Renew(id)
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

func TestExternalRenewInstancesConcurrently(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	if testing.Short() {
		t.SkipNow()
	}

	const numOfInstances = 10000
	const numOfIterations = 10
	const ttl = (20) * time.Second //defaultTTL //testMediumTTL

	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{ttl, testMinTTL, testMaxTTL, -1, "redis", "testaddress", "testpassword", nil}
	catalog := createExternalCatalog(conf, db)

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
						_, err := catalog.Renew(id)
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

func TestExternalRenewAndExpireInstancesConcurrently(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	if testing.Short() {
		t.SkipNow()
	}

	const numOfInstances = 10000
	const numOfIterations = 10
	const expirationProbability = 0.2
	const ttl = (15) * time.Second

	serviceInstances := make(map[string]*ServiceInstance)

	db := NewMockExternalRegistry(serviceInstances)

	conf := &externalConfig{ttl, testMinTTL, testMaxTTL, -1, "redis", "testaddress", "testpassword", nil}
	catalog := createExternalCatalog(conf, db)

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
						_, err := catalog.Renew(id)
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

func createNewExternalConfig(defaultTTL time.Duration) *externalConfig {
	return &externalConfig{defaultTTL, testMinTTL, testMaxTTL, -1, "redis", "testaddress", "testpassword", nil}
}
