// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package rules

import (
	"fmt"
)

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

// Empty returns whether the filter has any attributes that would cause rules to be filtered out. A filter is considered
// empty if no rules would be filtered out from any set of rules.
func (f Filter) Empty() bool {
	return len(f.IDs) == 0 && len(f.Tags) == 0 && len(f.Destinations) == 0 && f.RuleType == RuleAny
}

// String representation of the filter
func (f Filter) String() string {
	return fmt.Sprintf("%#v", f)
}

// FilterRules returns a list of filtered rules.
func FilterRules(f Filter, rules []Rule) []Rule {
	// Before we begin filtering we build sets of each filter field to avoid N^2 lookups.

	// Build set of acceptable IDs
	var ids map[string]bool
	if len(f.IDs) > 0 {
		ids = make(map[string]bool)
		for _, id := range f.IDs {
			ids[id] = true
		}
	}

	// Build set of acceptable destinations
	var dests map[string]bool
	if len(f.Destinations) > 0 {
		dests = make(map[string]bool)
		for _, d := range f.Destinations {
			dests[d] = true
		}
	}

	// Build set of tags the rule must contain
	var tags map[string]bool
	if len(f.Tags) > 0 {
		tags = make(map[string]bool)
		for _, t := range f.Tags {
			tags[t] = true
		}
	}

	// Iterate through the rules, building a new list of rules that pass the filter
	res := make([]Rule, 0, len(rules)) // Filtered rules
	for _, rule := range rules {
		// Filter by ID. The ID must be a member of the set of acceptable IDs.
		if ids != nil {
			if _, exists := ids[rule.ID]; !exists {
				continue
			}
		}

		// Filter by rule type
		if (f.RuleType == RuleAction && len(rule.Actions) == 0) ||
			(f.RuleType == RuleRoute && len(rule.Route) == 0) {
			continue
		}

		// Filter by destination. If the destination is not in the set
		// of accepted destinations, filter the rule out.
		if dests != nil {
			if _, exists := dests[rule.Destination]; !exists {
				continue
			}
		}

		// Ensure rule's tags are a subset of the tags specified by the filter.
		if tags != nil {
			c := 0
			for _, t := range rule.Tags {
				if _, exists := tags[t]; exists {
					c++
				}
			}

			if c != len(f.Tags) {
				continue
			}
		}

		// The rule has passed all the filters, so we add it to the list of filtered rules.
		res = append(res, rule)
	}

	return res
}
