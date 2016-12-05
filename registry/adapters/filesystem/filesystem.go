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

	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/server/protocol/amalgam8"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

const (
	module                               = "FSADAPTER"
	defaultPollingInterval time.Duration = 30 * time.Second
	minPollingInterval     time.Duration = 10 * time.Second
)

// Make sure we implement the ServiceDiscovery interface
var _ api.ServiceDiscovery = (*Adapter)(nil)

type serviceMap map[string][]*api.ServiceInstance // service name -> instance list
type instanceMap map[string]*api.ServiceInstance  // instance ID -> instance

type instancesList struct {
	Instances []*amalgam8.InstanceRegistration `json:"instances"`
}

// Adapter for Filesystem-based Service Discovery
type Adapter struct {
	// The name of the config file that contains the list of the instances
	filename string
	modTime  time.Time

	services  serviceMap
	instances instanceMap

	logger *log.Entry
	sync.RWMutex
}

// Config encapsulates FileSystem configuration parameters
type Config struct {
	Dir             string
	PollingInterval time.Duration
	Namespace       auth.Namespace
}

// New creates and initializes a new Filesystem-based Service Discovery adapter
func New(conf Config) (*Adapter, error) {
	if conf.PollingInterval == 0 {
		conf.PollingInterval = defaultPollingInterval
	}

	if conf.PollingInterval < minPollingInterval {
		conf.PollingInterval = minPollingInterval
	}

	adapter := &Adapter{
		filename:  filepath.Join(conf.Dir, fmt.Sprintf("%s.json", conf.Namespace)),
		services:  serviceMap{},
		instances: instanceMap{},
		logger:    logging.GetLogger(module).WithField("namespace", conf.Namespace),
	}

	adapter.refresh()

	// TODO: more efficient implementation using FileSystem notifications
	ticker := time.NewTicker(conf.PollingInterval)
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
		instances = append(instances, service...)
	}

	return instances, nil
}

// ListServiceInstances queries for the list of service instances currently registered for the given service.
func (a *Adapter) ListServiceInstances(serviceName string) ([]*api.ServiceInstance, error) {
	a.RLock()
	defer a.RUnlock()

	service := a.services[serviceName]
	instances := make([]*api.ServiceInstance, 0, len(service))
	instances = append(instances, service...)

	return instances, nil
}

func (a *Adapter) refresh() {
	a.Lock()
	defer a.Unlock()

	fsinfo, err := os.Stat(a.filename)
	if os.IsNotExist(err) {
		a.services = serviceMap{}
		a.instances = instanceMap{}
		a.modTime = time.Time{}
		return
	}

	if err != nil {
		a.logger.Warnf("Failed to read file %s. %s", a.filename, err)
		a.services = serviceMap{}
		a.instances = instanceMap{}
		return
	}

	if fsinfo.ModTime().After(a.modTime) {
		services, instances, err := a.getServices()
		if err == nil {
			a.services = services
			a.instances = instances
			a.modTime = fsinfo.ModTime()
			a.logger.Debugf("Catalog has been refreshed (%d services, %d instances)", len(a.services), len(a.instances))
		}
	}
}

func (a *Adapter) getServices() (serviceMap, instanceMap, error) {
	svcMap := serviceMap{}
	instMap := instanceMap{}
	data, err := ioutil.ReadFile(a.filename)

	if err != nil {
		a.logger.Warnf("Failed to read config file %s. %s", a.filename, err)
		return svcMap, instMap, err
	}

	instList := instancesList{}
	if err := json.Unmarshal(data, &instList); err != nil {
		a.logger.Warnf("Failed to parse config file %s. %s", a.filename, err)
		return svcMap, instMap, err
	}

	for _, instReg := range instList.Instances {
		var insts []*api.ServiceInstance

		instance := &api.ServiceInstance{
			ID:          computeInstanceID(instReg),
			ServiceName: instReg.ServiceName,
			Endpoint: api.ServiceEndpoint{
				Type:  instReg.Endpoint.Type,
				Value: instReg.Endpoint.Value,
			},
			Tags:     instReg.Tags,
			Status:   instReg.Status,
			Metadata: instReg.Metadata,
		}
		if instance.Status == "" {
			instance.Status = "UP"
		}
		instance.Tags = append(instance.Tags, "filesystem")

		insts = svcMap[instance.ServiceName]
		if insts == nil {
			insts = make([]*api.ServiceInstance, 0, 10)

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
