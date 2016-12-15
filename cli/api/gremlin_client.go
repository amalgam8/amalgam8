package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	api "github.com/amalgam8/amalgam8/cli/client"
	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/urfave/cli"
)

type gremlin struct {
	ctx    *cli.Context
	client api.Client
	debug  bool
}

// GremlinClient .
type GremlinClient interface {
	SetRecipes(topology, scenarios io.Reader, header, pattern string) (string, error)
	RecipeResults(id string, checks io.Reader) (*RecipeResults, error)
	DeleteRecipe(id string) (interface{}, error)
}

// NewGremlinClient .
func NewGremlinClient(ctx *cli.Context) (GremlinClient, error) {
	u, err := ValidateGremlinURL(ctx)
	if err != nil {
		fmt.Fprintf(ctx.App.Writer, fmt.Sprintf("%s: %q\n\n", err.Error(), u))
		return nil, err
	}

	// Read Token
	token := ctx.GlobalString(common.GremlinToken.Flag())

	// Check if a custom client has been set
	var client *http.Client
	if c, ok := ctx.App.Metadata["httpClient"].(*http.Client); ok {
		client = c
	}

	return &gremlin{
		debug:  ctx.GlobalBool(common.Debug.Flag()),
		client: api.NewClient(u, token, client),
	}, nil
}

// SetRecipes .
func (g *gremlin) SetRecipes(topology, scenarios io.Reader, header, pattern string) (string, error) {

	if header == "" {
		header = "X-Request-ID"
	}

	if pattern == "" {
		pattern = "*"
	}

	topologyBuf := new(bytes.Buffer)
	_, err := topologyBuf.ReadFrom(topology)
	if err != nil {
		return "", err
	}

	scenariosBuf := new(bytes.Buffer)
	_, err = scenariosBuf.ReadFrom(scenarios)
	if err != nil {
		return "", err
	}

	// Remove unnecesary space characters from topology
	topologyCompact := new(bytes.Buffer)
	err = json.Compact(topologyCompact, topologyBuf.Bytes())
	if err != nil {
		return "", err
	}

	// Remove unnecesary space characters from scenarios
	scenariosCompact := new(bytes.Buffer)
	err = json.Compact(scenariosCompact, scenariosBuf.Bytes())
	if err != nil {
		return "", err
	}

	recipe := RecipeRun{
		Topology:  topologyCompact.Bytes(),
		Scenarios: scenariosCompact.Bytes(),
		Header:    header,
		Pattern:   pattern,
	}

	// Create JSON recipe
	data, err := json.Marshal(&recipe)
	if err != nil {
		return "", err
	}

	payload := bytes.NewReader(data)

	headers := g.client.NewHeader()
	headers.Add("Accept", "application/json")
	result := &struct {
		ID string `json:"recipe_id"`
	}{}

	err = g.client.POST(recipesPath, payload, g.debug, headers, result)
	if err != nil {
		return "", err
	}

	return result.ID, nil
}

// RecipeResults .
func (g *gremlin) RecipeResults(id string, checks io.Reader) (*RecipeResults, error) {

	if id == "" {
		return nil, fmt.Errorf("ID can not be empty")
	}

	checksBuf := new(bytes.Buffer)
	_, err := checksBuf.ReadFrom(checks)
	if err != nil {
		return nil, err
	}

	// Remove unnecesary space characters from checks
	checksCompact := new(bytes.Buffer)
	err = json.Compact(checksCompact, checksBuf.Bytes())
	if err != nil {
		return nil, err
	}

	recipeChecks := RecipeChecks{
		Checklist: checksCompact.Bytes(),
	}

	// Create JSON recipe
	data, err := json.Marshal(&recipeChecks)
	if err != nil {
		return nil, err
	}

	payload := bytes.NewReader(data)

	headers := g.client.NewHeader()
	headers.Add("Accept", "application/json")

	results := &RecipeResults{}
	err = g.client.POST(recipesPath+"/"+id, payload, g.debug, headers, results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// DeleteRecipe .
func (g *gremlin) DeleteRecipe(id string) (interface{}, error) {

	if id == "" {
		return nil, fmt.Errorf("ID can not be empty")
	}

	var result string
	err := g.client.DELETE(recipesPath+"/"+id, g.debug, nil, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ValidateGremlinURL .
func ValidateGremlinURL(ctx *cli.Context) (string, error) {
	u := ctx.GlobalString(common.GremlinURL.Flag())
	if len(u) == 0 {
		return common.Empty, common.ErrGremlinURLNotFound
	}
	_, err := url.ParseRequestURI(u)
	if err != nil {
		return u, common.ErrGremlinURLInvalid
	}
	return u, nil
}
