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

package checker

import "github.com/amalgam8/controller/resources"

// MockChecker mocks interface
type MockChecker struct {
	RegisterError error

	DeregisterError error

	CheckError error

	GetVal   resources.ServiceCatalog
	GetError error
}

// Register mocks method
func (m *MockChecker) Register(id string) error {
	return m.RegisterError
}

// Deregister mocks method
func (m *MockChecker) Deregister(id string) error {
	return m.DeregisterError
}

// Check mocks method
func (m *MockChecker) Check(ids []string) error {
	return m.CheckError
}

// Get mocks method
func (m *MockChecker) Get(id string) (resources.ServiceCatalog, error) {
	return m.GetVal, m.GetError
}
