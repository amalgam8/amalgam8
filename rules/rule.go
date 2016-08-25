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

import "encoding/json"

type Rule struct {
	ID          string          `json:"id"`
	Priority    int             `json:"priority"`
	Tags        []string        `json:"tags,omitempty"`
	Destination string          `json:"destination"`
	Match       json.RawMessage `json:"match,omitempty"`
	Route       json.RawMessage `json:"route,omitempty"`
	Actions     json.RawMessage `json:"actions,omitempty"`
}
