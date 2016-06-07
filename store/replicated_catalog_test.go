package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/amalgam8/registry/auth"
)

func TestRPCNewCatalogNilConfiguration(t *testing.T) {
	catalog := createNewReplicatedCatalog(nil)

	assert.NotNil(t, catalog)

}

func TestRPCEmptyCatalog(t *testing.T) {
	catalog := createNewReplicatedCatalog(nil)

	instances, err := catalog.List("Calc", protocolPredicate)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceName, extractErrorCode(err))
	assert.Empty(t, instances)

}

func TestRPCRegisterInstance(t *testing.T) {
	catalog := createNewReplicatedCatalog(nil)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)
	id, err := doRegister(catalog, instance)

	assert.NoError(t, err)
	assert.NotNil(t, id)
	assert.NotEmpty(t, id)

	instances, err := catalog.List("Calc", protocolPredicate)

	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assertContainsInstance(t, instances, instance)
}

func TestRPCRegisterInstanceWithID(t *testing.T) {
	catalog := createNewReplicatedCatalog(nil)

	instance := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1:" + string(9080), Type: "tcp"},
		ID:          "some-nonsense-id",
	}

	registeredInstance, err := catalog.Register(instance)
	// We allow registration with ID for the replication
	assert.NoError(t, err)
	assert.NotNil(t, registeredInstance)

}

func TestRPCRegisterInstanceWithTTL(t *testing.T) {

	catalog := createNewReplicatedCatalog(nil)

	instance := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1:" + string(9080), Type: "tcp"},
		TTL:         time.Duration(15) * time.Second,
	}

	registeredInstance, err := catalog.Register(instance)

	assert.NoError(t, err)
	assert.NotNil(t, registeredInstance)
	assert.NotNil(t, registeredInstance.ID)
	assert.NotEmpty(t, registeredInstance.ID)
	assert.EqualValues(t, instance.TTL, registeredInstance.TTL)
}

func TestRPCRegisterInstanceSameServiceDifferentEndpoint(t *testing.T) {

	catalog := createNewReplicatedCatalog(nil)

	instance1 := newServiceInstance("Calc", "192.168.0.1", 9080)
	instance2 := instance1.DeepClone()
	instance2.Endpoint.Value = "192.168.0.2"

	id1, _ := doRegister(catalog, instance1)
	id2, err := doRegister(catalog, instance2)

	assert.NoError(t, err)
	assert.NotEqual(t, id1, id2)

	instances, err := catalog.List("Calc", protocolPredicate)

	assert.NoError(t, err)
	assert.Len(t, instances, 2)
	assertContainsInstance(t, instances, instance1)
	assertContainsInstance(t, instances, instance2)
}

func TestRPCRegisterInstanceDifferentServiceSameEndpoint(t *testing.T) {

	catalog := createNewReplicatedCatalog(nil)

	instance1 := newServiceInstance("Calc1", "192.168.0.1", 9080)
	instance2 := instance1.DeepClone()
	instance2.ServiceName = "Calc2"

	id1, _ := doRegister(catalog, instance1)
	id2, err := doRegister(catalog, instance2)

	assert.NoError(t, err)
	assert.NotEqual(t, id1, id2)

	instances1, err := catalog.List("Calc1", protocolPredicate)

	assert.NoError(t, err)
	assert.Len(t, instances1, 1)
	assertContainsInstance(t, instances1, instance1)

	instances2, err := catalog.List("Calc2", protocolPredicate)

	assert.NoError(t, err)
	assert.Len(t, instances2, 1)
	assertContainsInstance(t, instances2, instance2)
}

func TestRPCListServices(t *testing.T) {

	cases := make(map[string]bool)
	cases["Calc1"] = true
	cases["Calc2"] = true
	cases["Calc3"] = true
	cases["Calc4"] = true

	catalog := createNewReplicatedCatalog(nil)

	for key := range cases {
		instance := newServiceInstance(key, "192.168.0.1", 9080)
		doRegister(catalog, instance)
	}

	services := catalog.ListServices(protocolPredicate)
	for _, srv := range services {
		assert.NotNil(t, srv)
		assert.True(t, cases[srv.ServiceName])
	}
}

func TestRPCReRegisterExpiredInstance(t *testing.T) {

	conf := createNewConfig(testShortTTL)
	catalog := createNewReplicatedCatalog(conf)

	instance1 := newServiceInstance("Calc", "192.168.0.1", 9080)

	id1, _ := doRegister(catalog, instance1)
	time.Sleep(testShortTTL * 2)

	instance2 := newServiceInstance("Calc", "192.168.0.1", 9080)
	id2, err := doRegister(catalog, instance2)

	assert.NoError(t, err)
	assert.NotNil(t, id2)
	assert.NotEmpty(t, id2)
	assert.Equal(t, id1, id2)

	instances, err := catalog.List("Calc", protocolPredicate)

	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assertContainsInstance(t, instances, instance2)
}

func TestRPCDeregisterInstance(t *testing.T) {

	catalog := createNewReplicatedCatalog(nil)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)

	id, _ := doRegister(catalog, instance)
	err := catalog.Deregister(id)

	assert.NoError(t, err)

	instances, err := catalog.List("Calc", protocolPredicate)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceName, extractErrorCode(err))
	assert.Empty(t, instances)

}

func TestRPCDeregisterInstanceMultipleServiceInstances(t *testing.T) {

	catalog := createNewReplicatedCatalog(nil)

	instance1 := newServiceInstance("Calc", "192.168.0.1", 9080)
	instance2 := instance1.DeepClone()
	instance2.Endpoint.Value = "192.168.0.2"

	id1, _ := doRegister(catalog, instance1)
	doRegister(catalog, instance2)

	err := catalog.Deregister(id1)

	assert.NoError(t, err)

	instances, err := catalog.List("Calc", protocolPredicate)

	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assertContainsInstance(t, instances, instance2)

}

func TestRPCDeregisterInstanceNotRegistered(t *testing.T) {

	catalog := createNewReplicatedCatalog(nil)

	id := "service-ID"
	err := catalog.Deregister(id)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceInstance, extractErrorCode(err))

}

func TestRPCDeregisterInstanceAlreadyDeregistered(t *testing.T) {

	catalog := createNewReplicatedCatalog(nil)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)

	id, _ := doRegister(catalog, instance)
	catalog.Deregister(id)

	err := catalog.Deregister(id)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceInstance, extractErrorCode(err))

}

func TestRPCDeregisterInstanceAlreadyExpired(t *testing.T) {

	conf := createNewConfig(testShortTTL)
	catalog := createNewReplicatedCatalog(conf)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)

	id, _ := doRegister(catalog, instance)
	time.Sleep(testShortTTL * 2)

	err := catalog.Deregister(id)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceInstance, extractErrorCode(err))

}

func TestRPCRenewInstance(t *testing.T) {

	catalog := createNewReplicatedCatalog(nil)

	instance := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1" + string(9080), Type: "tcp"},
	}

	id, _ := doRegister(catalog, instance)
	err := catalog.Renew(id)

	assert.NoError(t, err)

}

func TestRPCRenewInstanceNotRegistered(t *testing.T) {

	catalog := createNewReplicatedCatalog(nil)

	err := catalog.Renew("some-bogus-id")

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceInstance, extractErrorCode(err))

}

func TestRPCRenewInstanceAlreadyDeregistered(t *testing.T) {

	catalog := createNewReplicatedCatalog(nil)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)

	id, _ := doRegister(catalog, instance)
	catalog.Deregister(id)
	err := catalog.Renew(id)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceInstance, extractErrorCode(err))

}

func TestRPCRenewInstanceAlreadyExpired(t *testing.T) {

	conf := createNewConfig(testShortTTL)
	catalog := createNewReplicatedCatalog(conf)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)

	id, _ := doRegister(catalog, instance)
	time.Sleep(testShortTTL * 2)

	err := catalog.Renew(id)

	assert.Error(t, err)
	assert.EqualValues(t, ErrorNoSuchServiceInstance, extractErrorCode(err))

}

func TestRPCFindInstanceByID(t *testing.T) {
	catalog := createNewReplicatedCatalog(nil)

	instance := newServiceInstance("Calc", "192.168.0.1", 9080)
	id, _ := doRegister(catalog, instance)

	actual, _ := catalog.Instance(id)
	actual.LastRenewal = instance.LastRenewal           // LastRenewal is NOT replicated here
	actual.RegistrationTime = instance.RegistrationTime // RegistrationTime is NOT replicated here
	assert.Equal(t, instance, actual)
	assert.EqualValues(t, instance.ServiceName, actual.ServiceName)
	assert.EqualValues(t, instance.Endpoint.Value, actual.Endpoint.Value)
	assert.EqualValues(t, instance.Metadata, actual.Metadata)
}

func createNewReplicatedCatalog(conf *Config) Catalog {
	rep := createMockupReplication()
	close(rep.(*mockupReplication).syncChan) //Make sure no deadlock occur
	var ns = auth.NamespaceFrom("ns1")
	rpc, _ := rep.GetReplicator(ns)
	catalog := newReplicatedCatalog(ns, conf, rpc)
	return catalog
}
