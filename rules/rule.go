package rules

import "encoding/json"

type Rule struct {
	ID          string          `json:"id"`
	Priority    int             `json:"priority"`
	Tags        []string        `json:"tags,omitempty"`
	Destination string          `json:"destination"`
	Match       json.RawMessage `json:"match"`
	Route       json.RawMessage `json:"route,omitempty"`
	Actions     json.RawMessage `json:"actions,omitempty"`
}
