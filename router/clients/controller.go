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

package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/resources"
	"github.com/amalgam8/sidecar/config"
)

// Controller TODO
type Controller interface {
	Register() error
	GetNGINXConfig(version *time.Time) (*resources.NGINXJson, error)
	GetCredentials() (TenantCredentials, error)
}

// Registry TODO
type Registry struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

// Kafka TODO
type Kafka struct {
	APIKey   string   `json:"api_key"`
	AdminURL string   `json:"admin_url"`
	RestURL  string   `json:"rest_url"`
	Brokers  []string `json:"brokers"`
	User     string   `json:"user"`
	Password string   `json:"password"`
	SASL     bool     `json:"sasl"`
}

type tenantInfo struct {
	Credentials TenantCredentials `json:"credentials"`
	Port        int               `json:"port"`
}

// TenantCredentials credentials
type TenantCredentials struct {
	Kafka    Kafka    `json:"kafka"`
	Registry Registry `json:"registry"`
}

type controller struct {
	config *config.Config
	client http.Client
}

// NewController TODO
func NewController(conf *config.Config) Controller {
	return &controller{
		config: conf,
		client: http.Client{},
	}
}

// Register TODO
func (c *controller) Register() error {

	bodyJSON := tenantInfo{
		Credentials: TenantCredentials{
			Kafka: Kafka{
				APIKey:   c.config.Kafka.APIKey,
				User:     c.config.Kafka.Username,
				Password: c.config.Kafka.Password,
				Brokers:  c.config.Kafka.Brokers,
				RestURL:  c.config.Kafka.RestURL,
				AdminURL: c.config.Kafka.RestURL,
			},
			Registry: Registry{
				URL:   c.config.Registry.URL,
				Token: c.config.Registry.Token,
			},
		},
		Port: c.config.Nginx.Port,
	}

	bodyBytes, err := json.Marshal(bodyJSON)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"url":    c.config.Controller.URL + "/v1/tenants",
			"method": "POST",
		}).Warn("Error marshalling JSON body")
		return err
	}
	reader := bytes.NewReader(bodyBytes)

	req, err := http.NewRequest("POST", c.config.Controller.URL+"/v1/tenants", reader)
	req.Header.Set("Content-type", "application/json")
	// TODO set Authorization header
	req.Header.Set("Authorization", c.config.Tenant.Token)

	resp, err := c.client.Do(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			//			"request_id": reqID,
		}).Warn("Failed to register with Controller")
		return &ConnectionError{Message: err.Error()}
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		logrus.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			//			"request_id": reqID,
			"body": string(respBytes),
		}).Warn("Controller returned bad response code")

		if resp.Header.Get("request-id") == "" {
			return &NetworkError{Response: resp}
		}

		switch resp.StatusCode {
		case http.StatusNotFound:
			return &TenantNotFoundError{}

		case http.StatusServiceUnavailable:
			return &ServiceUnavailable{}
		case http.StatusConflict:
			return &ConflictError{}
		default:
			return errors.New("Controller returned bad response code") // FIXME: custom error?
		}

	}

	return nil
}

func (c *controller) GetNGINXConfig(version *time.Time) (*resources.NGINXJson, error) {

	url, err := url.Parse(c.config.Controller.URL + "/v1/nginx")
	if err != nil {
		return nil, err
	}
	if version != nil {
		query := url.Query()
		query.Add("version", version.Format(time.RFC3339))
		url.RawQuery = query.Encode()
	}

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			//			"request_id": reqID,
		}).Warn("Error building request to get rules from controller")
		return nil, err
	}
	//TODO set auth header
	req.Header.Set("Authorization", c.config.Tenant.Token)

	resp, err := c.client.Do(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			//			"request_id": reqID,
		}).Warn("Failed to retrieve rules from controller")
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {

		logrus.Debug("No new rules received")
		return "", nil
	} else if resp.StatusCode != http.StatusOK {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		logrus.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			//			"request_id": reqID,
			"body": string(respBytes),
		}).Warn("Controller returned bad response code")
		return nil, errors.New("Controller returned bad response code") // FIXME: custom error?
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			//			"request_id": reqID,
		}).Warn("Error reading rules JSON from controller")
		return nil, err
	}

	templateConf := resources.NGINXJson{}
	if err = json.Unmarshal(body, &templateConf); err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			//			"request_id": reqID,
		}).Warn("Error reading rules JSON from controller")
		return nil, err
	}

	return &templateConf, err
}

func (c *controller) GetCredentials() (TenantCredentials, error) {

	respJSON := struct {
		Credentials TenantCredentials `json:"credentials"`
	}{}

	url, err := url.Parse(c.config.Controller.URL + "/v1/tenants")
	if err != nil {
		return respJSON.Credentials, err
	}

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			//			"request_id": reqID,
		}).Warn("Error building request to get creds from Controller")
		return respJSON.Credentials, err
	}
	//TODO set auth header
	req.Header.Set("Authorization", c.config.Tenant.Token)

	resp, err := c.client.Do(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			//			"request_id": reqID,
		}).Warn("Failed to retrieve creds from Controller")
		return respJSON.Credentials, &ConnectionError{Message: err.Error()}
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		logrus.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			//			"request_id": reqID,
			"body": string(respBytes),
		}).Warn("Controller returned bad response code")

		if resp.Header.Get("request-id") == "" {
			return respJSON.Credentials, &NetworkError{Response: resp}
		}

		switch resp.StatusCode {
		case http.StatusNotFound:
			return respJSON.Credentials, &TenantNotFoundError{}

		case http.StatusServiceUnavailable:
			return respJSON.Credentials, &ServiceUnavailable{}

		default:
			return respJSON.Credentials, errors.New("Controller returned bad response code") // FIXME: custom error?
		}

	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			//			"request_id": reqID,
		}).Warn("Error reading rules JSON from Controller")
		return respJSON.Credentials, err
	}

	err = json.Unmarshal(body, &respJSON)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			//			"request_id": reqID,
		}).Warn("Error reading creds JSON from Controller")
		return respJSON.Credentials, err
	}

	return respJSON.Credentials, nil
}
