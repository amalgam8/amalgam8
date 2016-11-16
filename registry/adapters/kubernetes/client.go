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
	urlpkg "net/url"

	log "github.com/Sirupsen/logrus"

	"sync"

	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

const (
	k8sTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

type client struct {
	httpClient *http.Client
	k8sURL     string
	k8sToken   string

	logger *log.Entry
}

func newClient(url, token string) (*client, error) {
	logger := logging.GetLogger(module)

	// Normalize url to not end with a slash
	for strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}

	u, err := urlpkg.Parse(url)
	if err != nil {
		return nil, err
	}

	// check if we have a client cached for this url
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	c, ok := clientCache[url]
	if ok {
		return c, nil
	}

	// Build a new client
	if token == "" {
		t, err := ioutil.ReadFile(k8sTokenFile)
		if err != nil {
			logger.Warnf("Failed to read kubernetes token. %s", err)
		}
		token = string(t)
	}

	hc := &http.Client{
		Timeout: time.Duration(2 * time.Second),
	}

	if u.Scheme == "https" {
		hc.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true, ServerName: u.Host},
		}
	}

	c = &client{
		httpClient: hc,
		k8sURL:     url,
		k8sToken:   token,
		logger:     logger,
	}
	clientCache[url] = c
	return c, nil
}

func (client *client) getEndpointsURL(namespace auth.Namespace) string {
	return fmt.Sprintf("%s/api/v1/namespaces/%s/endpoints", client.k8sURL, namespace)
}

func (client *client) getEndpointsList(namespace auth.Namespace) (*EndpointsList, error) {
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

var clientCache map[string]*client
var cacheMutex sync.Mutex
