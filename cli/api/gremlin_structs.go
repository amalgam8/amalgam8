package api

import "encoding/json"

const (
	recipesPath = "/api/v1/recipes"
)

// RecipeRun .
type RecipeRun struct {
	Topology  json.RawMessage `json:"topology" yaml:"topology"`
	Scenarios json.RawMessage `json:"scenarios" yaml:"scenarios"`
	Header    string          `json:"header" yaml:"header"`
	Pattern   string          `json:"header_pattern" yaml:"header_pattern"`
}

// RecipeChecks .
type RecipeChecks struct {
	Checklist json.RawMessage `json:"checklist"`
}

// RecipeResults .
type RecipeResults struct {
	Results []map[string]interface{} `json:"results"`
}
