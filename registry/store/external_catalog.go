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
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/garyburd/redigo/redis"
	"github.com/rcrowley/go-metrics"

	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/utils/database"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

type externalConfig struct {
	defaultTTL time.Duration
	minimumTTL time.Duration
	maximumTTL time.Duration

	namespaceCapacity int

	store    string
	address  string
	password string
	database database.Database
}

type externalFactory struct {
	conf *externalConfig
	pool *redis.Pool
}

func newExternalFactory(conf *externalConfig) CatalogFactory {
	if conf.store == "redis" {
		pool := &redis.Pool{
			MaxIdle:     10,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", conf.address)
				if err != nil {
					return nil, err
				}
				if conf.password != "" {
					if _, err := c.Do("AUTH", conf.password); err != nil {
						c.Close()
						return nil, err
					}
				}
				return c, nil
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if time.Since(t) < time.Minute {
					return nil
				}
				_, err := c.Do("PING")
				return err
			},
		}
		return &externalFactory{conf: conf, pool: pool}
	}
	return &externalFactory{conf: conf}
}

func (f *externalFactory) CreateCatalog(namespace auth.Namespace) (Catalog, error) {
	return newExternalCatalog(f.conf, namespace, f.pool)
}

type externalCatalog struct {
	conf      *externalConfig
	logger    *log.Entry
	db        ExternalRegistry
	namespace auth.Namespace

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

func newExternalCatalog(conf *externalConfig, namespace auth.Namespace, pool *redis.Pool) (Catalog, error) {
	if conf == nil {
		// If conf is null, we'll error out when checking the store.  Just error here in this case.
		return nil, fmt.Errorf("Config cannot be nil")
	}

	counterFactory := func() metrics.Counter { return metrics.NewCounter() }
	meterFactory := func() metrics.Meter { return metrics.NewMeter() }
	histogramFactory := func() metrics.Histogram { return metrics.NewHistogram(metrics.NewExpDecaySample(1028, 0.015)) }

	var db database.Database
	var reg ExternalRegistry

	if conf.store == "redis" {
		if conf.database == nil {
			if pool != nil {
				db = database.NewRedisDBWithPool(pool)
			} else {
				db = database.NewRedisDB(conf.address, conf.password)
			}
			reg = NewRedisRegistry(db)
		} else {
			db = conf.database
			reg = NewRedisRegistry(conf.database)
		}
	} else {
		return nil, fmt.Errorf("External store %s is not supported", conf.store)
	}

	catalog := &externalCatalog{
		conf:      conf,
		logger:    logging.GetLogger(module),
		namespace: namespace,
		db:        reg,

		instancesMetric:         metrics.GetOrRegister(instancesMetricName, counterFactory).(metrics.Counter),
		expirationMetric:        metrics.GetOrRegister(expirationMetricName, meterFactory).(metrics.Meter),
		lifetimeMetric:          metrics.GetOrRegister(lifetimeMetricName, histogramFactory).(metrics.Histogram),
		metadataLengthMetric:    metrics.GetOrRegister(metadataLengthMetricName, histogramFactory).(metrics.Histogram),
		metadataInstancesMetric: metrics.GetOrRegister(metadataInstancesMetricName, counterFactory).(metrics.Counter),
		tagsLengthMetric:        metrics.GetOrRegister(tagsLengthMetricName, histogramFactory).(metrics.Histogram),
		tagsInstancesMetric:     metrics.GetOrRegister(tagsInstancesMetricName, counterFactory).(metrics.Counter),
	}

	// Need to check if any entries in the DB have expired
	go func(namespace auth.Namespace, db database.Database, catalog *externalCatalog) {
		hashKeys, _ := db.ReadKeys(namespace.String())
		for _, value := range hashKeys {
			catalog.checkIfExpired(strings.Split(string(value), ".")[0])
		}
	}(namespace, db, catalog)

	return catalog, nil
}

func (ec *externalCatalog) Register(si *ServiceInstance) (*ServiceInstance, error) {
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
		newSI.TTL = ec.conf.defaultTTL
	} else if newSI.TTL < ec.conf.minimumTTL {
		newSI.TTL = ec.conf.minimumTTL
	} else if newSI.TTL > ec.conf.maximumTTL {
		newSI.TTL = ec.conf.maximumTTL
	}

	if newSI.RegistrationTime.IsZero() {
		newSI.RegistrationTime = time.Now()
		newSI.LastRenewal = newSI.RegistrationTime
	}

	ec.Lock()
	defer ec.Unlock()

	// Existing instances are simply overwritten, but need to take into account for capacity validation and metrics collection.
	instance, err := ec.db.ReadServiceInstanceByInstID(ec.namespace, instanceID)
	if err != nil {
		return nil, err
	}
	var alreadyExists bool
	if instance.ID != "" {
		alreadyExists = true
		ec.logger.Debugf("Overwriting existing instance ID %s due to re-registration", instanceID)
	}

	// Capacity validation - we don't check capacity for replication requests nor reregister requests
	if !alreadyExists && ec.conf.namespaceCapacity >= 0 {
		hashKeys, err := ec.db.ReadKeys(ec.namespace)
		if err != nil {
			return nil, err
		}
		if len(hashKeys) >= ec.conf.namespaceCapacity {
			ec.logger.Warnf("Failed to register service instance %s because quota exceeded (%d)", serviceName, len(hashKeys))
			return nil, NewError(ErrorNamespaceQuotaExceeded, "Quota exceeded", "")
		}

	}

	// Write the JSON registration data to the database
	err = ec.db.InsertServiceInstance(ec.namespace, newSI)
	if err != nil {
		return nil, err
	}

	ec.renew(newSI)

	metadataLength := len(newSI.Metadata)
	tagsLength := len(newSI.Tags)

	// Update the instances/metadata/tags counter metrics
	if !alreadyExists {
		// For a newly registered instance, simply inc the instances counter,
		// and if metadata/tags are used, inc the metadata/tags counter respectively
		ec.instancesMetric.Inc(1)
		if metadataLength > 0 {
			ec.metadataInstancesMetric.Inc(1)
		}
		if tagsLength > 0 {
			ec.tagsInstancesMetric.Inc(1)
		}
	} else {
		// For overwriting an existing instance, no need to inc the instances counter,\
		// but the metadata/tags counter are inc'ed/dec'ed as needed
		prevMetadataLength := len(instance.Metadata)
		prevTagsLength := len(instance.Tags)

		if prevMetadataLength > 0 && metadataLength == 0 {
			ec.metadataInstancesMetric.Dec(1)
		} else if prevMetadataLength == 0 && metadataLength > 0 {
			ec.metadataInstancesMetric.Inc(1)
		}

		if prevTagsLength > 0 && tagsLength == 0 {
			ec.tagsInstancesMetric.Dec(1)
		} else if prevTagsLength == 0 && tagsLength > 0 {
			ec.tagsInstancesMetric.Inc(1)
		}
	}

	// Update the metadata/tags histogram metrics
	if metadataLength > 0 {
		ec.metadataLengthMetric.Update(int64(metadataLength))
	}
	if tagsLength > 0 {
		ec.tagsLengthMetric.Update(int64(tagsLength))
	}

	return newSI.DeepClone(), nil
}

func (ec *externalCatalog) Deregister(instanceID string) (*ServiceInstance, error) {
	ec.Lock()
	defer ec.Unlock()

	instance := ec.delete(instanceID)
	if instance == nil {
		return nil, NewError(ErrorNoSuchServiceInstance, "no such service instance", instanceID)
	}

	return instance, nil
}

func (ec *externalCatalog) Renew(instanceID string) (*ServiceInstance, error) {
	ec.Lock()
	defer ec.Unlock()

	si, err := ec.db.ReadServiceInstanceByInstID(ec.namespace, instanceID)
	if err != nil {
		return nil, err
	}
	if si.ID == "" {
		return nil, NewError(ErrorNoSuchServiceInstance, "no such service instance", instanceID)
	}

	si.LastRenewal = time.Now()
	err = ec.db.InsertServiceInstance(ec.namespace, si)
	if err != nil {
		return nil, err
	}

	ec.renew(si)

	return si.DeepClone(), nil
}

func (ec *externalCatalog) SetStatus(instanceID, status string) (*ServiceInstance, error) {
	ec.Lock()
	defer ec.Unlock()

	si, err := ec.db.ReadServiceInstanceByInstID(ec.namespace, instanceID)
	if err != nil {
		return nil, err
	}
	if si.ID == "" {
		return nil, NewError(ErrorNoSuchServiceInstance, "no such service instance", instanceID)
	}

	si.Status = status
	err = ec.db.InsertServiceInstance(ec.namespace, si)
	if err != nil {
		return nil, err
	}

	ec.renew(si)

	return si.DeepClone(), nil
}

func (ec *externalCatalog) List(serviceName string, predicate Predicate) ([]*ServiceInstance, error) {
	ec.RLock()
	defer ec.RUnlock()

	siKey := fmt.Sprintf("*.%s", serviceName)

	service, err := ec.db.ListServiceInstancesByKey(ec.namespace, siKey)
	if err != nil {
		return nil, err
	}
	if len(service) == 0 {
		return nil, NewError(ErrorNoSuchServiceName, "no such service", serviceName)
	}

	instanceCollection := make([]*ServiceInstance, 0, len(service))
	for _, value := range service {
		if predicate == nil || predicate(value) {
			instanceCollection = append(instanceCollection, value)
		}
	}

	return instanceCollection, nil
}

func (ec *externalCatalog) Instance(instanceID string) (*ServiceInstance, error) {
	ec.RLock()
	defer ec.RUnlock()

	si, err := ec.db.ReadServiceInstanceByInstID(ec.namespace, instanceID)
	if err != nil {
		return nil, err
	}
	if si.ID == "" {
		return nil, NewError(ErrorNoSuchServiceInstance, "no such service instance", instanceID)
	}

	return si.DeepClone(), nil
}

func (ec *externalCatalog) ListServices(predicate Predicate) []*Service {
	ec.RLock()
	defer ec.RUnlock()

	serviceMap := make(map[string]ServiceInstanceMap)
	services := make([]*Service, 0, len(serviceMap))

	serviceMap, err := ec.db.ListAllServiceInstances(ec.namespace)
	if err != nil {
		return services
	}

	for serviceName, service := range serviceMap {
		for _, instance := range service {
			if predicate == nil || predicate(instance) {
				services = append(services, &Service{ServiceName: serviceName})
			}
		}
	}

	return services
}

func (ec *externalCatalog) checkIfExpired(instanceID string) {
	ec.Lock()
	defer ec.Unlock()

	instance, err := ec.db.ReadServiceInstanceByInstID(ec.namespace, instanceID)
	if err != nil {
		ec.logger.Debugf("Error reading instance data for instance ID %s", instanceID)
		return
	}
	if instance == nil {
		ec.logger.Debugf("Instance data not found for instance ID %s", instanceID)
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

		timeSinceHeartbeat = time.Now().Sub(instance.LastRenewal)
		if timeSinceHeartbeat <= instance.TTL {
			return
		}

		ec.logger.Debugf("Instance ID %s is expired", instance.ID)
		ec.delete(instanceID)
		ec.expirationMetric.Mark(1)
	}
}

// delete deletes the specified instanceID from the catalog internal datastructures.
// It assumes the catalog's write-lock is acquired by the calling goroutine.
func (ec *externalCatalog) delete(instanceID string) *ServiceInstance {
	instance, err := ec.db.ReadServiceInstanceByInstID(ec.namespace, instanceID)
	if err != nil || instance.ID == "" {
		return nil
	}

	hDel, _ := ec.db.DeleteServiceInstance(ec.namespace, fmt.Sprintf("%s.%s", instance.ID, instance.ServiceName))
	if hDel == 0 {
		return nil
	}

	lifetime := time.Now().Sub(instance.RegistrationTime)
	ec.lifetimeMetric.Update(int64(lifetime))

	hadMetadata := len(instance.Metadata) > 0
	hadTags := len(instance.Tags) > 0

	if hadMetadata {
		ec.metadataInstancesMetric.Dec(1)
	}
	if hadTags {
		ec.tagsInstancesMetric.Dec(1)
	}

	ec.instancesMetric.Dec(1)
	return instance
}

func (ec *externalCatalog) renew(instance *ServiceInstance) {
	instance.LastRenewal = time.Now()

	time.AfterFunc(instance.TTL, func() {
		ec.checkIfExpired(instance.ID)
	})
}
