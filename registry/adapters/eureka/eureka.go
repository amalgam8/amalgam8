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

	"github.com/amalgam8/amalgam8/pkg/api"
	eurekaapi "github.com/amalgam8/amalgam8/registry/server/protocol/eureka"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

const (
	module            = "EUREKAADAPTER"
	refreshInterval   = time.Duration(30) * time.Second
	hashcodeDelimiter = "_"
	actionAdded       = "ADDED"
	actionModified    = "MODIFIED"
	actionDeleted     = "DELETED"
)

// Make sure we implement the ServiceDiscovery interface
var _ api.ServiceDiscovery = (*Adapter)(nil)

type instanceMap map[string]*api.ServiceInstance // instance ID -> instance
type serviceMap map[string]instanceMap           // service name -> instanceMap

// Config encapsulates Eureka configuration parameters
type Config struct {
	URLs []string
}

// Adapter for Eureka Service Discovery
type Adapter struct {
	sync.RWMutex

	client *client

	services  serviceMap
	instances instanceMap

	versionDelta int64

	logger *log.Entry
}

// New creates and initializes a new Eureka Service Discovery adapter
func New(config Config) (*Adapter, error) {
	client, err := newClient(config.URLs)
	if err != nil {
		return nil, err
	}

	adapter := &Adapter{
		services:  serviceMap{},
		instances: instanceMap{},
		client:    client,
		logger:    logging.GetLogger(module),
	}

	adapter.refresh()

	ticker := time.NewTicker(refreshInterval)
	go func() {
		for _ = range ticker.C {
			adapter.refresh()
		}
	}()

	return adapter, nil
}

// ListServices queries for the list of services for which instances are currently registered.
func (a *Adapter) ListServices() ([]string, error) {
	a.RLock()
	defer a.RUnlock()

	services := make([]string, 0, len(a.services))
	for service := range a.services {
		services = append(services, service)
	}

	return services, nil
}

// ListInstances queries for the list of service instances currently registered.
func (a *Adapter) ListInstances() ([]*api.ServiceInstance, error) {
	a.RLock()
	defer a.RUnlock()

	instances := make([]*api.ServiceInstance, 0, len(a.services)*3)
	for _, service := range a.services {
		for _, instance := range service {
			instances = append(instances, instance)
		}
	}

	return instances, nil
}

// ListServiceInstances queries for the list of service instances currently registered for the given service.
func (a *Adapter) ListServiceInstances(serviceName string) ([]*api.ServiceInstance, error) {
	a.RLock()
	defer a.RUnlock()

	service := a.services[serviceName]
	instances := make([]*api.ServiceInstance, 0, len(service))
	for _, instance := range service {
		instances = append(instances, instance)
	}

	return instances, nil
}

func (a *Adapter) refresh() {
	var services serviceMap
	var instances instanceMap
	var err error

	// If this is the 1st time then we need to retrieve the full registry,
	// otherwise a delta could be sufficient
	if len(a.services) > 0 {
		services, instances = a.getServicesDelta()
	}

	if services == nil {
		services, instances, err = a.getServices()
		a.versionDelta = 0
	}

	if err == nil {
		a.Lock()
		defer a.Unlock()

		a.services = services
		a.instances = instances
	}
}

func (a *Adapter) getServices() (serviceMap, instanceMap, error) {
	services := serviceMap{}
	instances := instanceMap{}

	apps, err := a.client.getApplicationsFull()
	if err != nil {
		a.logger.WithFields(log.Fields{
			"error": err,
		}).Warnf("Failed to retrieve applications")
		return nil, nil, err
	}

	if apps != nil && apps.Application != nil {
		for _, app := range apps.Application {
			sname := app.Name
			svcInstances := instanceMap{}
			for _, inst := range app.Instances {
				si, err := translateInstance(inst)
				if err != nil {
					a.logger.WithFields(log.Fields{
						"error": err,
					}).Warnf("Failed to parse instance %+v", inst)
					continue
				}

				svcInstances[si.ID] = si
				instances[si.ID] = si
			}
			services[sname] = svcInstances
		}
	}

	a.logger.Infof("Update full registry completed successfully (services: %d, instances: %d)", len(services), len(instances))
	return services, instances, nil
}

func (a *Adapter) getServicesDelta() (serviceMap, instanceMap) {
	services, instances := a.copyServices()

	apps, err := a.client.getApplicationsDelta()
	if err != nil {
		a.logger.WithFields(log.Fields{
			"error": err,
		}).Warnf("Faild to retrieve applications delta")
		return nil, nil
	}

	if apps == nil {
		// Delta is not supported
		return nil, nil
	}

	// If we have the latest version, no need to do anything
	if apps.VersionDelta == a.versionDelta {
		a.logger.Infof("Delta update was skipped, because we have the latest version (%d)", apps.VersionDelta)
		return services, instances
	}

	var updated, deleted int
	for _, app := range apps.Application {
		for _, inst := range app.Instances {
			si, err := translateInstance(inst)
			if err != nil {
				a.logger.WithFields(log.Fields{
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
				deleted++
			case actionAdded, actionModified:
				instances[si.ID] = si
				insts := services[si.ServiceName]
				if insts == nil {
					insts = instanceMap{}
					services[si.ServiceName] = insts
				}
				insts[si.ID] = si
				updated++
			default:
				a.logger.Warnf("Unknown ActionType %s for instance %+v", inst.ActionType, si)
			}
		}
	}

	// Calculate the new hashcode and compare it to the server
	hashcode := calculateHashcode(instances)
	if apps.Hashcode != hashcode {
		a.logger.Infof("Failed to update delta (local: %s, remote %s). A full update is required", hashcode, apps.Hashcode)
		return nil, nil
	}

	a.versionDelta = apps.VersionDelta
	a.logger.Infof("Delta update completed successfully (updated: %d, deleted: %d, version: %d)", updated, deleted, apps.VersionDelta)

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

func (a *Adapter) copyServices() (serviceMap, instanceMap) {
	services := serviceMap{}
	instances := instanceMap{}

	a.Lock()
	defer a.Unlock()

	for name, insts := range a.services {
		cpyInsts := instanceMap{}
		for id, inst := range insts {
			cpyInsts[id] = inst
		}
		services[name] = cpyInsts
	}

	for id, inst := range a.instances {
		instances[id] = inst
	}

	return services, instances
}

func translateInstance(eurekaInstance *eurekaapi.Instance) (*api.ServiceInstance, error) {
	storeInstance, err := eurekaapi.Translate(eurekaInstance)
	if err != nil {
		return nil, err
	}

	apiInstance := &api.ServiceInstance{
		ID:          storeInstance.ID,
		ServiceName: storeInstance.ServiceName,
		Endpoint: api.ServiceEndpoint{
			Type:  storeInstance.Endpoint.Type,
			Value: storeInstance.Endpoint.Value,
		},
		Status:        storeInstance.Status,
		Metadata:      storeInstance.Metadata,
		LastHeartbeat: storeInstance.LastRenewal,
		TTL:           int(storeInstance.TTL / time.Second),
		Tags:          storeInstance.Tags,
	}

	return apiInstance, nil
}
