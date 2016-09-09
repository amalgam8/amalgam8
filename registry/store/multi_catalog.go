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
	"github.com/amalgam8/amalgam8/pkg/auth"
)

const (
	rwCatalogIndex int = 0
)

type multiConfig struct {
	factories []CatalogFactory
}

type multiFactory struct {
	conf *multiConfig
}

func newMultiFactory(conf *multiConfig) *multiFactory {
	return &multiFactory{conf: conf}
}

func (f *multiFactory) CreateCatalog(namespace auth.Namespace) (Catalog, error) {
	return newMultiCatalog(namespace, f.conf)
}

// MultiCatalog is a collection of catalogs.
// The catalog at index 0 (rwCalogIndex) is the Read-Write catalog. The other catalogs are Read-Only.
type multiCatalog struct {
	catalogs []Catalog
}

func newMultiCatalog(namespace auth.Namespace, conf *multiConfig) (*multiCatalog, error) {
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
	return mc.catalogs[rwCatalogIndex].Register(si)
}

func (mc *multiCatalog) Deregister(instanceID string) (*ServiceInstance, error) {
	return mc.catalogs[rwCatalogIndex].Deregister(instanceID)
}

func (mc *multiCatalog) Renew(instanceID string) (*ServiceInstance, error) {
	return mc.catalogs[rwCatalogIndex].Renew(instanceID)
}

func (mc *multiCatalog) SetStatus(instanceID, status string) (*ServiceInstance, error) {
	return mc.catalogs[rwCatalogIndex].SetStatus(instanceID, status)
}

func (mc *multiCatalog) List(serviceName string, predicate Predicate) ([]*ServiceInstance, error) {
	isErr := true
	instanceCollection := make([]*ServiceInstance, 0, 10)

	for _, catalog := range mc.catalogs {
		list, err := catalog.List(serviceName, predicate)
		// We don't log the error here, because an error ("no such service") is acceptable.
		// We will return an error at the end if and only if the list is empty
		if err == nil {
			isErr = false
			if len(list) > 0 {
				instanceCollection = append(instanceCollection, list...)
			}
		}
	}

	// If and only if all the sub-catalogs returned an error then we have to
	// return an  error
	if isErr {
		return nil, NewError(ErrorNoSuchServiceName, "no such service", serviceName)
	}

	return instanceCollection, nil
}

func (mc *multiCatalog) Instance(instanceID string) (*ServiceInstance, error) {
	for _, catalog := range mc.catalogs {
		si, err := catalog.Instance(instanceID)
		if err == nil {
			return si, nil
		}
	}

	return nil, NewError(ErrorNoSuchServiceInstance, "no such service instance", instanceID)
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
