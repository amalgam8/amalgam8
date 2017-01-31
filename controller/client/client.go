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

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/registry/server/env"
)

const defaultTimeout = 30 * time.Second
const rulesPath = "/v1/rules"

// Config stores the configurable attributes of the client.
type Config struct {

	// URL of the controller server.
	URL string

	// AuthToken is the token to be used for authentication with the controller.
	// If left empty, no authentication is used.
	AuthToken string

	// HTTPClient can be used to customize the underlying HTTP client behavior,
	// such as enabling TLS, setting timeouts, etc.
	// If left nil, a default HTTP client will be used.
	HTTPClient *http.Client
}

// Client implements the Amalgam8 Controller API
type Client struct {
	url        string
	authToken  string
	httpClient *http.Client
	debug      bool
}

// New constructs a new A8 controller client.
func New(conf Config) (api.RulesService, error) {
	if err := normalizeConfig(&conf); err != nil {
		return nil, err
	}

	return &Client{
		url:        conf.URL,
		authToken:  conf.AuthToken,
		httpClient: conf.HTTPClient,
	}, nil
}

// NewClient constructs a new A8 API controller client.
// It returns the Client structure
func NewClient(conf Config) (*Client, error) {
	if err := normalizeConfig(&conf); err != nil {
		return nil, err
	}

	return &Client{
		url:        conf.URL,
		authToken:  conf.AuthToken,
		httpClient: conf.HTTPClient,
	}, nil
}

// normalizeConfig validates and sets defaults for the client configuration.
func normalizeConfig(conf *Config) error {
	u, err := url.Parse(conf.URL)
	if err != nil {
		return err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		// TODO: custom error type
		return fmt.Errorf("client: Unsupported protocol scheme %v", u.Scheme)
	}

	if conf.HTTPClient == nil {
		conf.HTTPClient = &http.Client{
			Timeout: defaultTimeout,
		}
	}

	return nil
}

// ListRules returns the list of rules
func (c *Client) ListRules(filter *api.RuleFilter) (*api.RulesSet, error) {
	var rulesSet api.RulesSet

	path := rulesPath
	query := c.filterToStringQuery(filter)

	if query != "" {
		path += query
	}

	if err := c.doRequest("GET", path, nil, nil, &rulesSet, http.StatusOK); err != nil {
		logrus.WithError(err).Warn("Failed to retrieve rules from controller")
		return nil, err
	}

	return &rulesSet, nil
}

// CreateRules creates the rules
func (c *Client) CreateRules(rulesSet *api.RulesSet) (interface{}, error) {
	return c.setRules("POST", rulesSet, http.StatusCreated)
}

// UpdateRules updates the rules
func (c *Client) UpdateRules(rulesSet *api.RulesSet) (interface{}, error) {
	return c.setRules("PUT", rulesSet, http.StatusOK)
}

// DeleteRules deletes the rules
func (c *Client) DeleteRules(filter *api.RuleFilter) ([]byte, error) {

	path := rulesPath
	query := c.filterToStringQuery(filter)

	if query != "" {
		path += query
	}

	if err := c.doRequest("DELETE", path, nil, nil, nil, http.StatusOK); err != nil {
		logrus.WithError(err).Warn("Failed to delete rules in controller")
		return nil, err
	}

	return []byte("Request Completed"), nil
}

// ListAction retuns the list of action rules
func (c *Client) ListAction(filter *api.RuleFilter) (*api.RulesByService, error) {
	return c.getRulesByType(rulesPath+"/actions", filter)
}

// ListRoutes return sthe list of route rules
func (c *Client) ListRoutes(filter *api.RuleFilter) (*api.RulesByService, error) {
	return c.getRulesByType(rulesPath+"/routes", filter)
}

// setRules do the request to set the rules in the controller
func (c *Client) setRules(method string, rulesSet *api.RulesSet, status int) (interface{}, error) {

	path := rulesPath
	result := &struct {
		IDs []string `json:"ids"`
	}{}

	if err := c.doRequest(method, path, rulesSet, nil, result, status); err != nil {
		logrus.WithError(err).Warn("Failed to set rules in controller")
		return nil, err
	}

	return result, nil
}

// getRulesByType do the request to obtain the rules from the controller
func (c *Client) getRulesByType(path string, filter *api.RuleFilter) (*api.RulesByService, error) {
	rules := &api.RulesByService{}

	query := c.filterToStringQuery(filter)

	if query != "" {
		path += query
	}

	if err := c.doRequest("GET", path, nil, nil, rules, http.StatusOK); err != nil {
		logrus.WithError(err).Warn("Failed to retrieve rules from controller")
		return nil, err
	}

	return rules, nil
}

// doRequest executes the HTTP request
func (c *Client) doRequest(method string, path string, body interface{}, headers http.Header, respObj interface{}, statusCode int) error {
	var reader io.Reader
	var curlBody string
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			logrus.WithError(err).Warn("error marshaling HTTP request body")
			return err
		}
		reader = bytes.NewBuffer(b)
		curlBody = string(b)
	}

	uri := c.url + path

	req, err := http.NewRequest(method, uri, reader)
	if err != nil {
		logrus.WithError(err).Warn("error creating HTTP request")
		return err
	}

	// Add authorization header
	if c.authToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))
	}

	req.Header.Set("Accept", "application/json")

	if method == "PUT" || method == "POST" {
		if body != nil {
			// Body exists, and encoded as JSON
			req.Header.Set("Content-Type", "application/json")
		} else {
			// No body, but the server needs to know that too
			req.Header.Set("Content-Length", "0")
		}
	}

	if c.debug {
		c.printCurl(method, path, curlBody, req.Header)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logrus.WithError(err).Warn("error performing HTTP request")
		return err
	}

	defer resp.Body.Close()
	requestID := resp.Header.Get(env.RequestID)

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Warn("error reading HTTP response body")
		return err
	}

	if resp.StatusCode != statusCode {
		message := string(respBody)
		logrus.WithError(err).WithFields(logrus.Fields{
			"error_code": resp.StatusCode,
			"message":    message,
			"request_id": requestID,
		}).Warn("HTTP error")

		respError := struct {
			StatusCode int
			Message    string
		}{
			StatusCode: resp.StatusCode,
			Message:    message,
		}

		return fmt.Errorf(fmt.Sprintf("%+v", respError))
	}

	if len(respBody) > 0 && respObj != nil {
		err = json.Unmarshal(respBody, respObj)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"request_id": requestID,
			}).Warn("Error parsing JSON response from controller")
			return err
		}
	}
	return nil
}

// Debug is used to set the debugging flag.
func (c *Client) Debug(debug bool) {
	c.debug = debug
}

// printCurl prints the curl of a given request
func (c *Client) printCurl(method, url string, data interface{}, headers http.Header) {
	var curl bytes.Buffer
	fmt.Fprint(&curl, "curl ")
	for k := range headers {
		fmt.Fprintf(&curl, "-H '%s: %s' ", k, headers.Get(k))
	}

	fmt.Fprintf(&curl, "-X %s '%s' ", method, c.url+url)

	dataString := fmt.Sprint(data)
	if len(dataString) > 0 && dataString != "<nil>" {
		fmt.Fprintf(&curl, "--data '%s'", dataString)
	}

	fmt.Println(curl.String())
}

// filterToStringQuery convert the filter into a string
func (c *Client) filterToStringQuery(filter *api.RuleFilter) string {
	u := url.URL{}

	query := make(url.Values)
	for _, id := range filter.IDs {
		query.Add("id", id)
	}

	for _, dest := range filter.Destinations {
		query.Add("destination", dest)
	}

	for _, tag := range filter.Tags {
		query.Add("tag", tag)
	}

	u.RawQuery = query.Encode()
	return u.String()
}
