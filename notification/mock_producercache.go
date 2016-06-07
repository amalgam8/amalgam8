package notification

import "github.com/amalgam8/controller/resources"

// MockTenantProducerCache mocks interface
type MockTenantProducerCache struct {
	SendEventError error
}

// StartGC mocks method
func (m *MockTenantProducerCache) StartGC() {}

// SendEvent mocks method
func (m *MockTenantProducerCache) SendEvent(tenantID string, kafka resources.Kafka) error {
	return m.SendEventError
}

// Delete mocks method
func (m *MockTenantProducerCache) Delete(tenantID string) {}
