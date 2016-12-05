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

package nginx

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"errors"

	"time"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/controller/rules"
	"github.com/amalgam8/amalgam8/pkg/api"
)

const clientTimeout = 30 * time.Second

// Client for NGINX
type Client interface {
	Update([]api.ServiceInstance, []rules.Rule) error
}

type client struct {
	httpClient *http.Client
	url        string
}

// NewClient return new NGINX client
func NewClient(url string) Client {
	return &client{
		httpClient: &http.Client{
			Timeout: clientTimeout,
		},
		url: url,
	}
}

// Update the NGINX server
func (c *client) Update(newInstances []api.ServiceInstance, newRules []rules.Rule) error {
	conf := struct {
		Instances []api.ServiceInstance `json:"instances"`
		Rules     []rules.Rule          `json:"rules"`
	}{
		Instances: newInstances,
		Rules:     newRules,
	}

	data, err := json.Marshal(&conf)
	if err != nil {
		logrus.WithError(err).Error("Could not marshal request body")
		return err
	}

	// TODO: remove this overly verbose logging?
	logrus.WithField("data", string(data)).Debug("Updating NGINX server")

	reader := bytes.NewReader(data)
	req, err := http.NewRequest("POST", c.url+"/a8-admin", reader)
	if err != nil {
		logrus.WithError(err).Error("Building request for NGINX server failed")
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logrus.WithError(err).Error("Failed to send request to NGINX server")
		return err
	}

	// Read and close the response body so we can reuse the connection
	data, _ = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{
			"body":        string(data),
			"status_code": resp.StatusCode,
		}).Error("POST to NGINX server failed")
		return errors.New("nginx: POST to NGINX server failed")
	}

	return nil
}
