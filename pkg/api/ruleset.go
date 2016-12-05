package api

// RulesSet is the information returned from a rule query.
type RuleSet struct {
	// Rules that matched the filter.
	Rules []*Rule `json:"rules"`

	// Revision of the rules for this namespace.
	Revision int64 `json:"revision"`
}