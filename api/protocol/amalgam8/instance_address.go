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

import (
	"fmt"
)

const (
	//EndpointTypeTCP Denotes a TCP endpoint type
	EndpointTypeTCP = "tcp"
	//EndpointTypeUDP Denotes a UDP endpoint type
	EndpointTypeUDP = "udp"
	//EndpointTypeHTTP Denotes an HTTP endpoint type
	EndpointTypeHTTP = "http"
	//EndpointTypeHTTPS Denotes an HTTPS endpoint type
	EndpointTypeHTTPS = "https"
	//EndpointTypeUser Denotes a user-defined endpoint type
	EndpointTypeUser = "user"
)

// InstanceAddress encapsulates a service network endpoint
type InstanceAddress struct {
	Type  string `json:"type,omitempty"` // possible values: { tcp, udp, http, https, user}
	Value string `json:"value"`          // can't be empty string, or consists of only spaces

}

// String output the structure
func (a *InstanceAddress) String() string {
	return fmt.Sprintf("%s:%s", a.Type, a.Value)
}
