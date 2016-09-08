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

	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

const (
	k8sTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

type k8sClient struct {
	httpClient *http.Client
	k8sURL     string
	k8sToken   string

	logger *log.Entry
}

func newK8sClient(k8sURL, k8sToken string) (*k8sClient, error) {

	logger := logging.GetLogger(module)

	// Normalize k8sURL to not end with a slash
	for strings.HasSuffix(k8sURL, "/") {
		k8sURL = strings.TrimSuffix(k8sURL, "/")
	}

	u, err := url.Parse(k8sURL)
	if err != nil {
		return nil, err
	}

	if k8sToken == "" {
		t, err := ioutil.ReadFile(k8sTokenFile)
		if err != nil {
			logger.Warnf("Failed to read kubernetes token. %s", err)
		}
		k8sToken = string(t)
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
		k8sToken:   k8sToken,
		logger:     logger,
	}, nil
}

func (client *k8sClient) getEndpointsURL(namespace auth.Namespace) string {
	return fmt.Sprintf("%s/api/v1/namespaces/%s/endpoints", client.k8sURL, namespace)
}

func (client *k8sClient) getEndpointsList(namespace auth.Namespace) (*EndpointsList, error) {
	endpointsList := EndpointsList{}

	req, _ := http.NewRequest("GET", client.getEndpointsURL(namespace), nil)
	if client.k8sToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.k8sToken))
	}
	resp, err := client.httpClient.Do(req)
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
