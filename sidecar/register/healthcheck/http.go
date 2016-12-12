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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/sidecar/config"
)

const (
	defaultHTTPInterval = 30 * time.Second
	defaultHTTPTimeout  = 5 * time.Second
	defaultHTTPMethod   = http.MethodGet
	defaultHTTPCode     = 200
)

// HTTP performs periodic HTTP health checks.
type HTTP struct {
	client *http.Client

	url    string
	method string
	code   int
}

// NewHTTP creates a new HTTP health check.
func NewHTTP(conf config.HealthCheck) (Check, error) {
	if err := validateHTTPConfig(&conf); err != nil {
		return nil, err
	}
	var client *http.Client
	//if conf.UpstreamTrustPath != "" {
	if conf.Type == config.HTTPSHealthCheck &&
		conf.CACertPath != "" {

		logrus.Debug("HTTPS HealthCheck, CA cert path = " + conf.CACertPath)
		// Load trusted CA certs
		caCerts, err := ioutil.ReadFile(conf.CACertPath)
		if err != nil {
			logrus.WithError(err).Debug("Error reading CA trust .pem file: " + conf.CACertPath)
			return nil, err
		}
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(caCerts)
		tlsConfig := &tls.Config{RootCAs: certPool}
		tlsConfig.BuildNameToCertificate()
		transport := &http.Transport{TLSClientConfig: tlsConfig}
		client = &http.Client{Transport: transport,
			Timeout: conf.Timeout}
	} else {
		logrus.Debug("HTTP HealthCheck: " + conf.Type + " " + conf.CACertPath)
		client = &http.Client{Timeout: conf.Timeout}
	}
	return &HTTP{
		url:    conf.Value,
		client: client,
		method: conf.Method,
		code:   conf.Code,
	}, nil
}

// validateHTTPConfig validates, sanitizes, and sets defaults for an HTTP health check configuration.
func validateHTTPConfig(conf *config.HealthCheck) error {

	// Validate healthcheck type
	switch conf.Type {
	case "", config.HTTPHealthCheck, config.HTTPSHealthCheck:
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
		conf.Interval = defaultHTTPInterval
	}

	// Validate timeout
	if conf.Timeout == 0 {
		conf.Timeout = defaultHTTPTimeout
	}

	// Validate method
	switch conf.Method {
	case http.MethodGet, http.MethodHead, http.MethodPost,
		http.MethodPut, http.MethodPatch, http.MethodDelete,
		http.MethodConnect, http.MethodOptions, http.MethodTrace:
	case "":
		conf.Method = defaultHTTPMethod
	default:
		return fmt.Errorf("invalid method for an HTTP healthcheck: '%s'", conf.Method)
	}

	// Validate code
	if conf.Code == 0 {
		conf.Code = defaultHTTPCode
	} else if conf.Code < 100 || conf.Code > 599 {
		return fmt.Errorf("invalid code for an HTTP healthcheck: '%d'", conf.Code)
	}

	return nil
}

// Execute a HTTP operation on the URL and check the response code.
func (h *HTTP) Execute() error {
	req, err := http.NewRequest(h.method, h.url, nil)
	if err != nil {
		return err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != h.code {
		return fmt.Errorf("HTTP/HTTPS health check expected %v, got %v", h.code, resp.StatusCode)
	}

	return nil
}
