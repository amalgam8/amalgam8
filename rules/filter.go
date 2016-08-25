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

type Filter struct {
	IDs          []string
	Tags         []string
	Destinations []string
	RuleType     int
}

func (f Filter) String() string {
	return fmt.Sprintf("filter: IDs=%v Tags=%v Destinations=%v RuleType=%v", f.IDs, f.Tags, f.Destinations, f.RuleType)
}

// FilterRules returns a list of filtered rules.
func FilterRules(f Filter, rules []Rule) []Rule {
	// Set of acceptable destinations
	var dests map[string]bool
	if len(f.Destinations) > 0 {
		dests = make(map[string]bool)
		for _, d := range f.Destinations {
			dests[d] = true
		}
	}

	// Set of tags the rule must contain
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
