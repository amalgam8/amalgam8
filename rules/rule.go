package rules

import "encoding/json"

type Rule struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Match  json.RawMessage `json:"match"`
	Action json.RawMessage `json:"action"`
}

