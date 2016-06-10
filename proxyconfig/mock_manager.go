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

package proxyconfig

import "github.com/amalgam8/controller/resources"

// MockManager mocks interface
type MockManager struct {
	SetError    error
	GetVal      resources.ProxyConfig
	GetError    error
	DeleteError error
}

// Set mocks method
func (m *MockManager) Set(rules resources.ProxyConfig) error {
	return m.SetError
}

// Get mocks method
func (m *MockManager) Get(id string) (resources.ProxyConfig, error) {
	return m.GetVal, m.GetError
}

// Delete mocks method
func (m *MockManager) Delete(id string) error {
	return m.DeleteError
}
