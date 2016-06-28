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

package kubernetes

import (
	"fmt"
	"strings"
	"time"

	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"

	"github.com/amalgam8/registry/auth"
	"github.com/amalgam8/registry/utils/logging"
)

type k8sClient struct {
	httpClient *http.Client
	k8sURL     string
	k8sAuth    string

	logger *log.Entry
}

func newK8sClient(k8sURL string) (*k8sClient, error) {

	// Normalize k8sURL to not end with a slash
	for strings.HasSuffix(k8sURL, "/") {
		k8sURL = strings.TrimSuffix(k8sURL, "/")
	}

	u, err := url.Parse(k8sURL)
	if err != nil {
		return nil, err
	}

	t, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return nil, err
	}

	hc := &http.Client{
		Timeout: time.Duration(2 * time.Second),
	}

	if u.Scheme == "https" {
		hc.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true, ServerName: u.Host},
		}
	}

	return &k8sClient{
		httpClient: hc,
		k8sURL:     k8sURL,
		k8sAuth:    "Bearer " + string(t),
		logger:     logging.GetLogger(module),
	}, nil
}

func (client *k8sClient) getEndpointsURL(namespace auth.Namespace) string {
	return fmt.Sprintf("%s/api/v1/namespaces/%s/endpoints", client.k8sURL, namespace)
}

func (client *k8sClient) getEndpointsList(namespace auth.Namespace) (*EndpointsList, error) {
	endpointsList := EndpointsList{}

	req, _ := http.NewRequest("GET", client.getEndpointsURL(namespace), nil)
	req.Header.Set("Authorization", client.k8sAuth)
	resp, err := client.httpClient.Do(req)
	//resp, err := client.httpClient.Get(client.getEndpointsURL(namespace))
	if err != nil {
		return nil, fmt.Errorf("Failed to get endpointsURL [%s]", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response [%s]", err)
	}

	json.Unmarshal(body, &endpointsList)
	return &endpointsList, nil
}
