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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/pkg/datastructures"
	"github.com/amalgam8/amalgam8/pkg/errors"
	"github.com/amalgam8/amalgam8/registry/api"
	"github.com/amalgam8/amalgam8/registry/utils/logging"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	// EndpointsCacheResyncPeriod is the period in which we do a full resync of the endpoints cache.
	EndpointsCacheResyncPeriod = time.Duration(60) * time.Second

	// PodCacheResyncPeriod is the period in which we do a full resync of the pod cache.
	PodCacheResyncPeriod = time.Duration(60) * time.Second
)

// Make sure we implement the ServiceDiscovery interface
var _ api.ServiceDiscovery = (*Adapter)(nil)

// Package global logger
var logger = logging.GetLogger("KUBERNETES")

// Config stores configurable attributes of the Kubernetes adapter.
type Config struct {
	URL       string
	Token     string
	Namespace auth.Namespace

	// Client to be used by the Kubernetes adapter.
	// If no client is provided, then a client is created
	// according the specified URL/Token/Namespace, if provided,
	// or from the local service account, if running within a Kubernetes pod.
	Client *rest.RESTClient
}

// Adapter for Kubernetes Service Discovery.
type Adapter struct {
	endpointsCache      cache.Store
	endpointsController cache.ControllerInterface

	podCache      cache.Store
	podController cache.ControllerInterface

	// services maps a service name to a list of service instances.
	// This is stored precomputed so we won't have to recompute it
	// with every ListInstance()/ListServiceInstances() call.
	services map[string][]*api.ServiceInstance

	// servicePods maps a service name to a set of pod names implementing it.
	// podServices maps a pod name to a set of service names implemented by it.
	// These are maintained mainly for supporting dynamic label updates on pods.
	servicePods map[string]datastructures.StringSet
	podServices map[string]datastructures.StringSet

	namespace string

	stopChan chan struct{}
	mutex    sync.RWMutex
}

// New creates and starts a new Kubernetes Service Discovery adapter.
func New(config Config) (*Adapter, error) {
	var client *rest.RESTClient
	if config.Client != nil {
		client = config.Client
	} else {
		var err error
		client, err = buildClientFromConfig(config)
		if err != nil {
			return nil, err
		}
	}

	// If no namespace is specified, fallback to default namespace
	namespace := config.Namespace.String()
	if namespace == "" {
		namespace = "default"
	}

	adapter := &Adapter{
		services:    make(map[string][]*api.ServiceInstance),
		podServices: make(map[string]datastructures.StringSet),
		servicePods: make(map[string]datastructures.StringSet),
		namespace:   namespace,
	}

	adapter.endpointsCache, adapter.endpointsController = cache.NewInformer(
		cache.NewListWatchFromClient(client, "endpoints", namespace, nil),
		&v1.Endpoints{},
		EndpointsCacheResyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    adapter.addEndpoints,
			UpdateFunc: adapter.updateEndpoints,
			DeleteFunc: adapter.deleteEndpoints,
		},
	)

	adapter.podCache, adapter.podController = cache.NewInformer(
		cache.NewListWatchFromClient(client, "pods", namespace, nil),
		&v1.Pod{},
		PodCacheResyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    adapter.addPod,
			UpdateFunc: adapter.updatePod,
			DeleteFunc: adapter.deletePod,
		},
	)

	return adapter, adapter.Start()
}

// Start synchronizing the Kubernetes adapter.
func (a *Adapter) Start() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.stopChan != nil {
		err := fmt.Errorf("kubernetes adapter already started")
		logger.WithError(err).Errorf("Failed starting Kubernetes adapter")
		return err
	}
	a.stopChan = make(chan struct{})

	go a.endpointsController.Run(a.stopChan)
	go a.podController.Run(a.stopChan)

	return nil
}

// Stop synchronizing the Kubernetes adapter.
func (a *Adapter) Stop() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.stopChan == nil {
		err := fmt.Errorf("kubernetes adapter not started")
		logger.WithError(err).Errorf("Failed stopping Kubernetes adapter")
		return err
	}
	close(a.stopChan)
	a.stopChan = nil

	return nil
}

// ListServices queries for the list of services for which instances are currently registered.
func (a *Adapter) ListServices() ([]string, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	services := make([]string, 0, len(a.services))
	for service := range a.services {
		services = append(services, service)
	}

	return services, nil
}

// ListInstances queries for the list of service instances currently registered.
func (a *Adapter) ListInstances() ([]*api.ServiceInstance, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	instances := make([]*api.ServiceInstance, 0, len(a.services)*3)
	for _, service := range a.services {
		instances = append(instances, service...)
	}

	return instances, nil
}

// ListServiceInstances queries for the list of service instances currently registered for the given service.
func (a *Adapter) ListServiceInstances(serviceName string) ([]*api.ServiceInstance, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	service := a.services[serviceName]
	instances := make([]*api.ServiceInstance, 0, len(service))
	instances = append(instances, service...)

	return instances, nil
}

// addEndpoints is the callback invoked by the Kubernetes cache when an endpoints API resource is added.
func (a *Adapter) addEndpoints(obj interface{}) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	endpoints, ok := obj.(*v1.Endpoints)
	if !ok {
		logger.Warnf("Invalid endpoint added: object is of type %T", obj)
		return
	}

	logger.Debugf("Endpoints object added: %s", endpoints.Name)
	a.reloadServiceFromEndpoints(endpoints)
}

// updateEndpoints is the callback invoked by the Kubernetes cache when an endpoints API resource is updated.
func (a *Adapter) updateEndpoints(oldObj, newObj interface{}) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	endpoints, ok := newObj.(*v1.Endpoints)
	if !ok {
		logger.Warnf("Invalid endpoint update: new object is of type %T", newObj)
		return
	}

	logger.Debugf("Endpoints object updated: %s", endpoints.Name)
	a.reloadServiceFromEndpoints(endpoints)
}

// deleteEndpoints is the callback invoked by the Kubernetes cache when an endpoints API resource is deleted.
func (a *Adapter) deleteEndpoints(obj interface{}) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	endpoints, ok := extractDeletedObject(obj).(*v1.Endpoints)
	if !ok {
		logger.Warnf("Invalid endpoint deleted: object is of type %T", obj)
		return
	}

	logger.Debugf("Endpoints object deleted: %s", endpoints.Name)
	a.deleteService(endpoints.Name)
}

// addPod is the callback invoked by the Kubernetes cache when a pod API resource is added.
func (a *Adapter) addPod(obj interface{}) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	pod, ok := obj.(*v1.Pod)
	if !ok {
		logger.Warnf("Invalid pod added: object is of type %T", obj)
		return
	}

	// Reload any services implemented by the pod
	services := a.podServices[pod.Name]
	for service := range services {
		a.reloadServiceFromCache(service)
	}
}

// updatePod is the callback invoked by the Kubernetes cache when a pod API resource is updated.
func (a *Adapter) updatePod(oldObj, newObj interface{}) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	oldPod, ok := oldObj.(*v1.Pod)
	if !ok {
		logger.Warnf("Invalid pod update: old object is of type %T", oldObj)
		return
	}
	newPod, ok := newObj.(*v1.Pod)
	if !ok {
		logger.Warnf("Invalid pod update: new object is of type %T", newObj)
		return
	}

	// If no labels have changed, ignore
	if reflect.DeepEqual(oldPod.Labels, newPod.Labels) {
		return
	}

	// Reload any services implemented by the pod
	services := a.podServices[newPod.Name]
	for service := range services {
		a.reloadServiceFromCache(service)
	}
}

// deletePod is the callback invoked by the Kubernetes cache when a pod API resource is deleted.
func (a *Adapter) deletePod(obj interface{}) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	pod, ok := extractDeletedObject(obj).(*v1.Pod)
	if !ok {
		logger.Warnf("Invalid pod deleted: object is of type %T", obj)
		return
	}

	delete(a.podServices, pod.Name)
}

// reloadServiceFromCache rebuilds and stores the service instances for the given service,
// based on the cached service endpoints and pods resources.
func (a *Adapter) reloadServiceFromCache(serviceName string) {
	endpoints := a.getCachedServiceEndpoints(serviceName)
	if endpoints == nil {
		logger.Warnf("No endpoints cached for service '%s'", serviceName)
		return
	}

	a.reloadServiceFromEndpoints(endpoints)
}

// reloadServiceFromEndpoints rebuilds and stores the service instances for the given endpoints service,
// based on the given endpoints information, and cached pod resources.
func (a *Adapter) reloadServiceFromEndpoints(endpoints *v1.Endpoints) {
	serviceName := endpoints.Name
	instances := []*api.ServiceInstance{}
	pods := datastructures.NewDefaultStringSet()

	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			for _, port := range subset.Ports {
				instance, err := a.createServiceInstance(serviceName, address, port)
				if err != nil {
					logger.WithError(err).Warnf("Failed creating service '%s' instance for pod %s with address %s and port %s",
						serviceName, address.TargetRef.Name, address.String(), port.String())
					continue
				}
				instances = append(instances, instance)
			}

			if address.TargetRef != nil {
				pods.Add(address.TargetRef.Name)
			}
		}
	}

	a.services[endpoints.Name] = instances

	prevPods := a.servicePods[endpoints.Name]
	a.servicePods[endpoints.Name] = pods

	for pod := range prevPods.Difference(pods) {
		podServices := a.podServices[pod]
		if podServices != nil {
			podServices.Remove(endpoints.Name)
		}
	}
	for pod := range pods.Difference(prevPods) {
		podServices := a.podServices[pod]
		if podServices == nil {
			podServices = datastructures.NewDefaultStringSet()
			a.podServices[pod] = podServices
		}
		podServices.Add(endpoints.Name)
	}
}

// deleteService deletes any stored service instance for the given service.
func (a *Adapter) deleteService(serviceName string) {
	delete(a.services, serviceName)
}

// createServiceInstance creates a service instance based on the given service name, address and port.
// Cached pod information is used to build the tags and metadata fields.
func (a *Adapter) createServiceInstance(serviceName string, address v1.EndpointAddress, port v1.EndpointPort) (*api.ServiceInstance, error) {
	// Parse the service endpoint
	endpoint, err := buildEndpointFromAddress(address, port)
	if err != nil {
		return nil, err
	}

	// Extract the pod implementing the service
	var pod *v1.Pod
	if address.TargetRef != nil {
		pod = a.getCachedPod(address.TargetRef.Name)
	}

	// Determine the ID of the service instance.
	// For a pod, that would be the pod UID, followed by the port number.
	// For an externalName, that would be the IP address, followed by the port number.
	var id string
	if pod != nil {
		id = fmt.Sprintf("%s-%d", pod.UID, port.Port)
	} else {
		id = fmt.Sprintf("%s-%d", address.IP, port.Port)
	}

	// Extract the pod labels as instance tags
	var tags []string
	if pod != nil {
		tags = buildTagsFromLabels(pod.Labels)
	}

	// Extract the pod annotations as instance metadata
	var metadata json.RawMessage
	if pod != nil {
		metadata = buildMetadataFromAnnotations(pod.Annotations)
	}

	return &api.ServiceInstance{
		ID:          id,
		ServiceName: serviceName,
		Endpoint:    *endpoint,
		Tags:        tags,
		Metadata:    metadata,
		Status:      "UP", // Otherwise would be removed from the service endpoints
	}, nil
}

// getCachedServiceEndpoints returns the cached endpoints resource for the given service, or nil if doesn't exist.
func (a *Adapter) getCachedServiceEndpoints(serviceName string) *v1.Endpoints {
	key, err := cache.MetaNamespaceKeyFunc(&v1.Endpoints{
		ObjectMeta: v1.ObjectMeta{
			Namespace: a.namespace,
			Name:      serviceName,
		},
	})
	if err != nil {
		return nil
	}
	obj, exist, err := a.endpointsCache.GetByKey(key)
	if err != nil || !exist {
		return nil
	}
	endpoints, ok := obj.(*v1.Endpoints)
	if !ok {
		return nil
	}
	return endpoints
}

// getCachedPod returns the cached pod resource for the given pod name, or nil if doesn't exist.
func (a *Adapter) getCachedPod(podName string) *v1.Pod {
	key, err := cache.MetaNamespaceKeyFunc(&v1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Namespace: a.namespace,
			Name:      podName,
		},
	})
	if err != nil {
		return nil
	}
	obj, exist, err := a.podCache.GetByKey(key)
	if err != nil || !exist {
		return nil
	}
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return nil
	}
	return pod
}

// extractDeletedObject is used within "deleteXXX" cache callbacks, where the provided
// object may be a wrapper (DeletedFinalStateUnknown) around the actual deleted object.
func extractDeletedObject(obj interface{}) interface{} {
	deleted, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		return deleted.Obj
	}
	return obj
}

// buildEndpointFromAddress builds an api.ServiceEndpoint from the given address and port Kubernetes objects.
func buildEndpointFromAddress(address v1.EndpointAddress, port v1.EndpointPort) (*api.ServiceEndpoint, error) {
	var endpointType string
	endpointValue := fmt.Sprintf("%s:%d", address.IP, port.Port)

	switch port.Protocol {
	case v1.ProtocolUDP:
		endpointType = "udp"
	case v1.ProtocolTCP:
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

// buildTagsFromLabels builds a slice of string tags from the given resource labels.
// Each label is converted into a "key=value" string.
func buildTagsFromLabels(labels map[string]string) []string {
	tags := make([]string, 0, len(labels))
	for key, value := range labels {
		tags = append(tags, fmt.Sprintf("%s=%s", key, value))
	}
	return tags
}

// buildMetadataFromAnnotations builds a serialized JSON object from the given resource annotations.
func buildMetadataFromAnnotations(annotations map[string]string) json.RawMessage {
	bytes, err := json.Marshal(annotations)
	if err != nil {
		logger.WithError(err).Errorf("Error marshaling annotations to JSON")
		return nil
	}
	return json.RawMessage(bytes)
}

// buildClientFromConfig creates a new Kubernetes REST client based on the given configuration.
// If no URL and Token are specified, then these values are attempted to be read from the
// service account (if running within a Kubernetes pod).
func buildClientFromConfig(config Config) (*rest.RESTClient, error) {
	var kubeConfig *rest.Config
	if config.URL != "" || config.Token != "" {
		kubeConfig = &rest.Config{
			Host:        config.URL,
			BearerToken: config.Token,
		}
	} else {
		logger.Debugf("No Kubernetes credentials provided. Attempting to load from service account")
		var err error
		kubeConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, errors.Wrap(err, "Failed loading Kubernetes credentials from service account")
		}
	}

	client, err := rest.RESTClientFor(kubeConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Failed creating Kubernetes REST client")
	}
	return client, nil
}
