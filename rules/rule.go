package rules

import "encoding/json"

type Rule struct {
	ID     string          `json:"id"`
	Tags   []string        `json:"tags"`
	Match  json.RawMessage `json:"match"`
	Action json.RawMessage `json:"action"`
}
