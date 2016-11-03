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

package healthcheck

import (
	"time"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCheck struct {
	mock.Mock
}

func (m *MockCheck) Execute() error {
	args := m.Called()
	return args.Error(0)
}

func TestAgentDefaultConfig(t *testing.T) {
	check := &MockCheck{}
	agnt := NewAgent(check, 0)
	assert.NotNil(t, agnt)
	assert.Equal(t, agnt.(*agent).interval, defaultHealthCheckInterval)
}

func TestAgent(t *testing.T) {
	hc := &MockCheck{}
	hc.On("Execute").Return(nil) // No execute errors.

	expectedCalls := 10                // Number of times that the check's execute function should be called.
	interval := 100 * time.Millisecond // Amount of time between calls.

	agent := NewAgent(hc, interval)
	assert.NotNil(t, agent)

	statusChan := make(chan error)
	agent.Start(statusChan)

	expectedEnd := time.Now().Add(interval * time.Duration(expectedCalls-1))

	// Check the output
	count := 0
	for count < expectedCalls {
		select {
		case err := <-statusChan:
			assert.NoError(t, err)
			count++
		}
	}

	end := time.Now()
	assert.WithinDuration(t, expectedEnd, end, time.Millisecond*50)
	hc.AssertNumberOfCalls(t, "Execute", expectedCalls)

	// Make sure the agent stops calling execute in a timely fashion.
	agent.Stop()
	time.Sleep(2 * interval)
	hc.AssertNumberOfCalls(t, "Execute", expectedCalls)
}
