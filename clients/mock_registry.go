package clients

// MockRegistry mocks interface
type MockRegistry struct {
	GetTokenString string
	GetTokenError  error

	GetServicesArray []string
	GetServicesError error

	GetServiceVal   ServiceInfo
	GetServiceError error

	GetInstancesVal   []Instance
	GetInstancesError error

	CheckUptimeError error
}

// GetInstances mocks behavior of method
func (c *MockRegistry) GetInstances(token string, url string) ([]Instance, error) {
	return c.GetInstancesVal, c.GetInstancesError
}

// GetToken mocks behavior of method
func (c *MockRegistry) GetToken(username, password, org, space string, url string) (string, error) {
	return c.GetTokenString, c.GetTokenError
}

// GetService mocks behavior of method
func (c *MockRegistry) GetService(name, token string, url string) (ServiceInfo, error) {
	return c.GetServiceVal, c.GetServiceError
}

// GetServices mocks behavior of method
func (c *MockRegistry) GetServices(token string, url string) ([]string, error) {
	return c.GetServicesArray, c.GetServicesError
}

// CheckUptime mocks behavior of method
func (c *MockRegistry) CheckUptime(url string) error {
	return c.CheckUptimeError
}
