package api

import (
	"fmt"
	api "github.com/amalgam8/amalgam8/cli/client"
	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/utils"
	"github.com/urfave/cli"
	"io"
	"net/http"
)

type controller struct {
	ctx    *cli.Context
	client api.Client
	debug  bool
}

// ControllerClient .
type ControllerClient interface {
	Routes() (*RouteList, error)
	GetActions() (*ActionList, error)
	Rules(uri string) (*RuleList, error)
	SetRules(payload io.Reader) (interface{}, error)
	DeleteRules(uri string) (interface{}, error)
}

// NewControllerClient .
func NewControllerClient(ctx *cli.Context) (ControllerClient, error) {
	url, err := ValidateControllerURL(ctx)
	if err != nil {
		fmt.Fprintf(ctx.App.Writer, fmt.Sprintf("\n%s: %q\n\n", err.Error(), url))
		return nil, err
	}

	// Read Token
	token := ctx.GlobalString(common.ControllerToken.Flag())

	// Check if a custom client has been set
	var client *http.Client
	if c, ok := ctx.App.Metadata["httpClient"].(*http.Client); ok {
		client = c
	}

	return &controller{
		debug:  ctx.GlobalBool(common.Debug.Flag()),
		client: api.NewClient(url, token, client),
	}, nil
}

// Rules .
func (c *controller) Rules(uri string) (*RuleList, error) {
	rules := &RuleList{}
	err := c.client.GET(rulesPath+uri, c.debug, nil, rules)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

// SetRules .
func (c *controller) SetRules(payload io.Reader) (interface{}, error) {
	headers := c.client.NewHeader()
	headers.Add("Accept", "application/json")
	result := &struct {
		IDs []string `json:"ids"`
	}{}
	err := c.client.POST(rulesPath, payload, c.debug, headers, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteRules .
func (c *controller) DeleteRules(uri string) (interface{}, error) {
	var result string
	err := c.client.DELETE(rulesPath+uri, c.debug, nil, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Routes .
func (c *controller) Routes() (*RouteList, error) {
	routes := &RouteList{}
	err := c.client.GET(routesPath, c.debug, nil, routes)
	if err != nil {
		return nil, err
	}
	return routes, nil
}

// GetActions.
func (c *controller) GetActions() (*ActionList, error) {
	actions := &ActionList{}
	err := c.client.GET(actionPath, c.debug, nil, actions)
	if err != nil {
		return nil, err
	}
	return actions, nil
}

// ValidateControllerURL .
func ValidateControllerURL(ctx *cli.Context) (string, error) {
	url := ctx.GlobalString(common.ControllerURL.Flag())
	if len(url) == 0 {
		return "empty", common.ErrControllerURLNotFound
	}
	if !utils.IsURL(url) {
		return url, common.ErrControllerURLInvalid
	}
	return url, nil
}
