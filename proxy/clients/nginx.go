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
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
)

// NGINX client
type NGINX interface {
	UpdateHTTPUpstreams(conf NGINXJson) error
}

// NGINXJson sent to update http/https endpoints
type NGINXJson struct {
	Upstreams map[string]NGINXUpstream `json:"upstreams"`
	Services  map[string]NGINXService  `json:"services"`
	Faults    []NGINXFault             `json:"faults,omitempty"`
}

// NGINXService version info for lua
type NGINXService struct {
	Default   string `json:"default"`
	Selectors string `json:"selectors,omitempty"`
	Type      string `json:"type"`
}

// NGINXUpstream server info for lua
type NGINXUpstream struct {
	Upstreams []NGINXEndpoint `json:"servers"`
}

// NGINXEndpoint for lua
type NGINXEndpoint struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// NGINXVersion for lua
type NGINXVersion struct {
	Service   string `json:"service"`
	Default   string `json:"default"`
	Selectors string `json:"selectors"`
}

// NGINXFault for representing fault injection for lua
type NGINXFault struct {
	Source           string  `json:"source"`
	Destination      string  `json:"destination"`
	Header           string  `json:"header"`
	Pattern          string  `json:"pattern"`
	Delay            float64 `json:"delay"`
	DelayProbability float64 `json:"delay_probability"`
	AbortProbability float64 `json:"abort_probability"`
	AbortCode        int     `json:"return_code"`
}

type nginx struct {
	httpClient *http.Client
	url        string
}

// NewNGINX return new NGINX client
func NewNGINX(url string) NGINX {
	return &nginx{
		httpClient: &http.Client{},
		url:        url,
	}
}

// UpdateHTTPUpstreams updates http upstreams in lua dynamically
func (n *nginx) UpdateHTTPUpstreams(conf NGINXJson) error {

	data, err := json.Marshal(&conf)
	if err != nil {
		logrus.WithError(err).Error("Could not marshal request body")
		return err
	}

	reader := bytes.NewReader(data)
	req, err := http.NewRequest("POST", n.url+"/a8-admin", reader)
	if err != nil {
		logrus.WithError(err).Error("Building request for NGINX server failed")
		return err
	}

	resp, err := n.httpClient.Do(req)
	if err != nil {
		logrus.WithError(err).Error("Failed to send request to NGINX server")
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)

		logrus.WithError(err).WithFields(logrus.Fields{
			"body":        string(data),
			"status_code": resp.StatusCode,
		}).Error("POST to NGINX server failed")
		return err
	}

	return nil
}
