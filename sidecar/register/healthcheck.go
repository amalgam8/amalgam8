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
	"fmt"
	"net/url"

	"github.com/amalgam8/amalgam8/sidecar/config"
)

// HealthStatus is the reported health status reported by a health check.
type HealthStatus struct {
	HealthCheck HealthCheck
	Error       error
}

// HealthCheck is an interface for performing a health check.
type HealthCheck interface {
	Start(statusChan chan HealthStatus)
	Stop()
}

// BuildHealthChecks constructs health checks.
func BuildHealthChecks(checkConfs []config.HealthCheck) ([]HealthCheck, error) {
	healthChecks := make([]HealthCheck, len(checkConfs))
	for i, conf := range checkConfs {
		healthCheck, err := BuildHealthCheck(conf)
		if err != nil {
			return nil, err
		}
		healthChecks[i] = healthCheck
	}
	return healthChecks, nil
}

// BuildHealthCheck builds a HealthCheck out of the given health check configuration
func BuildHealthCheck(checkConf config.HealthCheck) (HealthCheck, error) {
	hcType := checkConf.Type
	if hcType == "" {
		// Parse the healthcheck type from URL scheme
		u, err := url.Parse(checkConf.Value)
		if err != nil {
			return nil, fmt.Errorf("Could not parse healthcheck value: '%s'", checkConf.Value)
		}

		hcType = u.Scheme
	}

	switch hcType {
	case "http", "https": // TODO: constants (when extra types come)
		return NewHTTPHealthCheck(checkConf)
	default:
		return nil, fmt.Errorf("Healthcheck type not supported: '%s'", hcType)
	}
}
