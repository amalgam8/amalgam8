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
	"net/http"
	"sync"
	"time"

	"net/url"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/sidecar/config"
)

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

// HTTPHealthCheck performs periodic HTTP health checks.
type HTTPHealthCheck struct {
	url      string
	client   *http.Client
	interval time.Duration
	method   string
	code     int

	stop   chan interface{}
	active bool
	mutex  sync.Mutex
}

// NewHTTPHealthCheck creates a new HTTP health check from the given configuration
func NewHTTPHealthCheck(checkConf config.HealthCheck) (*HTTPHealthCheck, error) {
	// TODO: Validate and set defaults
	// TODO: extract http health checks into http_healthcheck.go
	return &HTTPHealthCheck{
		url: checkConf.Value,
		client: &http.Client{
			Timeout: checkConf.Timeout,
		},
		interval: checkConf.Timeout,
		method:   checkConf.Method,
		code:     checkConf.Code,
	}, nil
}

// Start HTTP health check.
func (c *HTTPHealthCheck) Start(statusChan chan HealthStatus) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.active {
		return
	}
	c.active = true

	go c.check(statusChan)
}

// Stop HTTP health check.
func (c *HTTPHealthCheck) Stop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.active {
		return
	}
	c.active = false

	c.stop <- struct{}{}
}

// Start periodic health checks of a HTTP address. Perform an HTTP operation on the URL and check the response code.
func (c *HTTPHealthCheck) check(statusChan chan HealthStatus) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			logrus.WithField("url", c.url).Debug("Performing HTTP/HTTPS health check")

			status := HealthStatus{
				HealthCheck: c,
			}

			req, err := http.NewRequest(c.method, c.url, nil)
			if err != nil {
				status.Error = err
				statusChan <- status
				continue
			}

			resp, err := c.client.Do(req)
			if err != nil {
				status.Error = err
				statusChan <- status
				continue
			}

			if resp.StatusCode != c.code {
				status.Error = fmt.Errorf("HTTP/HTTPS health check expected %v, got %v", c.code, resp.StatusCode)
				statusChan <- status
				continue
			}

			statusChan <- status
		case <-c.stop:
			break
		}
	}
}
