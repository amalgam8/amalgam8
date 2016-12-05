package api

const (
	// RuleAny denotes any rule type
	RuleAny = iota

	// RuleRoute denotes rules with a routing field
	RuleRoute

	// RuleAction denotes rules with an action field
	RuleAction
)

// Filter to apply to sets of rules.
type Filter struct {
	// IDs is the set of acceptable rule IDs. A rule will pass the filter if its ID is a member of this set.
	// This field is ignored when len(IDs) <= 0.
	IDs []string

	// Tags is the set of tags each rule must contain. A rule will pass the filter if its tags are a subset of this
	// set. This field is ignored when len(Tags) <= 0.
	Tags []string

	// Destinations is the set of acceptable rule destinations. A rule will pass the filter if its destination is
	// a member of this set. This field is ignored when len(Destinations) <= 0.
	Destinations []string

	// RuleType is the type of rule to filter by.
	RuleType int
}