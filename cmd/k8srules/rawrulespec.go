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

package main

import (
	"encoding/json"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/meta"
	"k8s.io/client-go/pkg/api/unversioned"

	"github.com/amalgam8/amalgam8/pkg/adapters/rules/kubernetes"
)

// RoutingRule defines a generic rule, having arbitrary Spec (json.RawMessage).
// This delays marshaling errors from k8s watchers in the k8s client code to our code
type RoutingRule struct {
	unversioned.TypeMeta `json:",inline"`
	Metadata             api.ObjectMeta `json:"metadata"`

	Spec   json.RawMessage       `json:"spec,omitempty"`
	Status kubernetes.StatusSpec `json:"status,omitempty"`
}

// RoutingRuleList is a list of generic rules
type RoutingRuleList struct {
	unversioned.TypeMeta `json:",inline"`
	Metadata             unversioned.ListMeta `json:"metadata"`

	Items []RoutingRule `json:"items"`
}

// GetObjectKind implement k8s Object
func (r *RoutingRule) GetObjectKind() unversioned.ObjectKind {
	return &r.TypeMeta
}

// GetObjectMeta implement k8s ObjectMetaAccessor
func (r *RoutingRule) GetObjectMeta() meta.Object {
	return &r.Metadata
}

// GetObjectKind implement k8s Object
func (rl *RoutingRuleList) GetObjectKind() unversioned.ObjectKind {
	return &rl.TypeMeta
}

// GetListMeta implement k8s ListMetaAccessor
func (rl *RoutingRuleList) GetListMeta() unversioned.List {
	return &rl.Metadata
}

// delayed spec unmarshalling - convert from the generic Spec rule to a routing rule
func (r *RoutingRule) reify() (*kubernetes.RoutingRule, error) {
	k8srule := &kubernetes.RoutingRule{
		TypeMeta: r.TypeMeta,
		Metadata: r.Metadata,
		Status:   r.Status,
	}
	err := json.Unmarshal(r.Spec, &k8srule.Spec)
	return k8srule, err

}
