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

package clients

import "time"

// MockController mocks the Controller interface
type MockController struct {
	RegisterError error
	ConfigError   error
	ConfigString  string
	GetCredsError error
	GetCredsVal   TenantCredentials
}

// Register mocks interface
func (m *MockController) Register() error {
	return m.RegisterError
}

// GetNGINXConfig mocks interface
func (m *MockController) GetNGINXConfig(version *time.Time) (string, error) {
	return m.ConfigString, m.ConfigError
}

// GetCredentials mocks interface
func (m *MockController) GetCredentials() (TenantCredentials, error) {
	return m.GetCredsVal, m.GetCredsError
}
