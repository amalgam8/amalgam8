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

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/registry/server/protocol/amalgam8"
)

const defaultTimeout = time.Second * 30

// Make sure we implement both the ServiceDiscovery and ServiceRegistry interfaces.
var _ api.ServiceDiscovery = (*Client)(nil)
var _ api.ServiceRegistry = (*Client)(nil)

// Config stores the configurable attributes of the client.
type Config struct {

	// URL of the registry server.
	URL string

	// AuthToken is the bearer token to be used for authentication with the registry.
	// If left empty, no authentication is used.
	AuthToken string

	// HTTPClient can be used to customize the underlying HTTP client behavior,
	// such as enabling TLS, setting timeouts, etc.
	// If left nil, a default HTTP client will be used.
	HTTPClient *http.Client
}

// Client implements the ServiceDiscovery and ServiceRegistry interfaces using Amalgam8 Registry REST API.
type Client struct {
	config     Config
	httpClient *http.Client
}

// New constructs a new Client using the given configuration.
func New(config Config) (*Client, error) {
	// Validate and normalize configuration
	err := normalizeConfig(&config)
	if err != nil {
		return nil, err
	}

	client := &Client{
		config: config,
	}

	if config.HTTPClient != nil {
		client.httpClient = config.HTTPClient
	} else {
		client.httpClient = &http.Client{
			Timeout: defaultTimeout,
		}
	}

	return client, nil
}

// Register adds a service instance, described by the given ServiceInstance structure, to the registry.
func (client *Client) Register(instance *api.ServiceInstance) (*api.ServiceInstance, error) {
	// Record a pessimistic last heartbeat time - Better safe than sorry!
	lastHeartbeat := time.Now()

	body, err := client.doRequest("POST", amalgam8.InstanceCreateURL(), instance, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(body, &m)
	if err != nil {
		return nil, newError(ErrorCodeInternalClientError, "error unmarshaling HTTP response body", err, "")
	}

	// TODO: recover type conversion panic
	registeredInstance := &*instance
	registeredInstance.ID = m["id"].(string)
	registeredInstance.TTL = int(m["ttl"].(float64))
	registeredInstance.LastHeartbeat = lastHeartbeat
	return registeredInstance, nil
}

// Deregister removes a registered service instance, identified by the given ID, from the registry.
func (client *Client) Deregister(id string) error {
	_, err := client.doRequest("DELETE", amalgam8.InstanceURL(id), nil, http.StatusOK)
	return err
}

// Renew sends a heartbeat for the service instance identified by the given ID.
func (client *Client) Renew(id string) error {
	_, err := client.doRequest("PUT", amalgam8.InstanceHeartbeatURL(id), nil, http.StatusOK)
	return err
}

// ListServices queries for the list of services for which instances are currently registered.
func (client *Client) ListServices() ([]string, error) {
	body, err := client.doRequest("GET", amalgam8.ServiceNamesURL(), nil, http.StatusOK)
	if err != nil {
		return nil, err
	}

	s := struct {
		Services []string `json:"services"`
	}{}
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, newError(ErrorCodeInternalClientError, "error unmarshaling HTTP response body", err, "")
	}
	return s.Services, nil
}

// ListServiceInstances queries for the list of service instances currently registered for the given service.
func (client *Client) ListServiceInstances(serviceName string) ([]*api.ServiceInstance, error) {
	return client.ListInstancesWithFilter(InstanceFilter{
		ServiceName: serviceName,
	})
}

// ListInstances queries for the list of service instances currently registered.
func (client *Client) ListInstances() ([]*api.ServiceInstance, error) {
	return client.ListInstancesWithFilter(InstanceFilter{})
}

// ListInstancesWithFilter queries for the list of service instances currently registered that satisfy the given filter.
func (client *Client) ListInstancesWithFilter(filter InstanceFilter) ([]*api.ServiceInstance, error) {
	path := amalgam8.InstancesURL()
	queryParams := filter.asQueryParams()
	if len(queryParams) > 0 {
		path = fmt.Sprintf("%s?%s", path, queryParams.Encode())
	}

	body, err := client.doRequest("GET", path, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}

	s := struct {
		Instances []*api.ServiceInstance `json:"instances"`
	}{}
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, newError(ErrorCodeInternalClientError, "error unmarshaling HTTP response body", err, "")
	}
	return s.Instances, nil
}

func (client *Client) doRequest(method string, path string, body interface{}, status int) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, newError(ErrorCodeInternalClientError, "error marshaling HTTP request body", err, "")
		}
		reader = bytes.NewBuffer(b)
	}

	uri := client.config.URL + path
	req, err := http.NewRequest(method, uri, reader)
	if err != nil {
		return nil, newError(ErrorCodeInternalClientError, "error creating HTTP request", err, "")
	}

	// Add authorization header
	if client.config.AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.config.AuthToken))
	}

	if body != nil {
		// Body exists, and encoded as JSON
		req.Header.Set("Content-Type", "application/json")
	} else if method != "GET" {
		// No body, but the server needs to know that too
		req.Header.Set("Content-Length", "0")
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, newError(ErrorCodeConnectionFailure, "error performing HTTP request", err, "")
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, newError(ErrorCodeConnectionFailure, "error read HTTP response body", err, "")
	}

	if resp.StatusCode == status {
		return b, nil
	}

	requestID := resp.Header.Get("Sd-Request-Id")
	message := string(b)
	if requestID != "" {
		s := struct {
			Error string `json:"Error"`
		}{}
		err = json.Unmarshal(b, &s)
		if err != nil {
			message = s.Error
		}
	}
	switch resp.StatusCode {
	case http.StatusGone:
		return nil, newError(ErrorCodeUnknownInstance, message, nil, requestID)
	case http.StatusNotFound:
		if requestID != "" {
			return nil, newError(ErrorCodeInternalClientError, message, nil, requestID)
		}
		return nil, newError(ErrorCodeServiceUnavailable, message, nil, requestID)
	case http.StatusBadGateway:
		return nil, newError(ErrorCodeServiceUnavailable, message, nil, requestID)
	case http.StatusUnauthorized:
		return nil, newError(ErrorCodeUnauthorized, message, nil, requestID)
	case http.StatusInternalServerError:
		return nil, newError(ErrorCodeInternalServerError, message, nil, requestID)
	default:
		return nil, newError(ErrorCodeInternalClientError, message, nil, requestID)
	}
}

func normalizeConfig(config *Config) error {
	if config == nil {
		return newError(ErrorCodeInvalidConfiguration, "client configuration cannot be nil", nil, "")
	}

	// Normalize server URL to not end with a "/"
	config.URL = strings.TrimRight(config.URL, "/")

	url, err := url.Parse(config.URL)
	if err != nil {
		return newError(ErrorCodeInvalidConfiguration, "cannot parse server URL", err, "")
	}

	if url.Scheme != "http" && url.Scheme != "https" {
		return newError(ErrorCodeInvalidConfiguration, fmt.Sprintf("unsupported scheme %s", url.Scheme), nil, "")
	}

	return nil
}
