package api

// RulesStore for the controller.
type RulesStore interface {
	// GetRules returns the rules for this namespace that match the filter.
	ListRules(f *Filter) (*RuleSet, error)
}
