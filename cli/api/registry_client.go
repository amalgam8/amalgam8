package api

import (
	// "fmt"
	"fmt"
	api "github.com/amalgam8/amalgam8/cli/client"
	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/utils"
	"github.com/urfave/cli"
	"net/http"
)

type registry struct {
	ctx    *cli.Context
	client api.Client
	debug  bool
}

// RegistryClient .
type RegistryClient interface {
	Services() (*ServiceList, error)
	ServiceInstances(service string) (*InstanceList, error)
}

// NewRegistryClient .
func NewRegistryClient(ctx *cli.Context) (RegistryClient, error) {
	url, err := ValidateRegistryURL(ctx)
	if err != nil {
		fmt.Fprintf(ctx.App.Writer, fmt.Sprintf("\n%s: %q\n\n", err.Error(), url))
		return nil, err
	}

	// Read Token
	token := ctx.GlobalString(common.RegistryToken.Flag())

	// Check if a custom client has been set
	var client *http.Client
	if c, ok := ctx.App.Metadata["httpClient"].(*http.Client); ok {
		client = c
	}

	return &registry{
		debug:  ctx.GlobalBool(common.Debug.Flag()),
		client: api.NewClient(url, token, client),
	}, nil
}

// Services .
func (r *registry) Services() (*ServiceList, error) {
	services := &ServiceList{}
	err := r.client.GET(servicesPath, r.debug, nil, services)
	if err != nil {
		return nil, err
	}
	return services, nil
}

// ServiceInstances .
func (r registry) ServiceInstances(service string) (*InstanceList, error) {
	instances := &InstanceList{}
	err := r.client.GET(servicesPath+"/"+service, r.debug, nil, instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

// ValidateRegistryURL .
func ValidateRegistryURL(ctx *cli.Context) (string, error) {
	url := ctx.GlobalString(common.RegistryURL.Flag())
	if len(url) == 0 {
		return "empty", common.ErrRegistryURLNotFound
	}
	if !utils.IsURL(url) {
		return url, common.ErrRegistryURLInvalid
	}
	return url, nil
}
