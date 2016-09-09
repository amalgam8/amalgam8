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

package filesystem

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/api/protocol/amalgam8"
	"github.com/amalgam8/amalgam8/registry/store"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

const (
	module                               = "FSCATALOG"
	defaultPollingInterval time.Duration = 30 * time.Second
	minPollingInterval     time.Duration = 10 * time.Second
)

type serviceMap map[string][]*store.ServiceInstance // service name -> instance list
type instanceMap map[string]*store.ServiceInstance  // instance ID -> instance

type instancesList struct {
	Instances []*amalgam8.InstanceRegistration `json:"instances"`
}

type fsCatalog struct {
	namespace auth.Namespace

	// The name of the config file that contains the list of the instances
	filename string
	modTime  time.Time

	services  serviceMap
	instances instanceMap

	logger *log.Entry
	sync.RWMutex
}

func newFileSystemCatalog(namespace auth.Namespace, conf *Config) (*fsCatalog, error) {
	catalog := &fsCatalog{
		namespace: namespace,
		filename:  filepath.Join(conf.Dir, fmt.Sprintf("%s.conf", namespace)),
		services:  serviceMap{},
		instances: instanceMap{},
		logger:    logging.GetLogger(module).WithField("namespace", namespace),
	}

	catalog.refresh()

	// TODO: more efficient implementation using FileSystem notifications
	ticker := time.NewTicker(conf.PollingInterval)
	go func() {
		for _ = range ticker.C {
			catalog.refresh()
		}
	}()

	return catalog, nil
}

func (fsc *fsCatalog) ListServices(predicate store.Predicate) []*store.Service {
	fsc.RLock()
	defer fsc.RUnlock()

	serviceCollection := make([]*store.Service, 0, len(fsc.services))
	for service, instances := range fsc.services {
		for _, inst := range instances {
			if predicate == nil || predicate(inst) {
				serviceCollection = append(serviceCollection, &store.Service{ServiceName: service})
				break
			}
		}
	}

	return serviceCollection
}

func (fsc *fsCatalog) List(serviceName string, predicate store.Predicate) ([]*store.ServiceInstance, error) {
	fsc.RLock()
	defer fsc.RUnlock()

	service := fsc.services[serviceName]
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

func (fsc *fsCatalog) Instance(instanceID string) (*store.ServiceInstance, error) {
	fsc.RLock()
	defer fsc.RUnlock()

	instance := fsc.instances[instanceID]
	if instance == nil {
		return nil, store.NewError(store.ErrorNoSuchServiceInstance, "no such service instance", instanceID)
	}
	return instance.DeepClone(), nil
}

func (fsc *fsCatalog) Register(si *store.ServiceInstance) (*store.ServiceInstance, error) {
	fsc.logger.Infof("Unsupported API (Register) called")
	return nil, store.NewError(store.ErrorBadRequest, "Read-only Catalog: API Not Supported", "Register")
}

func (fsc *fsCatalog) Deregister(instanceID string) (*store.ServiceInstance, error) {
	fsc.logger.Infof("Unsupported API (Deregister) called")
	return nil, store.NewError(store.ErrorBadRequest, "Read-only Catalog: API Not Supported", "Deregister")
}

func (fsc *fsCatalog) Renew(instanceID string) (*store.ServiceInstance, error) {
	fsc.logger.Infof("Unsupported API (Renew) called")
	return nil, store.NewError(store.ErrorBadRequest, "Read-only Catalog: API Not Supported", "Renew")
}

func (fsc *fsCatalog) SetStatus(instanceID, status string) (*store.ServiceInstance, error) {
	fsc.logger.Infof("Unsupported API (SetStatus) called")
	return nil, store.NewError(store.ErrorBadRequest, "Read-only Catalog: API Not Supported", "SetStatus")
}

func (fsc *fsCatalog) refresh() {
	fsc.Lock()
	defer fsc.Unlock()

	fsinfo, err := os.Stat(fsc.filename)
	if os.IsNotExist(err) {
		fsc.services = serviceMap{}
		fsc.instances = instanceMap{}
		fsc.modTime = time.Time{}
		return
	}

	if err != nil {
		fsc.logger.Warnf("Failed to read file %s. %s", fsc.filename, err)
		fsc.services = serviceMap{}
		fsc.instances = instanceMap{}
		return
	}

	if fsinfo.ModTime().After(fsc.modTime) {
		services, instances, err := fsc.getServices()
		if err == nil {
			fsc.services = services
			fsc.instances = instances
			fsc.modTime = fsinfo.ModTime()
			fsc.logger.Debugf("Catalog [%s] has been refreshed (%d services, %d instances)", fsc.namespace, len(fsc.services), len(fsc.instances))
		}
	}
}

func (fsc *fsCatalog) getServices() (serviceMap, instanceMap, error) {
	svcMap := serviceMap{}
	instMap := instanceMap{}
	data, err := ioutil.ReadFile(fsc.filename)

	if err != nil {
		fsc.logger.Warnf("Failed to read config file %s. %s", fsc.filename, err)
		return svcMap, instMap, err
	}

	instList := instancesList{}
	if err := json.Unmarshal(data, &instList); err != nil {
		fsc.logger.Warnf("Failed to parse config file %s. %s", fsc.filename, err)
		return svcMap, instMap, err
	}

	for _, instReg := range instList.Instances {
		var insts []*store.ServiceInstance

		instance := &store.ServiceInstance{
			ID:               computeInstanceID(instReg),
			ServiceName:      instReg.ServiceName,
			Endpoint:         &store.Endpoint{Type: instReg.Endpoint.Type, Value: instReg.Endpoint.Value},
			Tags:             instReg.Tags,
			Status:           instReg.Status,
			Metadata:         instReg.Metadata,
			RegistrationTime: time.Now(),
		}
		if instance.Status == "" {
			instance.Status = "UP"
		}
		instance.Tags = append(instance.Tags, "filesystem")

		insts = svcMap[instance.ServiceName]
		if insts == nil {
			insts = make([]*store.ServiceInstance, 0, 10)

		}
		svcMap[instance.ServiceName] = append(insts, instance)
		instMap[instance.ID] = instance
	}

	return svcMap, instMap, nil
}

func computeInstanceID(instReg *amalgam8.InstanceRegistration) string {
	hash := sha256.New()
	hash.Write([]byte(strings.Join([]string{instReg.ServiceName, instReg.Endpoint.Type, instReg.Endpoint.Value}, "/")))
	md := hash.Sum(nil)
	mdStr := hex.EncodeToString(md)
	return mdStr[:16]
}
