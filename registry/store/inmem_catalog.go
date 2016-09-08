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
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/rcrowley/go-metrics"

	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

var defaultInMemoryConfig = &inMemoryConfig{DefaultConfig.DefaultTTL, DefaultConfig.MinimumTTL, DefaultConfig.MaximumTTL, DefaultConfig.NamespaceCapacity}

type inMemoryConfig struct {
	defaultTTL time.Duration
	minimumTTL time.Duration
	maximumTTL time.Duration

	namespaceCapacity int
}

type inMemoryFactory struct {
	conf *inMemoryConfig
}

func newInMemoryFactory(conf *inMemoryConfig) *inMemoryFactory {
	return &inMemoryFactory{conf: conf}
}

func (f *inMemoryFactory) CreateCatalog(namespace auth.Namespace) (Catalog, error) {
	return newInMemoryCatalog(f.conf), nil
}

type inMemoryService map[string]*ServiceInstance

type inMemoryCatalog struct {
	services  map[string]inMemoryService
	instances map[string]*ServiceInstance
	conf      *inMemoryConfig
	logger    *log.Entry

	// Metrics
	instancesMetric         metrics.Counter
	expirationMetric        metrics.Meter
	lifetimeMetric          metrics.Histogram
	metadataLengthMetric    metrics.Histogram
	metadataInstancesMetric metrics.Counter
	tagsLengthMetric        metrics.Histogram
	tagsInstancesMetric     metrics.Counter

	sync.RWMutex
}

func newInMemoryCatalog(conf *inMemoryConfig) *inMemoryCatalog {
	if conf == nil {
		conf = defaultInMemoryConfig
	}

	counterFactory := func() metrics.Counter { return metrics.NewCounter() }
	meterFactory := func() metrics.Meter { return metrics.NewMeter() }
	histogramFactory := func() metrics.Histogram { return metrics.NewHistogram(metrics.NewExpDecaySample(1028, 0.015)) }

	catalog := &inMemoryCatalog{
		services:  make(map[string]inMemoryService),
		instances: make(map[string]*ServiceInstance),
		conf:      conf,
		logger:    logging.GetLogger(module),

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

func (imc *inMemoryCatalog) Register(si *ServiceInstance) (*ServiceInstance, error) {
	serviceName := si.ServiceName
	if serviceName == "" {
		return nil, NewError(ErrorBadRequest, "Empty service name", "")
	}

	if len(serviceName) > serviceNameMaxLength {
		return nil, NewError(ErrorBadRequest, "Service name length too long", "")
	}

	if si.Endpoint != nil && len(si.Endpoint.Value) > valueMaxLength {
		return nil, NewError(ErrorBadRequest, "Endpoint value length too long", "")
	}

	if len(si.Status) > statusMaxLength {
		return nil, NewError(ErrorBadRequest, "Status length too long", "")
	}

	if si.Metadata != nil && len(si.Metadata) > metadataMaxLength {
		return nil, NewError(ErrorBadRequest, "Metadata length too long", "")
	}

	instanceID := si.ID

	if instanceID == "" {
		instanceID = computeInstanceID(si)
	}

	newSI := si.DeepClone()
	newSI.ID = instanceID
	if newSI.TTL == 0 {
		newSI.TTL = imc.conf.defaultTTL
	} else if newSI.TTL < imc.conf.minimumTTL {
		newSI.TTL = imc.conf.minimumTTL
	} else if newSI.TTL > imc.conf.maximumTTL {
		newSI.TTL = imc.conf.maximumTTL
	}

	// isReplication indicates whether this is a replication request or a client request
	isReplication := true
	if newSI.RegistrationTime.IsZero() {
		newSI.RegistrationTime = time.Now()
		isReplication = false
	}

	imc.Lock()
	defer imc.Unlock()

	// Existing instances are simply overwritten, but need to take into account for capacity validation and metrics collection.
	existingSI, alreadyExists := imc.instances[instanceID]
	if alreadyExists {
		imc.logger.Debugf("Overwriting existing instance ID %s due to re-registration", instanceID)
	}

	// Capacity validation - we don't check capacity for replication requests nor reregister requests
	if !isReplication && !alreadyExists && imc.conf.namespaceCapacity >= 0 {
		if len(imc.instances) >= imc.conf.namespaceCapacity {
			imc.logger.Warnf("Failed to register service instance %s becuase quota exceeded (%d)", serviceName, len(imc.instances))
			return nil, NewError(ErrorNamespaceQuotaExceeded, "Quota exceeded", "")
		}

	}

	service, exists := imc.services[serviceName]
	if !exists {
		service = make(map[string]*ServiceInstance)
		imc.services[serviceName] = service
	}

	service[instanceID] = newSI
	imc.instances[instanceID] = newSI

	imc.renew(newSI)

	metadataLength := len(newSI.Metadata)
	tagsLength := len(newSI.Tags)

	// Update the instances/metadata/tags counter metrics
	if !alreadyExists {
		// For a newly registered instance, simply inc the instances counter,
		// and if metadata/tags are used, inc the metadata/tags counter respectively
		imc.instancesMetric.Inc(1)
		if metadataLength > 0 {
			imc.metadataInstancesMetric.Inc(1)
		}
		if tagsLength > 0 {
			imc.tagsInstancesMetric.Inc(1)
		}
	} else {
		// For overwriting an existing instance, no need to inc the instances counter,\
		// but the metadata/tags counter are inc'ed/dec'ed as needed
		prevMetadataLength := len(existingSI.Metadata)
		prevTagsLength := len(existingSI.Tags)

		if prevMetadataLength > 0 && metadataLength == 0 {
			imc.metadataInstancesMetric.Dec(1)
		} else if prevMetadataLength == 0 && metadataLength > 0 {
			imc.metadataInstancesMetric.Inc(1)
		}

		if prevTagsLength > 0 && tagsLength == 0 {
			imc.tagsInstancesMetric.Dec(1)
		} else if prevTagsLength == 0 && tagsLength > 0 {
			imc.tagsInstancesMetric.Inc(1)
		}
	}

	// Update the metadata/tags histogram metrics
	if metadataLength > 0 {
		imc.metadataLengthMetric.Update(int64(metadataLength))
	}
	if tagsLength > 0 {
		imc.tagsLengthMetric.Update(int64(tagsLength))
	}

	return newSI.DeepClone(), nil
}

func (imc *inMemoryCatalog) Deregister(instanceID string) (*ServiceInstance, error) {
	imc.Lock()
	defer imc.Unlock()

	instance := imc.delete(instanceID)
	if instance == nil {
		return nil, NewError(ErrorNoSuchServiceInstance, "no such service instance", instanceID)
	}

	return instance, nil
}

func (imc *inMemoryCatalog) Renew(instanceID string) (*ServiceInstance, error) {
	imc.RLock()
	defer imc.RUnlock()

	instance, exists := imc.instances[instanceID]
	if !exists {
		return nil, NewError(ErrorNoSuchServiceInstance, "no such service instance", instanceID)
	}

	imc.renew(instance)
	return instance.DeepClone(), nil
}

func (imc *inMemoryCatalog) SetStatus(instanceID, status string) (*ServiceInstance, error) {
	imc.Lock()
	defer imc.Unlock()

	instance, exists := imc.instances[instanceID]
	if !exists {
		return nil, NewError(ErrorNoSuchServiceInstance, "no such service instance", instanceID)
	}

	instance.Status = status
	imc.renew(instance)
	return instance.DeepClone(), nil
}

func (imc *inMemoryCatalog) List(serviceName string, predicate Predicate) ([]*ServiceInstance, error) {
	imc.RLock()
	defer imc.RUnlock()

	service := imc.services[serviceName]

	if nil == service {
		return nil, NewError(ErrorNoSuchServiceName, "no such service", serviceName)
	}

	instanceCollection := make([]*ServiceInstance, 0, len(service))
	for _, instance := range service {
		if predicate == nil || predicate(instance) {
			instanceCollection = append(instanceCollection, instance.DeepClone())
		}
	}

	return instanceCollection, nil
}

func (imc *inMemoryCatalog) Instance(instanceID string) (*ServiceInstance, error) {
	imc.RLock()
	defer imc.RUnlock()

	instance, exists := imc.instances[instanceID]
	if !exists {
		return nil, NewError(ErrorNoSuchServiceInstance, "no such service instance", instanceID)
	}

	return instance.DeepClone(), nil
}

func (imc *inMemoryCatalog) ListServices(predicate Predicate) []*Service {
	imc.RLock()
	defer imc.RUnlock()

	services := make([]*Service, 0, len(imc.services))
	for service, instances := range imc.services {
		for _, instance := range instances {
			if predicate == nil || predicate(instance) {
				services = append(services, &Service{ServiceName: service})
				break
			}
		}
	}

	return services
}

func (imc *inMemoryCatalog) checkIfExpired(instanceID string) {
	var instance *ServiceInstance
	func() {
		// Ugly trick to make sure RUnlock() is called in case map lookup fails.
		imc.RLock()
		defer imc.RUnlock()
		instance = imc.instances[instanceID]
	}()

	if instance == nil {
		return
	}

	// If the status is OUT_OF_SERVICE do not expire
	if instance.Status == OutOfService {
		return
	}

	timeSinceHeartbeat := time.Now().Sub(instance.LastRenewal)
	if timeSinceHeartbeat > instance.TTL {

		// Since timeSinceHeartbeat was calculated based
		// on a possibly stale value of inst.LastRenewal,
		// we sync our goroutine and then recalculate it.
		// This should hopefully sync other goroutines running on the same CPU.
		imc.Lock()
		defer imc.Unlock()

		timeSinceHeartbeat = time.Now().Sub(instance.LastRenewal)
		if timeSinceHeartbeat <= instance.TTL {
			return
		}

		imc.logger.Debugf("Instance ID %s is expired", instance.ID)
		imc.delete(instanceID)
		imc.expirationMetric.Mark(1)
	}
}

// delete deletes the specified instanceID from the catalog internal datastructures.
// It assumes the catalog's write-lock is acquired by the calling goroutine.
func (imc *inMemoryCatalog) delete(instanceID string) *ServiceInstance {
	instance, exists := imc.instances[instanceID]
	if !exists {
		return nil
	}
	serviceName := instance.ServiceName

	delete(imc.services[serviceName], instanceID)
	if len(imc.services[serviceName]) == 0 {
		delete(imc.services, serviceName)
	}

	delete(imc.instances, instanceID)

	lifetime := time.Now().Sub(instance.RegistrationTime)
	imc.lifetimeMetric.Update(int64(lifetime))

	hadMetadata := len(instance.Metadata) > 0
	hadTags := len(instance.Tags) > 0

	if hadMetadata {
		imc.metadataInstancesMetric.Dec(1)
	}
	if hadTags {
		imc.tagsInstancesMetric.Dec(1)
	}

	imc.instancesMetric.Dec(1)
	return instance
}

func (imc *inMemoryCatalog) renew(instance *ServiceInstance) {
	instance.LastRenewal = time.Now()

	time.AfterFunc(instance.TTL, func() {
		imc.checkIfExpired(instance.ID)
	})
}
