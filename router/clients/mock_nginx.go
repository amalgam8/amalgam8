package clients

// MockNginx mocks NGINX client interface
type MockNginx struct {
	UpdateHTTPError error
}

// UpdateHTTPUpstreams mocks interface
func (m *MockNginx) UpdateHTTPUpstreams(conf NGINXJson) error {
	return m.UpdateHTTPError
}
