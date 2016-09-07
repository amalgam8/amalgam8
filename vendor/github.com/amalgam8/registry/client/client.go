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

	"github.com/amalgam8/registry/api/protocol/amalgam8"
)

const defaultTimeout = time.Second * 30

// Registry defines the interface used by clients for registering service instances with the registry.
type Registry interface {

	// Register adds a service instance, described by the given ServiceInstance structure, to the registry.
	// The returned ServiceInstance is mostly similar to the given one, but includes additional
	// attributes set by the registry server, such as the service instance ID and TTL.
	Register(instance *ServiceInstance) (*ServiceInstance, error)

	// Deregister removes a registered service instance, identified by the given ID, from the registry.
	Deregister(id string) error

	// Renew sends a heartbeat for the service instance identified by the given ID.
	Renew(id string) error
}

// Discovery defines the interface used by clients for discovering service instances from the registry.
type Discovery interface {

	// ListServices queries the registry for the list of services for which instances are currently registered.
	ListServices() ([]string, error)

	// ListInstances queries the registry for the list of service instances currently registered.
	// The given InstanceFilter can be used to filter the returned instances as well as the fields returned for each.
	ListInstances(filter InstanceFilter) ([]*ServiceInstance, error)

	// ListServiceInstances queries the registry for the list of service instances with status 'UP' currently
	// registered for the given service.
	ListServiceInstances(serviceName string) ([]*ServiceInstance, error)
}

// Client unifies the Registry and Discovery interfaces under a single interface.
type Client interface {
	Registry
	Discovery
}

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

// client implements the Client interface using Amalgam8 Service Registry REST API.
type client struct {
	config     Config
	httpClient *http.Client
}

// New constructs a new Client using the given configuration.
func New(config Config) (Client, error) {
	// Validate and normalize configuration
	err := normalizeConfig(&config)
	if err != nil {
		return nil, err
	}

	client := &client{
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

func (client *client) Register(instance *ServiceInstance) (*ServiceInstance, error) {
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

func (client *client) Deregister(id string) error {
	_, err := client.doRequest("DELETE", amalgam8.InstanceURL(id), nil, http.StatusOK)
	return err
}

func (client *client) Renew(id string) error {
	_, err := client.doRequest("PUT", amalgam8.InstanceHeartbeatURL(id), nil, http.StatusOK)
	return err
}

func (client *client) ListServices() ([]string, error) {
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

func (client *client) ListServiceInstances(serviceName string) ([]*ServiceInstance, error) {
	return client.ListInstances(InstanceFilter{
		ServiceName: serviceName,
	})
}

func (client *client) ListInstances(filter InstanceFilter) ([]*ServiceInstance, error) {
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
		Instances []*ServiceInstance `json:"instances"`
	}{}
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, newError(ErrorCodeInternalClientError, "error unmarshaling HTTP response body", err, "")
	}
	return s.Instances, nil
}

func (client *client) doRequest(method string, path string, body interface{}, status int) ([]byte, error) {
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
