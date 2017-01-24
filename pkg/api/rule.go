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

import "encoding/json"

// Rule represents an individual rule.
type Rule struct {
	ID          string   `json:"id"`
	Priority    int      `json:"priority"`
	Tags        []string `json:"tags,omitempty"`
	Destination string   `json:"destination"`
	Match       *Match   `json:"match,omitempty"`
	Route       *Route   `json:"route,omitempty"`
	Actions     []Action `json:"actions,omitempty"`
}

// Source definition.
type Source struct {
	Name string   `json:"name"`
	Tags []string `json:"tags,omitempty"`
}

// Match definition
type Match struct {
	Source  *Source           `json:"source,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Route definition
type Route struct {
	Backends []Backend `json:"backends"`
}

// URI for backends.
type URI struct {
	Path          string `json:"path"`
	Prefix        string `json:"prefix"`
	PrefixRewrite string `json:"prefix_rewrite"`
}

// Backend represents a backend to route to.
type Backend struct {
	Name    string   `json:"name,omitempty"`
	Tags    []string `json:"tags"`
	URI     *URI     `json:"uri,omitempty"`
	Weight  float64  `json:"weight,omitempty"`
	Timeout float64  `json:"timeout,omitempty"`
	Retries int      `json:"retries,omitempty"`
}

// Action to take.
type Action struct {
	internal   interface{}
	actionType string
}

// MarshalJSON implement the json marshal function
func (a *Action) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.internal)
}

// UnmarshalJSON implement the json unmarshal function
func (a *Action) UnmarshalJSON(data []byte) error {
	action := struct {
		Type string `json:"action"`
	}{}
	err := json.Unmarshal(data, &action)
	if err != nil {
		return err
	}

	a.actionType = action.Type

	switch action.Type {
	case "delay":
		delay := DelayAction{}
		if err = json.Unmarshal(data, &delay); err != nil {
			return err
		}
		a.internal = delay
	case "abort":
		abort := AbortAction{}
		if err = json.Unmarshal(data, &abort); err != nil {
			return err
		}
		a.internal = abort
	case "trace":
		trace := TraceAction{}
		if err = json.Unmarshal(data, &trace); err != nil {
			return err
		}
		a.internal = trace
	}
	return nil
}

// GetType returns action type
func (a *Action) GetType() string {
	return a.actionType
}

// Internal returns action type interface
func (a *Action) Internal() interface{} {
	return a.internal
}

// DelayAction definition
type DelayAction struct {
	Action      string   `json:"action"`
	Probability float64  `json:"probability"`
	Tags        []string `json:"tags"`
	Duration    float64  `json:"duration"`
}

// AbortAction definition
type AbortAction struct {
	Action      string   `json:"action"`
	Probability float64  `json:"probability"`
	Tags        []string `json:"tags"`
	ReturnCode  int      `json:"return_code"`
}

// TraceAction definition
type TraceAction struct {
	Action   string   `json:"action"`
	Tags     []string `json:"tags"`
	LogKey   string   `json:"log_key"`
	LogValue string   `json:"log_value"`
}
