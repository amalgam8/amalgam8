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

const (
	defaultHTTPHealthcheckInterval = 30 * time.Second
	defaultHTTPHealthcheckTimeout  = 5 * time.Second
	defaultHTTPHealthcheckMethod   = http.MethodGet
	defaultHTTPHealthcheckCode     = 200
)

// HTTPHealthCheck performs periodic HTTP health checks.
type HTTPHealthCheck struct {
	client *http.Client

	url      string
	interval time.Duration
	method   string
	code     int

	stop   chan interface{}
	active bool
	mutex  sync.Mutex
}

// NewHTTPHealthCheck creates a new HTTP health check from the given configuration
func NewHTTPHealthCheck(conf config.HealthCheck) (*HTTPHealthCheck, error) {
	err := validateHTTPConfig(&conf)
	if err != nil {
		return nil, err
	}

	return &HTTPHealthCheck{
		url: conf.Value,
		client: &http.Client{
			Timeout: conf.Timeout,
		},
		interval: conf.Interval,
		method:   conf.Method,
		code:     conf.Code,
	}, nil
}

// validateHTTPConfig validates, sanitizes, and sets defaults for an HTTP healthcheck configuration
func validateHTTPConfig(conf *config.HealthCheck) error {

	// Validate healthcheck type
	switch conf.Type {
	case "", "http", "https":
	default:
		return fmt.Errorf("invalid type for an HTTP healthcheck: '%s'", conf.Type)
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
		conf.Interval = defaultHTTPHealthcheckInterval
	}

	// Validate timeout
	if conf.Timeout == 0 {
		conf.Timeout = defaultHTTPHealthcheckTimeout
	}

	// Validate method
	switch conf.Method {
	case http.MethodGet, http.MethodHead, http.MethodPost,
		http.MethodPut, http.MethodPatch, http.MethodDelete,
		http.MethodConnect, http.MethodOptions, http.MethodTrace:
	case "":
		conf.Method = defaultHTTPHealthcheckMethod
	default:
		return fmt.Errorf("invalid method for an HTTP healthcheck: '%s'", conf.Method)
	}

	// Validate code
	if conf.Code == 0 {
		conf.Code = defaultHTTPHealthcheckCode
	} else if conf.Code < 100 || conf.Code > 599 {
		return fmt.Errorf("invalid code for an HTTP healthcheck: '%d'", conf.Code)
	}

	return nil
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
