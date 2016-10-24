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
	"net"
	"net/url"
	"time"

	"github.com/amalgam8/amalgam8/sidecar/config"
)

const (
	defaultTCPInterval = 30 * time.Second
	defaultTCPTimeout  = 5 * time.Second
)

// TCP health check.
type TCP struct {
	url     string
	timeout time.Duration
}

// NewTCP creates a new TCP health check.
func NewTCP(conf config.HealthCheck) (Check, error) {
	if err := validateTCPConfig(&conf); err != nil {
		return nil, err
	}

	return &TCP{
		url:     conf.Value,
		timeout: conf.Timeout,
	}, nil
}

// validateTCPConfig validates, sanitizes, and sets defaults for a TCP health check configuration.
func validateTCPConfig(conf *config.HealthCheck) error {

	// Validate health check type
	if conf.Type != config.TCPHealthCheck {
		return fmt.Errorf("invalid type for a TCP healthcheck: '%s'", conf.Type)
	}

	// Validate URL
	if conf.Value == "" {
		return fmt.Errorf("empty URL for HTTP healthcheck")
	}
	u, err := url.Parse(conf.Value)
	if err != nil {
		return fmt.Errorf("error parsing URL '%s' for HTTP healthcheck: %v", conf.Value, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("invalid URL '%s' for HTTP healthcheck", conf.Value)
	}

	// Validate interval
	if conf.Interval == 0 {
		conf.Interval = defaultTCPInterval
	}

	// Validate timeout
	if conf.Timeout == 0 {
		conf.Timeout = defaultTCPTimeout
	}

	return nil
}

// Execute the TCP health check by attempting to connect to the URL via TCP.
func (t *TCP) Execute() error {
	conn, err := net.DialTimeout("tcp", t.url, t.timeout)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
