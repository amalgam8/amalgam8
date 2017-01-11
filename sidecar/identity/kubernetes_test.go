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
	"testing"

	"fmt"

	"strconv"
	"strings"

	"github.com/stretchr/testify/suite"
	fakekubernetes "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
	fakecore "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/util/intstr"
)

const (
	testKubernetesPodName   = "my-pod"
	testKubernetesNamespace = "my-namespace"
	testKubernetesPodIP     = "192.168.10.16"
)

var (
	clientset     *mockClientset
	coreClient    *mockCoreClient
	podClient     *mockPodClient
	serviceClient *mockServiceClient
)

type KubernetesProviderSuite struct {
	suite.Suite

	provider Provider
}

func TestKubernetesIdentityProviderSuite(t *testing.T) {
	suite.Run(t, new(KubernetesProviderSuite))
}

func (s *KubernetesProviderSuite) SetupTest() {
	var err error

	clientset = &mockClientset{}
	coreClient = &mockCoreClient{}
	podClient = &mockPodClient{}
	serviceClient = &mockServiceClient{}

	s.provider, err = newKubernetesProvider(testKubernetesPodName, testKubernetesNamespace, clientset)
	s.Require().NoError(err, "Error creating Kubernetes Identity Provider")
	s.Require().NotNil(s.provider, "Kubernetes Identity Provider should not be nil")
}

func (s *KubernetesProviderSuite) TestPodQueryError() {
	podClient.err = fmt.Errorf("error querying for pod")

	si, err := s.provider.GetIdentity()
	s.Require().Error(err, "Expected GetIdentity() to fail due to pod query error")
	s.Require().Nil(si, "Expected a nil service instance upon error")
}

func (s *KubernetesProviderSuite) TestServiceQueryError() {
	serviceClient.err = fmt.Errorf("error querying for services")

	si, err := s.provider.GetIdentity()
	s.Require().Error(err, "Expected GetIdentity() to fail due to services query error")
	s.Require().Nil(si, "Expected a nil service instance upon error")
}

func (s *KubernetesProviderSuite) TestNoIdentityIfNoServicesExist() {
	podClient.pod = createPod("service: service-X", "http:tcp:8080")

	si, err := s.provider.GetIdentity()
	s.Require().NoError(err, "Expected GetIdentity() to not fail")
	s.Require().Nil(si, "Expected a nil service instance when no services exist")
}

func (s *KubernetesProviderSuite) TestNoIdentityIfNoServicesImplemented() {
	podClient.pod = createPod("service: service-X", "http:tcp:8080")
	serviceClient.services = createServiceList(createService("service-Y", "http:tcp:8080:8080"))

	si, err := s.provider.GetIdentity()
	s.Require().NoError(err, "Expected GetIdentity() to not fail")
	s.Require().Nil(si, "Expected a nil service instance when no services implemented")
}

func (s *KubernetesProviderSuite) TestIdentitySingleServiceImplemented() {
	podClient.pod = createPod("service: service-X", "http:tcp:8080")
	serviceClient.services = createServiceList(createService("service-X", "http:tcp:80:8080"))

	si, err := s.provider.GetIdentity()
	s.Require().NoError(err, "Expected GetIdentity() to not fail")
	s.Require().NotNil(si, "Expected a non-nil service instance")

	s.Require().Equal("service-X", si.ServiceName, "Unexpected service name")
}

func (s *KubernetesProviderSuite) TestIdentityMiddleServiceImplemented() {
	podClient.pod = createPod("service: service-Y", "http:tcp:8081")
	serviceClient.services = createServiceList(
		createService("service-X", "http:tcp:80:8080"),
		createService("service-Y", "http:tcp:81:8081"),
		createService("service-Z", "http:tcp:82:8082"),
	)

	si, err := s.provider.GetIdentity()
	s.Require().NoError(err, "Expected GetIdentity() to not fail")
	s.Require().NotNil(si, "Expected a non-nil service instance")

	s.Require().Equal("service-Y", si.ServiceName, "Unexpected service name")
}

func (s *KubernetesProviderSuite) TestIdentityMultipleServicesImplemented() {
	podClient.pod = createPod("service1: service-X, service2: service-Y, service3: service-Z",
		"http:tcp:8080, http:tcp:8081, http:tcp:8082")
	serviceClient.services = createServiceList(
		createServiceWithSelector("service-X", "service1: service-X", "http:tcp:80:8080"),
		createServiceWithSelector("service-Y", "service2: service-Y", "http:tcp:81:8081"),
		createServiceWithSelector("service-Z", "service3: service-Z", "http:tcp:82:8082"),
	)

	si, err := s.provider.GetIdentity()
	s.Require().NoError(err, "Expected GetIdentity() to not fail")
	s.Require().NotNil(si, "Expected a non-nil service instance")

	s.Require().Equal("service-X", si.ServiceName, "Unexpected service name")
}

func (s *KubernetesProviderSuite) TestIdentityMultipleLabels() {
	podClient.pod = createPod("service: service-X, version: v1, env: production", "http:tcp:8080")
	serviceClient.services = createServiceList(createService("service-X", "http:tcp:80:8080"))

	si, err := s.provider.GetIdentity()
	s.Require().NoError(err, "Expected GetIdentity() to not fail")
	s.Require().NotNil(si, "Expected a non-nil service instance")

	s.Require().Len(si.Tags, len(podClient.pod.Labels), "Unexpected number of service instance tags")
	s.Require().Contains(si.Tags, "service=service-X", "Expected tag not found")
	s.Require().Contains(si.Tags, "version=v1", "Expected tag not found")
	s.Require().Contains(si.Tags, "env=production", "Expected tag not found")
}

func (s *KubernetesProviderSuite) TestIdentityEndpoint() {
	podClient.pod = createPod("service: service-X", "http:tcp:8080")
	serviceClient.services = createServiceList(createService("service-X", "http:tcp:80:8080"))

	si, err := s.provider.GetIdentity()
	s.Require().NoError(err, "Expected GetIdentity() to not fail")
	s.Require().NotNil(si, "Expected a non-nil service instance")

	s.Require().Equal("http", si.Endpoint.Type)
	s.Require().Equal(fmt.Sprintf("%s:%d", testKubernetesPodIP, 8080), si.Endpoint.Value)
}

func (s *KubernetesProviderSuite) TestIdentityEndpointMultipleContainerPorts() {
	podClient.pod = createPod("service: service-X", "http:tcp:8080, https:tcp:8443")
	serviceClient.services = createServiceList(createService("service-X", "https:tcp:443:8443"))

	si, err := s.provider.GetIdentity()
	s.Require().NoError(err, "Expected GetIdentity() to not fail")
	s.Require().NotNil(si, "Expected a non-nil service instance")

	s.Require().Equal("https", si.Endpoint.Type)
	s.Require().Equal(fmt.Sprintf("%s:%d", testKubernetesPodIP, 8443), si.Endpoint.Value)
}

func (s *KubernetesProviderSuite) TestIdentityEndpointNamedTargetPort() {
	podClient.pod = createPod("service: service-X", "http:tcp:8080")
	serviceClient.services = createServiceList(createService("service-X", "http:tcp:80:http"))

	si, err := s.provider.GetIdentity()
	s.Require().NoError(err, "Expected GetIdentity() to not fail")
	s.Require().NotNil(si, "Expected a non-nil service instance")

	s.Require().Equal("http", si.Endpoint.Type)
	s.Require().Equal(fmt.Sprintf("%s:%d", testKubernetesPodIP, 8080), si.Endpoint.Value)
}

type mockClientset struct {
	fakekubernetes.Clientset
}

type mockCoreClient struct {
	fakecore.FakeCore
}

type mockPodClient struct {
	fakecore.FakePods
	pod v1.Pod
	err error
}

type mockServiceClient struct {
	fakecore.FakeServices
	services v1.ServiceList
	err      error
}

func (mcs *mockClientset) Core() core.CoreInterface {
	return coreClient
}

func (mcs *mockCoreClient) Pods(namespace string) core.PodInterface {
	return podClient
}

func (mcs *mockCoreClient) Services(namespace string) core.ServiceInterface {
	return serviceClient
}

func (mpc *mockPodClient) Get(name string) (*v1.Pod, error) {
	if mpc.err != nil {
		return nil, mpc.err
	}

	pod := mpc.pod
	return &pod, nil
}

func (msc *mockServiceClient) List(opts v1.ListOptions) (*v1.ServiceList, error) {
	if msc.err != nil {
		return nil, msc.err
	}

	services := msc.services
	return &services, nil
}

// createPod creates a pod with the given labels and ports.
// labels is a comma-separated list of "key:value" elements.
// ports is a comma-separated list of "name:protocol:portNum" elements.
func createPod(labels string, ports string) v1.Pod {
	pod := v1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      testKubernetesPodName,
			Namespace: testKubernetesNamespace,
			Labels:    map[string]string{},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Ports: []v1.ContainerPort{},
				},
			},
		},
		Status: v1.PodStatus{
			PodIP: testKubernetesPodIP,
		},
	}

	for _, label := range splitTrim(labels, ",") {
		parts := splitTrim(label, ":")
		pod.Labels[parts[0]] = parts[1]
	}

	for _, port := range splitTrim(ports, ",") {
		var containerPort v1.ContainerPort

		parts := splitTrim(port, ":")

		containerPort.Name = parts[0]
		containerPort.Protocol = v1.Protocol(strings.ToUpper(parts[1]))

		cp, _ := strconv.Atoi(parts[2])
		containerPort.ContainerPort = int32(cp)

		pod.Spec.Containers[0].Ports = append(pod.Spec.Containers[0].Ports, containerPort)
	}

	return pod
}

// createService creates a service with the given name and ports.
// The service selector is "service:<name>".
// ports is a comma-separated list of "name:protocol:servicePort:targetPort" elements.
func createService(name string, ports string) v1.Service {
	return createServiceWithSelector(name, "service:"+name, ports)
}

// createServiceWithSelector creates a service with the given name, selector and ports.
// selector is a comma-separated list of "key:value" elements.
// ports is a comma-separated list of "name:protocol:servicePort:targetPort" elements.
func createServiceWithSelector(name string, selector string, ports string) v1.Service {
	service := v1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: testKubernetesNamespace,
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{},
			Ports:    []v1.ServicePort{},
		},
	}

	for _, label := range splitTrim(selector, ",") {
		parts := splitTrim(label, ":")
		service.Spec.Selector[parts[0]] = parts[1]
	}

	for _, port := range splitTrim(ports, ",") {
		var servicePort v1.ServicePort

		parts := splitTrim(port, ":")

		servicePort.Name = parts[0]
		servicePort.Protocol = v1.Protocol(strings.ToUpper(parts[1]))

		sp, _ := strconv.Atoi(parts[2])
		servicePort.Port = int32(sp)

		if targetPortNum, err := strconv.Atoi(parts[3]); err != nil {
			servicePort.TargetPort = intstr.FromString(parts[3])
		} else {
			servicePort.TargetPort = intstr.FromInt(targetPortNum)
		}

		service.Spec.Ports = append(service.Spec.Ports, servicePort)
	}

	return service
}

func createServiceList(services ...v1.Service) v1.ServiceList {
	return v1.ServiceList{
		Items: services,
	}
}

func splitTrim(s string, sep string) []string {
	parts := strings.Split(s, sep)
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}
