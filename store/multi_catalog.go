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
	instanceCollection := make([]*ServiceInstance, 0, 10)
	for _, catalog := range mc.catalogs {
		list, err := catalog.List(serviceName, predicate)
		if err == nil {
			if len(list) > 0 {
				instanceCollection = append(instanceCollection, list...)
			}
		}

	}

	if len(instanceCollection) == 0 {
		return nil, NewError(ErrorNoSuchServiceName, "no such service ", serviceName)
	}

	return instanceCollection, nil
}

func (mc *multiCatalog) Instance(instanceID string) (*ServiceInstance, error) {
	return mc.catalogs[0].Instance(instanceID)
}

func (mc *multiCatalog) ListServices(predicate Predicate) []*Service {
	smap := make(map[string]*Service)
	for _, catalog := range mc.catalogs {
		lServices := catalog.ListServices(predicate)
		if len(lServices) > 0 {
			for _, svc := range lServices {
				smap[svc.ServiceName] = svc
			}
		}
	}

	services := make([]*Service, 0, len(smap))
	for _, svc := range smap {
		services = append(services, svc)
	}

	return services
}
