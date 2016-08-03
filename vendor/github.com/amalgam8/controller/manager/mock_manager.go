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

package manager

import "github.com/amalgam8/controller/resources"

// MockManager mocks interface
type MockManager struct {
	SetError    error
	GetVal      resources.TenantEntry
	GetError    error
	DeleteError error
}

// Delete mocks method
func (m *MockManager) Delete(id string) error {
	return m.DeleteError
}

// Create mocks method
func (m *MockManager) Create(id, token string, rules resources.TenantInfo) error {
	return nil
}

// Set mocks method
func (m *MockManager) Set(id string, rules resources.TenantInfo) error {
	return m.SetError
}

// Get mocks method
func (m *MockManager) Get(id string) (resources.TenantEntry, error) {
	return m.GetVal, m.GetError
}

// SetVersion mocks method
func (m *MockManager) SetVersion(id string, version resources.Version) error {
	return nil
}

// DeleteVersion mocks method
func (m *MockManager) DeleteVersion(id, service string) error {
	return nil
}

// GetVersion mocks method
func (m *MockManager) GetVersion(id, service string) (resources.Version, error) {
	return resources.Version{}, nil
}
