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

package register

import (
	"sync"

	"github.com/amalgam8/amalgam8/sidecar/register/healthcheck"
)

// HealthChecker uses a collection of periodic health checks to manage registration. Registration is maintained while
// the state of the health checker is healthy. If the state changes to unhealthy, the registration agent is stopped.
// Similarly, if the state changes to healthy the registration agent is started again. The health checker is considered
// unhealthy if any health check's last reported status is unhealthy.
type HealthChecker struct {
	active       bool
	stop         chan struct{}
	agents       []*healthcheck.Agent
	mutex        sync.Mutex
	registration *RegistrationAgent
}

// NewHealthChecker instantiates a health checker.
func NewHealthChecker(registration *RegistrationAgent, checks []*healthcheck.Agent) *HealthChecker {
	if len(checks) == 0 {
		panic("No health checks provided")
	}

	return &HealthChecker{
		agents:       checks,
		registration: registration,
	}
}

// Start checking
func (checker *HealthChecker) Start() {
	checker.mutex.Lock()
	defer checker.mutex.Unlock()

	if checker.active {
		return
	}
	checker.active = true

	go checker.maintainRegistration()
}

// Stop checking
func (checker *HealthChecker) Stop() {
	checker.mutex.Lock()
	defer checker.mutex.Unlock()

	if !checker.active {
		return
	}

	checker.active = false

	checker.stop <- struct{}{}
}

// maintainRegistration
func (checker *HealthChecker) maintainRegistration() {
	// Receives a value whenever the status of a health check agent changes from healthy to unhealthy or vice versa.
	healthChan := make(chan bool, len(checker.agents))

	// Start agents and begin monitoring for changes in health status.
	statusChans := make([]chan error, len(checker.agents))
	for i, agent := range checker.agents {
		// Create a channel for receiving health check statuses from the agent.
		statusChans[i] = make(chan error, 1)

		// Start the agent.
		agent.Start(statusChans[i])

		// Begin monitoring the agent.
		go func(statusChan chan error) {
			wasHealthy := false
			for {
				select {
				case err, open := <-statusChan:
					// Check if the channel has been closed.
					if !open {
						return
					}

					// Report changes in the health check agent's status.
					healthy := err == nil
					if healthy != wasHealthy {
						healthChan <- healthy
					}
					wasHealthy = healthy
				}
			}
		}(statusChans[i])
	}

	// Monitor the number of agents reporting healthy statuses. Maintain registration if all agents are reporting
	// healthy statuses, and unregister if any agents become unhealthy.
	numHealthy := 0
	for {
		select {
		case healthy := <-healthChan:
			if healthy {
				numHealthy++
				if numHealthy == len(checker.agents) { // Overall state has become healthy.
					checker.registration.Start()
				}
			} else {
				numHealthy--
				if numHealthy == len(checker.agents)-1 { // Overall state has become unhealthy.
					checker.registration.Stop()
				}
			}
		case <-checker.stop:
			// Stop the agents.
			for _, agent := range checker.agents {
				agent.Stop()
			}

			// Stop the monitors.
			for _, statusChan := range statusChans {
				close(statusChan)
			}
		}
	}
}
