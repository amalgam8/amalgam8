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
)

// Client defines the interface used by clients of the Amalgam8 Service Registry.
type Client interface {
	Register(instance *ServiceInstance) (*ServiceInstance, error)
	Deregister(id string) error
	Renew(id string) error
	ListServices() ([]string, error)
	ListInstances(instanceFilter *InstanceFilter) ([]*ServiceInstance, error)
	ListServiceInstances(serviceName string) ([]*ServiceInstance, error)
}

// TODO: Allow custom transport/HTTP client/TLS config
// TODO: revamp URLs construction (resolve, URL-encode)

type ClientConfig struct {
	URL       string `json:"url"`
	AuthToken string `json:"auth_token"`
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
		httpClient: &http.Client{},
	}
	return client, nil
}

func (client *RESTClient) Register(instance *ServiceInstance) (*ServiceInstance, error) {
	// Record a pessimistic last heartbeat time - Better safe than sorry!
	lastHeartbeat := time.Now()

	url := fmt.Sprintf("%s/api/v1/instances", client.config.URL)
	body, err := client.doRequest("POST", url, instance, http.StatusCreated)
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
	uri := fmt.Sprintf("%s/api/v1/instances/%s", client.config.URL, id)
	_, err := client.doRequest("DELETE", uri, nil, http.StatusOK)
	return err
}

func (client *RESTClient) Renew(id string) error {
	uri := fmt.Sprintf("%s/api/v1/instances/%s/heartbeat", client.config.URL, id)
	_, err := client.doRequest("PUT", uri, nil, http.StatusOK)
	return err
}

func (client *RESTClient) ListServices() ([]string, error) {
	uri := fmt.Sprintf("%s/api/v1/services", client.config.URL)
	body, err := client.doRequest("GET", uri, nil, http.StatusOK)
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
	uri := fmt.Sprintf("%s/api/v1/instances", client.config.URL)

	if instanceFilter != nil {
		queryParams := instanceFilter.asQueryParams()
		if len(queryParams) > 0 {
			uri = fmt.Sprintf("%s?%s", uri, queryParams.Encode())
		}
		instanceFilter.asQueryParams()
	}

	body, err := client.doRequest("GET", uri, nil, http.StatusOK)
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

func (client *RESTClient) doRequest(method string, uri string, body interface{}, status int) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, newError(ErrorCodeInternalClientError, "error marshaling HTTP request body", err, "")
		}
		reader = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, uri, reader)
	if err != nil {
		return nil, newError(ErrorCodeInternalClientError, "error creating HTTP request", err, "")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.config.AuthToken))
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
