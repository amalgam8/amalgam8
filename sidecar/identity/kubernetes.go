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

package identity

import (
	"strings"

	"github.com/amalgam8/amalgam8/pkg/api"
	kubepkg "github.com/amalgam8/amalgam8/pkg/kubernetes"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/labels"
	"k8s.io/client-go/pkg/util/intstr"
)

const (
	defaultNamespace = "default"
)

// kubernetesProvider provide access to the identity based on the current Kubernetes pod.
type kubernetesProvider struct {
	podName   string
	namespace string
	client    kubernetes.Interface
}

func newKubernetesProvider(podName string, namespace string, client kubernetes.Interface) (Provider, error) {
	if namespace == "" {
		namespace = defaultNamespace
	}

	return &kubernetesProvider{
		podName:   podName,
		namespace: namespace,
		client:    client,
	}, nil
}

func (kb *kubernetesProvider) GetIdentity() (*api.ServiceInstance, error) {
	pod, err := kb.client.Core().Pods(kb.namespace).Get(kb.podName)
	if err != nil {
		return nil, err
	}

	services, err := kb.client.Core().Services(kb.namespace).List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	service := selectService(pod, services)
	if service == nil {
		return nil, nil
	}

	addr := selectServiceAddress(pod, service)
	port := selectServicePort(pod, service)

	return kubepkg.BuildServiceInstance(service.Name, pod, addr, port)
}

func selectService(pod *v1.Pod, services *v1.ServiceList) *v1.Service {
	matches := make([]v1.Service, 0, 1)
	for _, service := range services.Items {
		selector := labels.SelectorFromValidatedSet(labels.Set(service.Spec.Selector))
		if !selector.Empty() && selector.Matches(labels.Set(pod.Labels)) {
			matches = append(matches, service)
		}
	}

	if len(matches) == 0 {
		logger.Warnf("No Kubernetes service implemented by local pod")
		return nil
	}

	service := matches[0]
	logger.Infof("Kubernetes service implemented by local pod: '%s'", service.Name)

	// In Kubernetes a pod may implement more than one service.
	// Due to the current limitation in which the proxy supports only a single "source" service,
	// We arbitrarily select the first service as the source, and ignore (with a warning) the others
	if len(matches) > 1 {
		ignored := make([]string, len(matches)-1)
		for i := 1; i < len(matches); i++ {
			ignored[i-1] = matches[i].Name
		}
		logger.Warnf("Multiple Kubernetes services implemented by local pod. Ignored services: %s",
			strings.Join(ignored, ", "))
	}

	return &service
}

func selectServiceAddress(pod *v1.Pod, service *v1.Service) *v1.EndpointAddress {
	return &v1.EndpointAddress{
		IP:       pod.Status.PodIP,
		NodeName: &pod.Spec.NodeName,
		TargetRef: &v1.ObjectReference{
			Kind:            "Pod",
			Namespace:       pod.ObjectMeta.Namespace,
			Name:            pod.ObjectMeta.Name,
			UID:             pod.ObjectMeta.UID,
			ResourceVersion: pod.ObjectMeta.ResourceVersion,
		},
	}
}

func selectServicePort(pod *v1.Pod, service *v1.Service) *v1.EndpointPort {
	if len(service.Spec.Ports) == 0 {
		logger.Warnf("No ports defined for Kubernetes service '%s'", service.Name)
		return nil
	}

	if len(service.Spec.Ports) > 1 {
		ignored := make([]string, len(service.Spec.Ports)-1)
		for i := 1; i < len(service.Spec.Ports); i++ {
			ignored[i-1] = service.Spec.Ports[i].TargetPort.String()
		}
		logger.Warnf("Multiple ports defined for service '%s'. Ignored ports: %s",
			service.Name, strings.Join(ignored, ", "))
	}

	port := &service.Spec.Ports[0]
	portNum := findContainerPort(pod, port)

	return &v1.EndpointPort{
		Name:     port.Name,
		Port:     portNum,
		Protocol: port.Protocol,
	}
}

func findContainerPort(pod *v1.Pod, servicePort *v1.ServicePort) int32 {
	switch servicePort.TargetPort.Type {
	case intstr.Int:
		return int32(servicePort.TargetPort.IntValue())
	case intstr.String:
		portName := servicePort.TargetPort.String()
		for _, container := range pod.Spec.Containers {
			for _, containerPort := range container.Ports {
				if containerPort.Name == portName {
					return containerPort.ContainerPort
				}
			}
		}
	default:
		logger.Errorf("Unrecognized port type '%v' for service port '%s'", servicePort.TargetPort.Type, servicePort.TargetPort)
	}

	logger.Warnf("Could not find matching port for service port '%s' on pod %s", servicePort.Name, pod.Name)
	return 0
}
