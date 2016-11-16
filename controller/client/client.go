package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"time"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/controller/rules"
	"github.com/amalgam8/amalgam8/registry/server/env"
)

const defaultTimeout = 30 * time.Second

// RuleResponse is the information returned from a rule query.
type RuleResponse struct {
	// Rules that matched the filter.
	Rules []rules.Rule `json:"rules"`

	// Revision of the rules for this namespace.
	Revision int64 `json:"revision"`
}

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

// Client for the controller.
type Client interface {
	// GetRules returns the rules for this namespace that match the filter.
	GetRules(f rules.Filter) (RuleResponse, error)
}

// New constructs a new controller client.
func New(conf Config) (Client, error) {
	if err := normalizeConfig(&conf); err != nil {
		return nil, err
	}

	return &client{
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

type client struct {
	url        string
	authToken  string
	httpClient *http.Client
}

func (c *client) GetRules(filter rules.Filter) (RuleResponse, error) {
	var ruleResponse RuleResponse

	u, err := url.Parse(c.url + "/v1/rules")
	if err != nil {
		return ruleResponse, err
	}

	query := u.Query()
	for _, id := range filter.IDs {
		query.Add("id", id)
	}

	for _, tag := range filter.Tags {
		query.Add("tag", tag)
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		logrus.WithError(err).Warn("Error building request to get rules from controller")
		return ruleResponse, err
	}
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logrus.WithError(err).Warn("Failed to retrieve rules from controller")
		return ruleResponse, err
	}
	defer resp.Body.Close()

	requestID := resp.Header.Get(env.RequestID)

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"request_id": requestID,
		}).Warn("Error reading response from controller")
		return ruleResponse, err
	}

	if resp.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"request_id":  requestID,
			"body":        string(data),
		}).Warn("Controller returned unexpected response code")
		return ruleResponse, errors.New("client: received unexpected response code") // FIXME: custom error?
	}

	if err = json.Unmarshal(data, &ruleResponse); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"request_id": requestID,
		}).Warn("Error reading rules JSON from controller")
		return ruleResponse, err
	}

	return ruleResponse, nil
}

// setAuthHeader optionally sets an authorization header. If the token is empty we assume no authentication is enabled
// on the controller and do not add the header.
func (c *client) setAuthHeader(req *http.Request) {
	if c.authToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", c.authToken))
	}
}
