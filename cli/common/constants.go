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

package common

import (
	"fmt"
	"strings"
)

// Used for EnvVars or Flags
const (
	// ControllerURL .
	ControllerURL Const = "CONTROLLER_URL"
	// ControllerToken .
	ControllerToken Const = "CONTROLLER_TOKEN"
	// GremlinURL .
	GremlinURL Const = "GREMLIN_URL"
	// GremlinToken .
	GremlinToken Const = "GREMLIN_TOKEN"

	// Debug .
	Debug Const = "DEBUG"
)

// Global Configurations
const (
	// DefaultLanguage .
	DefaultLanguage = "en-US"
)

// App Metadata
const (
	// Terminal .
	Terminal = "term"
)

// Other
const (
	// Empty .
	Empty = "empty"
)

// Const .
type Const string

// Flag returns the flag representation of a given Const.
// For 'CONSTANT_STRING' it will return 'constant-string'
func (c Const) Flag() string {
	return strings.ToLower(fmt.Sprint(c))
}

// EnvVar returns the env var representation of a given Const.
// For 'CONSTANT_STRING' it will return 'A8_CONSTANT_STRING'
func (c Const) EnvVar() string {
	return "A8_" + strings.ToUpper(fmt.Sprint(c))
}
