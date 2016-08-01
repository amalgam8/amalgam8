package checker

import (
	"github.com/amalgam8/controller/resources"
	"github.com/amalgam8/sidecar/router/clients"
	"github.com/amalgam8/sidecar/router/nginx"
)

type MockListener struct {
	mockNginx nginx.Nginx
}

func (m *MockListener) CatalogChange(catalog resources.ServiceCatalog) error {
	return m.mockNginx.Update(clients.NGINXJson{})
}

func (m *MockListener) RulesChange(proxyConfig resources.ProxyConfig) error {
	return m.mockNginx.Update(clients.NGINXJson{})
}
