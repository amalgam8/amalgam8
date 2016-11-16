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

package uptime

// URL returns URL path used for uptime checks
func URL() string {
	return uptimePath
}

// HealthyURL returns URL path used for healthy checks
func HealthyURL() string {
	return healthyPath
}

const (
	uptimePath  = "/uptime"
	healthyPath = "/"
)
