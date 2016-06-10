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

package store

import (
	"bytes"
)

// Endpoint represents a network endpoint.
// Immutable by convention.
type Endpoint struct {
	Type  string
	Value string
}

func (e *Endpoint) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(e.Value)
	return buffer.String()
}

// DeepClone creates a deep copy of the receiver
func (e *Endpoint) DeepClone() *Endpoint {
	cloned := *e
	return &cloned
}
