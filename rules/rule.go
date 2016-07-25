package rules

import "encoding/json"

type Rule struct {
	Name     string   `json:"name"`
	ID       string   `json:"id"`
	Selector Selector `json:"selector"`
	Action   Action   `json:"action"`
}

// Selector &&
type Selector struct {
	Source      string `json:"source,omitempty"`
	Destination string `json:"destination,omitempty"`
	Header      string `json:"header,omitempty"`
	CookieValue string `json:"cookie_value,omitempty"`
}

type Action struct {
	Operation  string          `json:"op"`
	Parameters json.RawMessage `json:"parameters"`
}



