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

type Manager interface {
	AddRules(namespace string, rules []Rule) (NewRules, error)
	GetRules(namespace string, filter Filter) (RetrievedRules, error)
	UpdateRules(namespace string, rules []Rule) error
	DeleteRules(namespace string, filter Filter) error
	SetRules(namespace string, filter Filter, rules []Rule) (NewRules, error)
}

type NewRules struct {
	IDs []string
}

type RetrievedRules struct {
	// Rules filtered
	Rules []Rule

	// Revision of the rules for this namespace. Each time the rules for this namespace are changed, this the
	// revision is incremented.
	// FIXME: if a Redis DB is updated once a millisecond the revision will roll over after 292,471,208.678 years.
	Revision int64
}
