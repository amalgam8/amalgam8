package checker

import (
	"net/http"

	"github.com/amalgam8/registry/client"
)

// RegistryFactory TODO
type RegistryFactory interface {
	NewRegistryClient(token, url string) (client.Client, error)
}

type registryFactory struct {
	httpClient *http.Client
}

// NewRegistryFactory creates new RegistryFactory interface
func NewRegistryFactory() RegistryFactory {
	return &registryFactory{
		httpClient: &http.Client{},
	}
}

// NewRegistryClient creates new Registry Client
func (r *registryFactory) NewRegistryClient(token, url string) (client.Client, error) {

	registry, err := client.New(client.Config{
		AuthToken:  token,
		URL:        url,
		HTTPClient: r.httpClient,
	})

	return registry, err
}
