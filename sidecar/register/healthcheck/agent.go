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
	"sync"
	"time"
)

const (
	defaultHealthCheckInterval = 30 * time.Second
)

// Agent executes a health check a given interval.
type Agent struct {
	stop   chan interface{}
	active bool
	mutex  sync.Mutex

	interval    time.Duration
	healthCheck Check
}

// NewAgent creates a new health check agent.
func NewAgent(check Check, interval time.Duration) *Agent {
	if interval == 0 {
		interval = defaultHealthCheckInterval
	}

	return &Agent{
		healthCheck: check,
		interval:    interval,
	}
}

// Start health check agent.
func (a *Agent) Start(statusChan chan error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.active {
		return
	}
	a.active = true

	go a.run(statusChan)
}

// Stop health check agent.
func (a *Agent) Stop() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !a.active {
		return
	}
	a.active = false

	a.stop <- struct{}{}
}

// run periodic health checks until the agent is stopped.
func (a *Agent) run(statusChan chan error) {
	// Perform an initial health check on start.
	statusChan <- a.healthCheck.Execute()

	// Begin periodic checks.
	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			statusChan <- a.healthCheck.Execute()
		case <-a.stop:
			break
		}
	}
}
