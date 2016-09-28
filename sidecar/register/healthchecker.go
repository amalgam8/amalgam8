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

	"github.com/Sirupsen/logrus"
)

// HealthChecker uses a collection of periodic health checks to manage registration. Registration is maintained while
// the state of the health checker is healthy. If the state changes to unhealthy, the registration agent is stopped.
// Similarly, if the state changes to healthy, the registration agent is started again. If any health check's last
// reported status is unhealthy, the health checker is considered unhealthy.
type HealthChecker struct {
	active       bool
	stop         chan struct{}
	checks       []HealthCheck
	mutex        sync.Mutex
	registration *RegistrationAgent
}

// NewHealthChecker instantiates a health checker.
func NewHealthChecker(registration *RegistrationAgent, checks []HealthCheck) *HealthChecker {
	if len(checks) == 0 {
		panic("No health checks provided")
	}

	return &HealthChecker{
		checks:       checks,
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

	go checker.report()
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

func (checker *HealthChecker) report() {
	wasHealthy := false // Initialize as unhealthy so that we register on the first successful health check

	statusChan := make(chan HealthStatus, len(checker.checks))
	for _, healthCheck := range checker.checks {
		healthCheck.Start(statusChan)
	}

	// Set of health checks that have most recently reported unhealthy statuses.
	unhealthyStatuses := make(map[HealthCheck]interface{})
	for {
		select {
		case status := <-statusChan:
			logrus.WithField("status", status).Debug("Recieved health status")

			// Update our set
			if status.Error != nil {
				unhealthyStatuses[status.HealthCheck] = struct{}{}
			} else {
				delete(unhealthyStatuses, status.HealthCheck)
			}

			healthy := len(unhealthyStatuses) == 0
			if healthy && !wasHealthy {
				logrus.Debug("Service is healthy, registering")
				checker.registration.Start()
			} else if !healthy && wasHealthy {
				logrus.Warn("Service is unhealthy, unregistering")
				checker.registration.Stop()
			}

			// Record overall health state for next status
			wasHealthy = healthy
		case <-checker.stop:
			for _, healthCheck := range checker.checks {
				healthCheck.Stop()
			}
		}
	}

}
