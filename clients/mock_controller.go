package clients

import "time"

// MockController mocks the Controller interface
type MockController struct {
	RegisterError error
	ConfigError   error
	ConfigString  string
	GetCredsError error
	GetCredsVal   TenantCredentials
}

// Register mocks interface
func (m *MockController) Register() error {
	return m.RegisterError
}

// GetNGINXConfig mocks interface
func (m *MockController) GetNGINXConfig(version *time.Time) (string, error) {
	return m.ConfigString, m.ConfigError
}

// GetCredentials mocks interface
func (m *MockController) GetCredentials() (TenantCredentials, error) {
	return m.GetCredsVal, m.GetCredsError
}
