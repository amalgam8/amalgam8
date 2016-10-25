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
	"fmt"

	"github.com/amalgam8/amalgam8/sidecar/config"
)

// BuildAgents constructs health check agents.
func BuildAgents(confs []config.HealthCheck) ([]*Agent, error) {
	healthChecks := make([]*Agent, len(confs))
	for i, conf := range confs {
		healthCheck, err := BuildAgent(conf)
		if err != nil {
			return nil, err
		}
		healthChecks[i] = healthCheck
	}
	return healthChecks, nil
}

// BuildAgent constructs a health check agent using the given configuration.
func BuildAgent(conf config.HealthCheck) (*Agent, error) {
	var check Check
	var err error

	switch conf.Type {
	case config.HTTPHealthCheck, config.HTTPSHealthCheck:
		check, err = NewHTTP(conf)
	case config.TCPHealthCheck:
		check, err = NewTCP(conf)
	case config.CommandHealthCheck:
		check, err = NewCommand(conf)
	default:
		return nil, fmt.Errorf("Healthcheck type not supported: '%s'", conf.Type)
	}

	return NewAgent(check, conf.Timeout), err
}
