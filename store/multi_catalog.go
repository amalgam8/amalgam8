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
	"github.com/amalgam8/registry/auth"
)

type multiConfig struct {
	factories []CatalogFactory
}

type multiFactory struct {
	conf *multiConfig
}

func newMultiFactory(conf *multiConfig) CatalogFactory {
	return &multiFactory{conf: conf}
}

func (f *multiFactory) CreateCatalog(namespace auth.Namespace) (Catalog, error) {
	return newMultiCatalog(namespace, f.conf)
}

type multiCatalog struct {
	catalogs []Catalog
}

func newMultiCatalog(namespace auth.Namespace, conf *multiConfig) (Catalog, error) {
	if len(conf.factories) == 1 {
		return conf.factories[0].CreateCatalog(namespace)
	}

	catalogs := make([]Catalog, len(conf.factories))
	for i, factory := range conf.factories {
		catalog, err := factory.CreateCatalog(namespace)
		if err != nil {
			return nil, err
		}
		catalogs[i] = catalog
	}

	return &multiCatalog{catalogs: catalogs}, nil
}

func (mc *multiCatalog) Register(si *ServiceInstance) (*ServiceInstance, error) {
	return mc.catalogs[0].Register(si)
}

func (mc *multiCatalog) Deregister(instanceID string) error {
	return mc.catalogs[0].Deregister(instanceID)
}

func (mc *multiCatalog) Renew(instanceID string) error {
	return mc.catalogs[0].Renew(instanceID)
}

func (mc *multiCatalog) SetStatus(instanceID, status string) error {
	return mc.catalogs[0].SetStatus(instanceID, status)
}

func (mc *multiCatalog) List(serviceName string, predicate Predicate) ([]*ServiceInstance, error) {
	return mc.catalogs[0].List(serviceName, predicate)
}

func (mc *multiCatalog) Instance(instanceID string) (*ServiceInstance, error) {
	return mc.catalogs[0].Instance(instanceID)
}

func (mc *multiCatalog) ListServices(predicate Predicate) []*Service {
	return mc.catalogs[0].ListServices(predicate)
}
