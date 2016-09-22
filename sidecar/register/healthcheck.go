package register

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// BuildHealthChecks constructs health checks from URLs.
func BuildHealthChecks(urls []string) ([]HealthCheck, error) {
	healthChecks := make([]HealthCheck, len(urls))
	for i, u := range urls {
		parsedURL, err := url.Parse(u)
		if err != nil {
			return healthChecks, err
		}

		switch parsedURL.Scheme {
		case "http", "https":
			healthChecks[i] = &HTTPHealthCheck{
				url:    u,
				client: &http.Client{},
			}
		default:
			return healthChecks, errors.New("Health check scheme is unsupported: " + u)
		}
	}
	return healthChecks, nil
}

// HealthCheck is an interface for performing a health check.
type HealthCheck interface {

	// Check performs a health check.
	Check(errChan chan error, timeout time.Duration)
}

// HTTPHealthCheck performs HTTP health checks.
type HTTPHealthCheck struct {
	url    string
	client *http.Client
}

// Check a HTTP URL with a GET. If the URL returns a response other than HTTP 200 or does not return a response before
// the timeout the check sends an error on errChan. If the health check is successful nil is sent.
func (c *HTTPHealthCheck) Check(errChan chan error, timeout time.Duration) {
	c.client.Timeout = timeout

	resp, err := c.client.Get(c.url)
	if err != nil {
		errChan <- err
		return
	}

	if resp.StatusCode != http.StatusOK {
		errChan <- fmt.Errorf("HTTP/HTTPS health check returned %v", resp.StatusCode)
		return
	}

	errChan <- nil
}
