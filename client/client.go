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

// Client defines the interface used by clients of the Amalgam8 Service Registry.
type Client interface {
	Register(instance *ServiceInstance) (*ServiceInstance, error)
	Deregister(id string) error
	Renew(id string) error
	ListServices() ([]string, error)
	ListInstances(instanceFilter *InstanceFilter) ([]*ServiceInstance, error)
	ListServiceInstances(serviceName string) ([]*ServiceInstance, error)
}

type ClientConfig struct {
	URL       string
	AuthToken string
	HTTPClient *http.Client
}

// RESTClient implements the Client interface using Amalgam8 Service Registry REST API.
type RESTClient struct {
	config     ClientConfig
	httpClient *http.Client
}

func NewRESTClient(config ClientConfig) (*RESTClient, error) {
	// Validate and normalize configuration
	err := normalizeConfig(&config)
	if err != nil {
		return nil, err
	}

	client := &RESTClient{
		config:     config,
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

func (client *RESTClient) Register(instance *ServiceInstance) (*ServiceInstance, error) {
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

func (client *RESTClient) Deregister(id string) error {
	_, err := client.doRequest("DELETE", amalgam8.InstanceURL(id), nil, http.StatusOK)
	return err
}

func (client *RESTClient) Renew(id string) error {
	_, err := client.doRequest("PUT", amalgam8.InstanceHeartbeatURL(id), nil, http.StatusOK)
	return err
}

func (client *RESTClient) ListServices() ([]string, error) {
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

func (client *RESTClient) ListServiceInstances(serviceName string) ([]*ServiceInstance, error) {
	return client.ListInstances(&InstanceFilter{
		ServiceName: serviceName,
	})
}

func (client *RESTClient) ListInstances(instanceFilter *InstanceFilter) ([]*ServiceInstance, error) {
	path := amalgam8.InstancesURL()
	if instanceFilter != nil {
		queryParams := instanceFilter.asQueryParams()
		if len(queryParams) > 0 {
			path = fmt.Sprintf("%s?%s", path, queryParams.Encode())
		}
		instanceFilter.asQueryParams()
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

func (client *RESTClient) doRequest(method string, path string, body interface{}, status int) ([]byte, error) {
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
		} else {
			return nil, newError(ErrorCodeServiceUnavailable, message, nil, requestID)
		}
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

func normalizeConfig(config *ClientConfig) error {
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

	url.Query()
	if config.AuthToken == "" {
		return newError(ErrorCodeInvalidConfiguration, "missing authentication token", nil, "")
	}

	return nil
}
