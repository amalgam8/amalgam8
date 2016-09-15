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

package eureka

import (
	"fmt"
	"sort"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	eurekaapi "github.com/amalgam8/amalgam8/registry/api/protocol/eureka"
	"github.com/amalgam8/amalgam8/registry/store"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

const (
	module            = "EUREKACATALOG"
	refreshInterval   = time.Duration(30) * time.Second
	hashcodeDelimiter = "_"
	actionAdded       = "ADDED"
	actionModified    = "MODIFIED"
	actionDeleted     = "DELETED"
)

type instanceMap map[string]*store.ServiceInstance // instance ID -> instance
type serviceMap map[string]instanceMap             // service name -> instanceMap

type eurekaCatalog struct {
	client *eurekaClient

	services     serviceMap
	instances    instanceMap
	versionDelta int64

	logger *log.Entry
	sync.RWMutex
}

func newEurekaCatalog(client *eurekaClient) (*eurekaCatalog, error) {
	catalog := &eurekaCatalog{
		services:  serviceMap{},
		instances: instanceMap{},
		client:    client,
		logger:    logging.GetLogger(module),
	}

	catalog.refresh()

	ticker := time.NewTicker(refreshInterval)
	go func() {
		for _ = range ticker.C {
			catalog.refresh()
		}
	}()

	return catalog, nil
}

func (ec *eurekaCatalog) ListServices(predicate store.Predicate) []*store.Service {
	ec.RLock()
	defer ec.RUnlock()

	serviceCollection := make([]*store.Service, 0, len(ec.services))
	for service, instances := range ec.services {
		for _, inst := range instances {
			if predicate == nil || predicate(inst) {
				serviceCollection = append(serviceCollection, &store.Service{ServiceName: service})
				break
			}
		}
	}

	return serviceCollection
}

func (ec *eurekaCatalog) List(serviceName string, predicate store.Predicate) ([]*store.ServiceInstance, error) {
	ec.RLock()
	defer ec.RUnlock()

	service := ec.services[serviceName]
	if nil == service {
		return nil, store.NewError(store.ErrorNoSuchServiceName, "no such service", serviceName)
	}

	instanceCollection := make([]*store.ServiceInstance, 0, len(service))
	for _, inst := range service {
		if predicate == nil || predicate(inst) {
			instanceCollection = append(instanceCollection, inst.DeepClone())
		}
	}
	return instanceCollection, nil
}

func (ec *eurekaCatalog) Instance(instanceID string) (*store.ServiceInstance, error) {
	ec.RLock()
	defer ec.RUnlock()

	instance := ec.instances[instanceID]
	if instance == nil {
		return nil, store.NewError(store.ErrorNoSuchServiceInstance, "no such service instance", instanceID)
	}
	return instance.DeepClone(), nil
}

func (ec *eurekaCatalog) Register(si *store.ServiceInstance) (*store.ServiceInstance, error) {
	ec.logger.Infof("Unsupported API (Register) called")
	return nil, store.NewError(store.ErrorBadRequest, "Read-only Catalog: API Not Supported", "Register")
}

func (ec *eurekaCatalog) Deregister(instanceID string) (*store.ServiceInstance, error) {
	ec.logger.Infof("Unsupported API (Deregister) called")
	return nil, store.NewError(store.ErrorBadRequest, "Read-only Catalog: API Not Supported", "Deregister")
}

func (ec *eurekaCatalog) Renew(instanceID string) (*store.ServiceInstance, error) {
	ec.logger.Infof("Unsupported API (Renew) called")
	return nil, store.NewError(store.ErrorBadRequest, "Read-only Catalog: API Not Supported", "Renew")
}

func (ec *eurekaCatalog) SetStatus(instanceID, status string) (*store.ServiceInstance, error) {
	ec.logger.Infof("Unsupported API (SetStatus) called")
	return nil, store.NewError(store.ErrorBadRequest, "Read-only Catalog: API Not Supported", "SetStatus")
}

func (ec *eurekaCatalog) refresh() {
	var services serviceMap
	var instances instanceMap
	var err error

	// If this is the 1st time then we need to retrieve the full registry,
	// otherwise a delta could be sufficient
	if len(ec.services) > 0 {
		services, instances = ec.getServicesDelta()
	}

	if services == nil {
		services, instances, err = ec.getServices()
		ec.versionDelta = 0
	}

	if err == nil {
		ec.Lock()
		defer ec.Unlock()

		ec.services = services
		ec.instances = instances
	}
}

func (ec *eurekaCatalog) getServices() (serviceMap, instanceMap, error) {
	services := serviceMap{}
	instances := instanceMap{}

	apps, err := ec.client.getApplicationsFull()
	if err != nil {
		ec.logger.WithFields(log.Fields{
			"error": err,
		}).Warnf("Faild to retrieve applications")
		return nil, nil, err
	}

	if apps != nil && apps.Application != nil {
		for _, app := range apps.Application {
			sname := app.Name
			svcInstances := instanceMap{}
			for _, inst := range app.Instances {
				si, err := eurekaapi.BuildA8Instance(inst)
				if err != nil {
					ec.logger.WithFields(log.Fields{
						"error": err,
					}).Warnf("Failed to parse instance %+v", inst)
					continue
				}
				if inst.Lease != nil && inst.Lease.RegistrationTs > 0 {
					si.RegistrationTime = time.Unix(inst.Lease.LastRenewalTs/1e3, (inst.Lease.LastRenewalTs%1e3)*1e6)
				}
				svcInstances[si.ID] = si
				instances[si.ID] = si
			}
			services[sname] = svcInstances
		}
	}

	ec.logger.Infof("Update full registry completed successfully (services: %d, instances: %d)", len(services), len(instances))
	return services, instances, nil
}

func (ec *eurekaCatalog) getServicesDelta() (serviceMap, instanceMap) {
	services, instances := ec.copyServices()

	apps, err := ec.client.getApplicationsDelta()
	if err != nil {
		ec.logger.WithFields(log.Fields{
			"error": err,
		}).Warnf("Faild to retrieve applications delta")
		return nil, nil
	}

	if apps == nil {
		// Delta is not supported
		return nil, nil
	}

	// If we have the latest version, no need to do anything
	if apps.VersionDelta == ec.versionDelta {
		ec.logger.Infof("Delta update was skipped, because we have the latest version (%d)", apps.VersionDelta)
		return services, instances
	}

	var updCount, delCount int
	for _, app := range apps.Application {
		for _, inst := range app.Instances {
			si, err := eurekaapi.BuildA8Instance(inst)
			if err != nil {
				ec.logger.WithFields(log.Fields{
					"error": err,
				}).Warnf("Failed to parse instance %+v", inst)
				return nil, nil
			}

			switch inst.ActionType {
			case actionDeleted:
				delete(instances, si.ID)
				if svc, ok := services[si.ServiceName]; ok {
					delete(svc, si.ID)
					if len(svc) == 0 {
						delete(services, si.ServiceName)
					}
				}
				delCount++
			case actionAdded, actionModified:
				if inst.Lease != nil && inst.Lease.RegistrationTs > 0 {
					si.RegistrationTime = time.Unix(inst.Lease.LastRenewalTs/1e3, (inst.Lease.LastRenewalTs%1e3)*1e6)
				}
				instances[si.ID] = si
				insts := services[si.ServiceName]
				if insts == nil {
					insts = instanceMap{}
					services[si.ServiceName] = insts
				}
				insts[si.ID] = si
				updCount++
			default:
				ec.logger.Warnf("Unknown ActionType %s for instance %+v", inst.ActionType, si)

			}
		}
	}

	// Calculate the new hashcode and compare it to the server
	hashcode := calculateHashcode(instances)
	if apps.Hashcode != hashcode {
		ec.logger.Infof("Failed to update delta (local: %s, remote %s). A full update is required", hashcode, apps.Hashcode)
		return nil, nil
	}

	ec.versionDelta = apps.VersionDelta
	ec.logger.Infof("Delta update completed successfully (updated: %d, deleted: %d, version: %d)", updCount, delCount, apps.VersionDelta)

	return services, instances
}

func calculateHashcode(instances instanceMap) string {
	var hashcode string

	if len(instances) == 0 {
		return hashcode
	}

	hashMap := map[string]uint32{}
	for _, si := range instances {
		if count, ok := hashMap[si.Status]; !ok {
			hashMap[si.Status] = 1
		} else {
			hashMap[si.Status] = count + 1
		}
	}

	var keys []string
	for status := range hashMap {
		keys = append(keys, status)
	}
	sort.Strings(keys)

	for _, status := range keys {
		count := hashMap[status]
		hashcode = hashcode + fmt.Sprintf("%s%s%d%s", status, hashcodeDelimiter, count, hashcodeDelimiter)

	}

	return hashcode
}

func (ec *eurekaCatalog) copyServices() (serviceMap, instanceMap) {
	services := serviceMap{}
	instances := instanceMap{}

	ec.Lock()
	defer ec.Unlock()

	for name, insts := range ec.services {
		cpyInsts := instanceMap{}
		for id, inst := range insts {
			cpyInsts[id] = inst
		}
		services[name] = cpyInsts
	}

	for id, inst := range ec.instances {
		instances[id] = inst
	}

	return services, instances
}
