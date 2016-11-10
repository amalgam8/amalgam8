package api

import ()

const (
	rulesPath  = "/v1/rules"
	routesPath = "/v1/rules/routes"
	actionPath = "/v1/rules/actions"
)

// RuleList .
type RuleList struct {
	Rules []Rule `json:"rules" yaml:"rules"`
}

// Rule represents an individual rule.
// TODO: use json.rawmessage for some structs???
type Rule struct {
	ID          string        `json:"id" yaml:"id"`
	Priority    int           `json:"priority,omitempty" yaml:"priority,omitempty"`
	Tags        []string      `json:"tags,omitempty" yaml:"tags,omitempty"`
	Destination string        `json:"destination,omitempty" yaml:"destination,omitempty"`
	Match       *MatchRules   `json:"match,omitempty" yaml:"match,omitempty"`
	Route       *RouteRules   `json:"route,omitempty" yaml:"route,omitempty"`
	Actions     *ActionsRules `json:"actions,omitempty" yaml:"actions,omitempty"`
}

// MatchRules .
type MatchRules struct {
	Source  *source           `json:"source,omitempty" yaml:"source,omitempty"`
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	All     []struct {
		Source  source            `json:"source,omitempty" yaml:"source,omitempty"`
		Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	} `json:"all,omitempty" yaml:"all,omitempty"`
	Any []struct {
		Source  source            `json:"source,omitempty" yaml:"source,omitempty"`
		Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	} `json:"any,omitempty" yaml:"any,omitempty"`
	None []struct {
		Source  source            `json:"source,omitempty" yaml:"source,omitempty"`
		Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	} `json:"none,omitempty" yaml:"none,omitempty"`
}

// RouteRules .
type RouteRules struct {
	Backends []struct {
		Name    string   `json:"name,omitempty" yaml:"name,omitempty"`
		Timeout string   `json:"timeout,omitempty" yaml:"timeout,omitempty"`
		Tags    []string `json:"tags,omitempty" yaml:"tags,omitempty"`
		Weight  float32  `json:"weight,omitempty" yaml:"weight,omitempty"`
	} `json:"backends,omitempty" yaml:"backends,omitempty"`
}

// ActionsRules .
type ActionsRules []struct {
	Action      string   `json:"action,omitempty" yaml:"action,omitempty"`
	Duration    float32  `json:"duration,omitempty" yaml:"duration,omitempty"`
	Probability float32  `json:"probability,omitempty" yaml:"probability,omitempty"`
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	ReturnCode  int      `json:"return_code,omitempty" yaml:"return_code,omitempty"`
	LogKey      string   `json:"log_key,omitempty" yaml:"log_key,omitempty"`
	LogValue    string   `json:"log_value,omitempty" yaml:"log_value,omitempty"`
}

type source struct {
	Name string   `json:"name,omitempty"`
	Tags []string `json:"tags,omitempty"`
}

// RouteList .
type RouteList struct {
	ServiceRoutes map[string][]Route `json:"services"`
}

// Route .
type Route struct {
	ID          string      `json:"id"`
	Priority    int         `json:"priority"`
	Destination string      `json:"destination"`
	Match       *MatchRules `json:"match"`
	Routes      *RouteRules `json:"route"`
}

// ActionList .
type ActionList struct {
	ServiceActions map[string][]Action `json:"services"`
}

// Action .
type Action struct {
	ID          string        `json:"id"`
	Priority    int           `json:"priority"`
	Destination string        `json:"destination"`
	Match       *MatchRules   `json:"match"`
	Actions     *ActionsRules `json:"actions"`
}
