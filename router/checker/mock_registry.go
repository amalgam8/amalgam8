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

package checker

import "github.com/amalgam8/registry/client"

// MockRegistryFactory mocks RegistryFactory interface
type MockRegistryFactory struct {
	RegClient *MockRegistryClient
}

// NewRegistryClient mocks interface
func (r *MockRegistryFactory) NewRegistryClient(token, url string) (client.Client, error) {
	return r.RegClient, nil
}

// MockRegistryClient mocks Registry Client interface
type MockRegistryClient struct {
	ListServicesVal []string
	ListServicesErr error

	ListInstancesVal []*client.ServiceInstance
	ListInstancesErr error

	ListServiceInstancesVal []*client.ServiceInstance
	ListServiceInstancesErr error
}

// Register mocks interface
func (m *MockRegistryClient) Register(instance *client.ServiceInstance) (*client.ServiceInstance, error) {
	return &client.ServiceInstance{}, nil
}

// Deregister mocks interface
func (m *MockRegistryClient) Deregister(id string) error {
	return nil
}

// Renew mocks interface
func (m *MockRegistryClient) Renew(id string) error {
	return nil
}

// ListServices mocks interface
func (m *MockRegistryClient) ListServices() ([]string, error) {
	return m.ListServicesVal, m.ListServicesErr
}

// ListInstances mocks interface
func (m *MockRegistryClient) ListInstances(filter client.InstanceFilter) ([]*client.ServiceInstance, error) {
	return m.ListInstancesVal, m.ListInstancesErr
}

// ListServiceInstances mocks interface
func (m *MockRegistryClient) ListServiceInstances(serviceName string) ([]*client.ServiceInstance, error) {
	return m.ListServiceInstancesVal, m.ListServiceInstancesErr
}
