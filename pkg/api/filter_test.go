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

package api

import (
	"reflect"
	"testing"
)

func TestFilterRules(t *testing.T) {
	rule1 := Rule{
		ID:          "id1",
		Priority:    1,
		Tags:        []string{"tag1", "tag2"},
		Destination: "service1",
		Match:       nil,
		Actions:     []Action{},
	}
	rule2 := Rule{
		ID:          "id1",
		Priority:    1,
		Tags:        []string{},
		Destination: "service2",
		Match:       nil,
		Route:       &Route{},
	}

	cases := []struct {
		In, Out []Rule
		Filter  RuleFilter
	}{
		{ // Empty filter does not do anything
			In: []Rule{
				rule1,
			},
			Out: []Rule{
				rule1,
			},
			Filter: RuleFilter{},
		},
		{ // Filter by tags
			In: []Rule{
				rule1,
				rule2,
			},
			Out: []Rule{
				rule1,
			},
			Filter: RuleFilter{
				Tags: []string{"tag1"},
			},
		},
		{ // Filter by destination
			In: []Rule{
				rule1,
				rule2,
			},
			Out: []Rule{
				rule2,
			},
			Filter: RuleFilter{
				Destinations: []string{"service2"},
			},
		},
		{ // Filter action rules
			In: []Rule{
				rule1,
				rule2,
			},
			Out: []Rule{
				rule1,
			},
			Filter: RuleFilter{
				RuleType: RuleAction,
			},
		},
		{ // Filter route rules
			In: []Rule{
				rule1,
				rule2,
			},
			Out: []Rule{
				rule2,
			},
			Filter: RuleFilter{
				RuleType: RuleRoute,
			},
		},
	}
	for _, c := range cases {
		actual := c.Filter.Apply(c.In)
		if !reflect.DeepEqual(actual, c.Out) {
			t.Errorf("FilterRules(%v, %v): expected %v, got %v", c.Filter, c.In, c.Out, actual)
		}
	}
}
