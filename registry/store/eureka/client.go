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

package eureka

import (
	"fmt"
	"strings"
	"time"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"

	eurekaapi "github.com/amalgam8/amalgam8/registry/api/protocol/eureka"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

const (
	requestTimeout = time.Duration(60) * time.Second
)

type eurekaClient struct {
	httpClient *http.Client
	eurekaURLs []string

	logger *log.Entry
}

func newEurekaClient(eurekaURLs []string) (*eurekaClient, error) {

	logger := logging.GetLogger(module)

	urls := make([]string, len(eurekaURLs))
	for i, u := range eurekaURLs {
		for strings.HasSuffix(u, "/") {
			u = strings.TrimSuffix(u, "/")
		}

		if _, err := url.Parse(u); err != nil {
			return nil, err
		}

		urls[i] = u
	}

	hc := &http.Client{
		Timeout: requestTimeout,
	}

	return &eurekaClient{
		httpClient: hc,
		eurekaURLs: urls,
		logger:     logger,
	}, nil
}

func (client *eurekaClient) getApplications(path string) (*eurekaapi.Applications, error) {
	var err error

	for _, eurl := range client.eurekaURLs {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/%s", eurl, path), nil)
		req.Header.Set("Accept", "application/json")

		resp, err2 := client.httpClient.Do(req)
		if err2 != nil {
			err = err2
			continue
		}

		defer resp.Body.Close()

		body, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			err = err2
			continue
		}

		var appsList eurekaapi.ApplicationsList
		err2 = json.Unmarshal(body, &appsList)
		if err2 != nil {
			err = err2
			continue
		}

		return appsList.Applications, nil
	}

	return nil, err
}

func (client *eurekaClient) getApplicationsFull() (*eurekaapi.Applications, error) {
	return client.getApplications("apps")
}

func (client *eurekaClient) getApplicationsDelta() (*eurekaapi.Applications, error) {
	apps, err := client.getApplications("apps/delta")
	if err != nil {
		return nil, err
	}

	if apps == nil || apps.VersionDelta == -1 {
		return nil, fmt.Errorf("Delta is not supported")
	}

	return apps, nil
}
