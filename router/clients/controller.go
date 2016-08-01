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
	GetProxyConfig(version *time.Time) (*resources.ProxyConfig, error)
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

func (c *controller) GetProxyConfig(version *time.Time) (*resources.ProxyConfig, error) {
	url, err := url.Parse(c.config.Controller.URL + "/v1/tenants")
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
	//TODO handle global auth
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
		return nil, nil
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

	respJSON := &resources.ProxyConfig{}
	if err = json.Unmarshal(body, respJSON); err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			//			"request_id": reqID,
		}).Warn("Error reading rules JSON from controller")
		return nil, err
	}

	return respJSON, nil
}
