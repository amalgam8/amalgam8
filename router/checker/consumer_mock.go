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

// MockConsumer mocks interface
type MockConsumer struct {
	CloseError        error
	ReceiveEventError error
	ReceiveEventKey   string
	ReceiveEventValue string
}

// ReceiveEvent mocks method
func (c *MockConsumer) ReceiveEvent() (string, string, error) {
	return c.ReceiveEventKey, c.ReceiveEventValue, c.ReceiveEventError
}

// Close mocks method
func (c *MockConsumer) Close() error {
	return c.CloseError
}
