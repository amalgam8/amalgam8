package checker

import "github.com/amalgam8/controller/resources"

// MockChecker mocks interface
type MockChecker struct {
	RegisterError error

	DeregisterError error

	CheckError error

	GetVal   resources.ServiceCatalog
	GetError error
}

// Register mocks method
func (m *MockChecker) Register(id string) error {
	return m.RegisterError
}

// Deregister mocks method
func (m *MockChecker) Deregister(id string) error {
	return m.DeregisterError
}

// Check mocks method
func (m *MockChecker) Check(ids []string) error {
	return m.CheckError
}

// Get mocks method
func (m *MockChecker) Get(id string) (resources.ServiceCatalog, error) {
	return m.GetVal, m.GetError
}
