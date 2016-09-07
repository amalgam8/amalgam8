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

package health

// Status is the result of a health check run.
type Status struct {
	Healthy    bool                   `json:"healthy"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

var (
	// Healthy is a healthy status with no additional properties.
	Healthy = Status{Healthy: true}
)

// StatusHealthy creates a new healthy status with given message property.
// To return a default healthy status, with no properties, just use health.Healthy
func StatusHealthy(message string) Status {
	s := Status{
		Healthy: true,
	}
	if len(message) > 0 {
		s.Properties = map[string]interface{}{"message": message}
	}
	return s
}

// StatusHealthyWithProperties creates a new healthy status with given properties.
func StatusHealthyWithProperties(properties map[string]interface{}) Status {
	return Status{
		Healthy:    true,
		Properties: properties,
	}
}

// StatusUnhealthy creates a new unhealthy status with the given message and error properties.
func StatusUnhealthy(message string, cause error) Status {
	s := Status{
		Healthy: false,
	}
	if len(message) > 0 {
		s.Properties = map[string]interface{}{"message": message}
	}
	if cause != nil && len(cause.Error()) > 0 {
		if s.Properties == nil {
			s.Properties = make(map[string]interface{})
		}
		s.Properties["cause"] = cause.Error()
	}
	return s
}

// StatusUnhealthyWithProperties creates a new unhealthy status with given properties.
func StatusUnhealthyWithProperties(properties map[string]interface{}) Status {
	return Status{
		Healthy:    false,
		Properties: properties,
	}
}
