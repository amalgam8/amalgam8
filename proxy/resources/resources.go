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

package resources

import "time"

// MetaData service instance metadata
type MetaData struct {
	Version string
}

// ServiceCatalog TODO
type ServiceCatalog struct {
	Services   []Service
	LastUpdate time.Time
}

// Service TODO
type Service struct {
	Name      string
	Endpoints []Endpoint
}

// Endpoint TODO
type Endpoint struct {
	Type     string
	Value    string
	Metadata MetaData
}

// ByService TODO
type ByService []Service

// Len TODO
func (a ByService) Len() int {
	return len(a)
}

// Swap TODO
func (a ByService) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Less TODO
func (a ByService) Less(i, j int) bool {
	return a[i].Name < a[j].Name
}

// ByEndpoint TODO
type ByEndpoint []Endpoint

// Len TODO
func (a ByEndpoint) Len() int {
	return len(a)
}

// Swap TODO
func (a ByEndpoint) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Less TODO
func (a ByEndpoint) Less(i, j int) bool {
	if a[i].Value == a[j].Value {
		if a[i].Type == a[j].Type {
			return a[i].Metadata.Version < a[j].Metadata.Version
		}
		return a[i].Type < a[j].Type
	}
	return a[i].Value < a[j].Value
}
