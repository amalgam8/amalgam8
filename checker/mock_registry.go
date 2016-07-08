package checker

import "github.com/amalgam8/registry/client"

// MockRegistryFactory mocks RegistryFactory interface
type MockRegistyFactory struct {
	RegClient *MockRegistryClient
}

// NewRegistryClient mocks interface
func (r *MockRegistyFactory) NewRegistryClient(token, url string) (client.Client, error) {
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
