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

package amalgam8

import "strings"

// InstanceLinks type defines the REST links relating to a service instance
type InstanceLinks struct {
	Self      string `json:"self,omitempty"`
	Heartbeat string `json:"heartbeat,omitempty"`
}

// BuildLinks composes URL values in the InstanceLinks structure for the instance identifier.
// URL's are based off of the given base URL value.
func BuildLinks(baseURL, id string) *InstanceLinks {
	return &InstanceLinks{
		Self:      strings.Join([]string{baseURL, InstanceURL(id)}, ""),
		Heartbeat: strings.Join([]string{baseURL, InstanceHeartbeatURL(id)}, ""),
	}
}
