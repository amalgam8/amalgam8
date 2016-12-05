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

package kubernetes

import (
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

const (
	module = "K8SADAPTER"
)

// Make sure we implement the ServiceDiscovery interface
var _ api.ServiceDiscovery = (*Adapter)(nil)

// Config stores configurable attributes of the kubernetes adapter
type Config struct {
	URL       string
	Token     string
	Namespace auth.Namespace
}

// Adapter for Eureka Service Discovery
type Adapter struct {
	namespace auth.Namespace
	client    *client

	services map[string][]*api.ServiceInstance

	logger *log.Entry
	sync.RWMutex
}

// New creates and initializes a new Kubernetes Service Discovery adapter
func New(config Config) (*Adapter, error) {
	client, err := newClient(config.URL, config.Token)
	if err != nil {
		return nil, err
	}

	catalog := &Adapter{
		services:  map[string][]*api.ServiceInstance{},
		namespace: config.Namespace,
		client:    client,
		logger:    logging.GetLogger(module),
	}

	catalog.refresh()

	// TODO: more efficient implementation using Kubernetes API watch interface
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for _ = range ticker.C {
			catalog.refresh()
		}
	}()

	return catalog, nil
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

	services, err := a.getServices()
	if err == nil {
		a.services = services
	}
}

func (a *Adapter) getServices() (map[string][]*api.ServiceInstance, error) {
	services := map[string][]*api.ServiceInstance{}
	endpointsList, err := a.client.getEndpointsList(a.namespace)
	if err != nil {
		a.logger.Warnf("Unable to get endpoints: %s", err)
		return services, err
	}
	for _, endpoints := range endpointsList.Items {
		sname := endpoints.ObjectMeta.Name
		insts := []*api.ServiceInstance{}
		for _, subset := range endpoints.Subsets {
			for _, address := range subset.Addresses {
				for _, port := range subset.Ports {
					var uid string
					var version string
					var tags []string

					// Parse the service endpoint
					endpoint, err := parseEndpoint(address, port)
					if err != nil {
						a.logger.WithError(err).Warningf("Skipping endpoint %s for service %s", address.TargetRef.Name, sname)
						continue
					}

					// Parse UID and version out of the pod name
					if address.TargetRef != nil {
						uid = address.TargetRef.UID
						podName := address.TargetRef.Name
						rcName := podName[:strings.LastIndex(podName, "-")]
						versionIndex := strings.LastIndex(rcName, "-")
						if versionIndex != -1 {
							version = rcName[versionIndex+1:]
						}
					} else {
						uid = address.IP
					}

					// Tag the service instance with the version
					if version != "" {
						tags = append(tags, version)
					}
					tags = append(tags, "kubernetes")

					inst := &api.ServiceInstance{
						ID:          fmt.Sprintf("%s-%d", uid, port.Port),
						ServiceName: sname,
						Endpoint:    *endpoint,
						Status:      "UP",
						Tags:        tags,
						TTL:         0,
					}
					insts = append(insts, inst)
				}
			}
			services[sname] = insts
		}
	}
	return services, nil
}

func parseEndpoint(address EndpointAddress, port EndpointPort) (*api.ServiceEndpoint, error) {
	var endpointType string
	endpointValue := fmt.Sprintf("%s:%d", address.IP, port.Port)

	switch port.Protocol {
	case ProtocolUDP:
		endpointType = "udp"
	case ProtocolTCP:
		portName := strings.ToLower(port.Name)
		switch portName {
		case "http":
			fallthrough
		case "https":
			endpointType = portName
		default:
			endpointType = "tcp"
		}
	default:
		return nil, fmt.Errorf("unsupported kubernetes endpoint port protocol: %s", port.Protocol)
	}

	return &api.ServiceEndpoint{
		Type:  endpointType,
		Value: endpointValue,
	}, nil
}
