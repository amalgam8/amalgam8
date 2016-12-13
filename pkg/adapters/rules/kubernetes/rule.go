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

package kubernetes

import (
	"encoding/json"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/meta"
	"k8s.io/client-go/pkg/api/unversioned"

	a8api "github.com/amalgam8/amalgam8/pkg/api"
)

const (
	// ResourceName defines the name of the third party resource in kubernetes
	ResourceName = "rule"
	// ResourceGroupName defines the group name of the third party resource in kubernetes
	ResourceGroupName = "amalgam8.io"
	// ResourceVersion defines the version of the third party resource in kubernetes
	ResourceVersion = "v1"
	// ResourceDescription defines the description of the third party resource in kubernetes
	ResourceDescription = "A specification of an Amalgam8 rule resource"
	// RuleStateValid defines the state of a valid rule resource
	RuleStateValid = "valid"
	// RuleStateInvalid defines the state of an invalid rule resource
	RuleStateInvalid = "invalid"
)

// StatusSpec defines third party resource status
type StatusSpec struct {
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

// Rule defines the third party resource spec
type Rule struct {
	unversioned.TypeMeta `json:",inline"`
	Metadata             api.ObjectMeta `json:"metadata"`

	Spec   a8api.Rule `json:"spec,omitempty"`
	Status StatusSpec `json:"status,omitempty"`
}

// RuleList defines list of rules
type RuleList struct {
	unversioned.TypeMeta `json:",inline"`
	Metadata             unversioned.ListMeta `json:"metadata"`

	Items []Rule `json:"items"`
}

// GetObjectKind - Required to satisfy Object interface
func (r *Rule) GetObjectKind() unversioned.ObjectKind {
	return &r.TypeMeta
}

// GetObjectMeta - Required to satisfy ObjectMetaAccessor interface
func (r *Rule) GetObjectMeta() meta.Object {
	return &r.Metadata
}

// GetObjectKind - Required to satisfy Object interface
func (rl *RuleList) GetObjectKind() unversioned.ObjectKind {
	return &rl.TypeMeta
}

// GetListMeta - Required to satisfy ListMetaAccessor interface
func (rl *RuleList) GetListMeta() unversioned.List {
	return &rl.Metadata
}

// The code below is used only to work around a known problem with third-party
// resources and ugorji. If/when these issues are resolved, the code below
// should no longer be required.

// RuleListCopy defines list of rules
type RuleListCopy RuleList

// RuleCopy defines a rule resource
type RuleCopy Rule

// UnmarshalJSON parses the JSON-encoded data and stores the result in the value pointed to by r
func (r *Rule) UnmarshalJSON(data []byte) error {
	tmp := RuleCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	tmp2 := Rule(tmp)
	*r = tmp2
	return nil
}

// UnmarshalJSON parses the JSON-encoded data and stores the result in the value pointed to by rl
func (rl *RuleList) UnmarshalJSON(data []byte) error {
	tmp := RuleListCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	tmp2 := RuleList(tmp)
	*rl = tmp2
	return nil
}
