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

import "fmt"

// InvalidRuleError occurs when a rule is not valid
type InvalidRuleError struct{}

// Error description
func (e *InvalidRuleError) Error() string {
	return "Invalid Rule Error"
}

// RedisInsertError occurs when there is an issue writing to Redis
type RedisInsertError struct{}

// Error description
func (e *RedisInsertError) Error() string {
	return "Redis insert failed"
}

// JSONMarshalError describes a JSON marshal or unmarshaling error
type JSONMarshalError struct {
	Message string
}

// Error description
func (e *JSONMarshalError) Error() string {
	return fmt.Sprintf("Error marshaling JSON: %v", e.Message)
}
