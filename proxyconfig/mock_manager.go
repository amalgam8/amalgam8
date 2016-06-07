package proxyconfig

import "github.com/amalgam8/controller/resources"

// MockRules mocks interface
type MockManager struct {
	SetError    error
	GetVal      resources.ProxyConfig
	GetError    error
	DeleteError error
}

// Set mocks method
func (m *MockManager) Set(rules resources.ProxyConfig) error {
	return m.SetError
}

// Get mocks method
func (m *MockManager) Get(id string) (resources.ProxyConfig, error) {
	return m.GetVal, m.GetError
}

// Delete mocks method
func (m *MockManager) Delete(id string) error {
	return m.DeleteError
}
