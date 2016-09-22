package register

import (
	"net/url"
	"time"
	"fmt"
	"net/http"
	"errors"
)

func BuildHealthChecks(urls []string) ([]HealthCheck, error) {
	healthChecks := make([]HealthCheck, len(urls))
	for i, u := range urls {
		parsedURL, err := url.Parse(u)
		if err != nil {
			return healthChecks, err
		}

		switch parsedURL.Scheme {
		case "http", "https":
			healthChecks[i] = HTTPHealthCheck{
				client: http.Client{},
			}
		default:
			return healthChecks, errors.New("Health check scheme is unsupported: " + u)
		}
	}
	return healthChecks, nil
}

type HealthCheck interface {
	Check(chan error, time.Duration)
}

type HTTPHealthCheck struct {
	url string
	client *http.Client
}

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