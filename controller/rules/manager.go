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

import "github.com/amalgam8/amalgam8/pkg/api"

// Manager is an interface for managing collections of rules mapped by namespace.
type Manager interface {
	// AddRules validates the rules and adds them to the collection for the namespace.
	AddRules(namespace string, rules []api.Rule) (NewRules, error)

	// GetRules returns a collection of filtered rules from the namespace.
	GetRules(namespace string, filter api.RuleFilter) (RetrievedRules, error)

	// UpdateRules updates rules by ID in the namespace.
	UpdateRules(namespace string, rules []api.Rule) error

	// DeleteRules deletes rules that match the filter in the namespace.
	DeleteRules(namespace string, filter api.RuleFilter) error

	// SetRules deletes the rules that match the filter and adds the new rules as a single
	// atomic transaction.
	SetRules(namespace string, filter api.RuleFilter, rules []api.Rule) (NewRules, error)
}

// NewRules provides information about newly added rules.
type NewRules struct {
	// IDs of the added rules.
	IDs []string
}

// RetrievedRules are the results of a read from a manager.
type RetrievedRules struct {
	// Rules that passed the filter.
	Rules []api.Rule

	// Revision of the rules for this namespace. Each time the collection of rules for the namespace are changed
	// the revision is incremented.
	Revision int64
}
