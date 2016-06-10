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

package notification

import "github.com/amalgam8/controller/resources"

// MockTenantProducerCache mocks interface
type MockTenantProducerCache struct {
	SendEventError error
}

// StartGC mocks method
func (m *MockTenantProducerCache) StartGC() {}

// SendEvent mocks method
func (m *MockTenantProducerCache) SendEvent(tenantID string, kafka resources.Kafka) error {
	return m.SendEventError
}

// Delete mocks method
func (m *MockTenantProducerCache) Delete(tenantID string) {}
