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

package protocol

//Type represents the API protocol type
type Type uint32

//Defines the API protocol types
const (
	Amalgam8 Type = 1 << iota // Amalgam8 protocol
	Eureka                    // Eureka protocol
)

// NameOf returns the name of the given protocol type value
func NameOf(protocol Type) string {
	switch protocol {
	case Amalgam8:
		return "Amalgam8"
	case Eureka:
		return "Eureka"
	default:
		return "Unknown"
	}
}
