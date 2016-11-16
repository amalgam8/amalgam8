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
	"time"
)

// Client .
type Client interface {
	GET(apiURL string, debug bool, extraHeaders http.Header, result interface{}) error
	POST(apiURL string, body io.Reader, debug bool, headers http.Header, result interface{}) error
	PUT(apiURL string, body io.Reader, debug bool, headers http.Header, result interface{}) error
	DELETE(apiURL string, debug bool, extraHeaders http.Header, result interface{}) error
	NewHeader() http.Header
	SetHTTPClient(client *http.Client)
	SetURL(baseURL string)
	SetToken(token string)
}

// A8client .
type A8client struct {
	url        string
	token      string
	httpClient *http.Client
}

// NewClient .
func NewClient(url, token string, HTTPclient *http.Client) Client {
	// If not http client has been defined, use the default HTTP client
	if HTTPclient == nil {
		HTTPclient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &A8client{
		url:        url,
		token:      token,
		httpClient: HTTPclient,
	}
}

// SetURL .
func (c *A8client) SetURL(baseURL string) {
	c.url = baseURL
}

// SetToken .
func (c *A8client) SetToken(token string) {
	c.token = token
}

// SetHTTPClient .
func (c *A8client) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

// NewHeader .
func (c *A8client) NewHeader() http.Header {
	return http.Header{}
}

// GET .
func (c *A8client) GET(apiURL string, debug bool, headers http.Header, result interface{}) error {
	method := "GET"
	req, err := c.BuildRequest(method, apiURL, nil, headers)
	if err != nil {
		return err
	}

	if debug {
		c.printCurl(method, apiURL, nil, req.Header)
	}

	err = c.Do(req, &result)
	if err != nil {
		return err
	}

	return nil
}

// POST .
func (c *A8client) POST(apiURL string, body io.Reader, debug bool, headers http.Header, result interface{}) error {
	method := "POST"
	body, data := copyBody(body)
	req, err := c.BuildRequest(method, apiURL, body, headers)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	if debug {
		c.printCurl(method, apiURL, data, req.Header)
	}

	err = c.Do(req, &result)
	if err != nil {
		return err
	}

	return nil
}

// PUT .
func (c *A8client) PUT(apiURL string, body io.Reader, debug bool, headers http.Header, result interface{}) error {
	method := "PUT"
	body, data := copyBody(body)
	req, err := c.BuildRequest(method, apiURL, body, headers)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	if debug {
		c.printCurl(method, apiURL, data, req.Header)
	}

	err = c.Do(req, &result)
	if err != nil {
		return err
	}

	return nil
}

// DELETE .
func (c *A8client) DELETE(apiURL string, debug bool, headers http.Header, result interface{}) error {
	method := "DELETE"
	req, err := c.BuildRequest(method, apiURL, nil, headers)
	if err != nil {
		return err
	}

	if debug {
		c.printCurl(method, apiURL, nil, req.Header)
	}

	err = c.Do(req, &result)
	if err != nil {
		return err
	}

	return nil
}

// Do returns the response of a http request
func (c *A8client) Do(req *http.Request, result interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	err = parseResponse(resp, &result)
	if err != nil {
		return err
	}

	return nil
}

// BuildRequest builds a http request
func (c *A8client) BuildRequest(method, apiURL string, body io.Reader, headers http.Header) (*http.Request, error) {

	u, err := url.ParseRequestURI(c.url + apiURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	if len(c.token) != 0 {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", c.token))
	}

	req.Header.Set("Accept", "application/json")
	for k := range headers {
		req.Header.Set(k, headers.Get(k))
	}

	return req, nil
}

// copyBody .
func copyBody(body io.Reader) (io.Reader, []byte) {
	if body == nil {
		return nil, nil
	}

	// Clone body
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return bytes.NewReader(data), nil
	}

	return bytes.NewReader(data), data
}

// printCurl .
func (c *A8client) printCurl(method, url string, data []byte, headers http.Header) {
	var curl bytes.Buffer
	fmt.Fprint(&curl, "curl ")
	for k := range headers {
		fmt.Fprintf(&curl, "-H '%s: %s' ", k, headers.Get(k))
	}

	fmt.Fprintf(&curl, "-X %s '%s' ", method, c.url+url)
	if len(data) != 0 {
		fmt.Fprintf(&curl, "--data '%s'", data)
	}

	fmt.Println(curl.String())
}

// parseResponse .
func parseResponse(resp *http.Response, dest interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return clientError{
			StatusCode: resp.StatusCode,
			Body:       string(body),
		}
	}

	if len(body) > 0 && dest != nil {
		err = json.Unmarshal(body, dest)
		if err != nil {
			return err
		}
	}

	return nil
}
