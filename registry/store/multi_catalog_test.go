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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/amalgam8/amalgam8/pkg/auth"
)

type mockCatalog struct {
	nServices  int
	nInstances int
}

func (mc *mockCatalog) CreateCatalog(auth.Namespace) (Catalog, error) {
	return mc, nil
}

func (mc *mockCatalog) Register(si *ServiceInstance) (*ServiceInstance, error) {
	return nil, nil
}

func (mc *mockCatalog) Deregister(instanceID string) (*ServiceInstance, error) {
	return nil, nil
}

func (mc *mockCatalog) Renew(instanceID string) (*ServiceInstance, error) {
	return nil, nil
}

func (mc *mockCatalog) SetStatus(instanceID, status string) (*ServiceInstance, error) {
	return nil, nil
}

func (mc *mockCatalog) Instance(instanceID string) (*ServiceInstance, error) {
	return nil, nil
}

func (mc *mockCatalog) List(serviceName string, predicate Predicate) ([]*ServiceInstance, error) {
	if mc.nInstances == 0 {
		return nil, NewError(ErrorNoSuchServiceName, "no such service ", serviceName)
	}

	list := make([]*ServiceInstance, mc.nInstances)
	for i := 0; i < mc.nInstances; i++ {
		list[i] = newServiceInstance("Calc", fmt.Sprintf("192.168.0.%d", i), 9080)
	}

	return list, nil
}

func (mc *mockCatalog) ListServices(predicate Predicate) []*Service {
	if mc.nServices == 0 {
		return nil
	}

	list := make([]*Service, mc.nServices)
	for i := 0; i < mc.nServices; i++ {
		list[i] = &Service{fmt.Sprintf("Service-%d", i)}
	}

	return list
}

func TestNewMultiCatalog(t *testing.T) {

	inmemF := newInMemoryFactory(nil)
	conf := &multiConfig{[]CatalogFactory{inmemF, inmemF}}
	factory := newMultiFactory(conf)
	catalog, err := factory.CreateCatalog(auth.NamespaceFrom("ns1"))

	assert.NoError(t, err)
	assert.NotNil(t, catalog)
}

func TestMultiCatalogListServicesNoContent(t *testing.T) {

	inmemF := newInMemoryFactory(nil)
	nServices := 0
	conf := &multiConfig{[]CatalogFactory{inmemF, &mockCatalog{nServices, 0}}}
	factory := newMultiFactory(conf)
	catalog, err := factory.CreateCatalog(auth.NamespaceFrom("ns1"))

	assert.NoError(t, err)
	assert.NotNil(t, catalog)

	list := catalog.ListServices(nil)
	assert.NotNil(t, list)
	assert.EqualValues(t, nServices, len(list))
}

func TestMultiCatalogListServicesWithContent(t *testing.T) {

	inmemF := newInMemoryFactory(nil)
	nServices := 5
	conf := &multiConfig{[]CatalogFactory{inmemF, &mockCatalog{nServices, 0}}}
	factory := newMultiFactory(conf)
	catalog, err := factory.CreateCatalog(auth.NamespaceFrom("ns1"))

	assert.NoError(t, err)
	assert.NotNil(t, catalog)

	list := catalog.ListServices(nil)
	assert.NotNil(t, list)
	assert.EqualValues(t, nServices, len(list))
}

func TestMultiCatalogListNoContent(t *testing.T) {

	inmemF := newInMemoryFactory(nil)
	nInstances := 0
	conf := &multiConfig{[]CatalogFactory{inmemF, &mockCatalog{0, nInstances}}}
	factory := newMultiFactory(conf)
	catalog, err := factory.CreateCatalog(auth.NamespaceFrom("ns1"))

	assert.NoError(t, err)
	assert.NotNil(t, catalog)

	list, err := catalog.List("Calc", nil)
	assert.Error(t, err)
	assert.Nil(t, list)
}

func TestMultiCatalogListWithContent(t *testing.T) {

	inmemF := newInMemoryFactory(nil)
	nInstances := 5
	conf := &multiConfig{[]CatalogFactory{inmemF, &mockCatalog{0, nInstances}}}
	factory := newMultiFactory(conf)
	catalog, err := factory.CreateCatalog(auth.NamespaceFrom("ns1"))

	assert.NoError(t, err)
	assert.NotNil(t, catalog)

	list, err := catalog.List("Calc", nil)
	assert.NoError(t, err)
	assert.NotNil(t, list)
	assert.EqualValues(t, nInstances, len(list))
}

func TestMultiCatalogRegister(t *testing.T) {

	inmemF := newInMemoryFactory(nil)
	nServices := 3
	nInstances := 5
	conf := &multiConfig{[]CatalogFactory{inmemF, &mockCatalog{nServices, nInstances}}}
	factory := newMultiFactory(conf)
	catalog, err := factory.CreateCatalog(auth.NamespaceFrom("ns1"))

	assert.NoError(t, err)
	assert.NotNil(t, catalog)

	si1 := newServiceInstance("Service-0", "192.168.1.1", 9080)
	_, err = catalog.Register(si1)
	assert.NoError(t, err)

	si2 := newServiceInstance("Service-New", "192.168.1.2", 9080)
	_, err = catalog.Register(si2)
	assert.NoError(t, err)

	l1, err := catalog.List("Service-New", nil)
	assert.NoError(t, err)
	assert.NotNil(t, l1)
	assert.EqualValues(t, nInstances+1, len(l1))

	l2 := catalog.ListServices(nil)
	assert.NoError(t, err)
	assert.NotNil(t, l2)
	assert.EqualValues(t, nServices+1, len(l2))

}
