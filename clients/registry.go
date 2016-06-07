package clients

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sort"

	log "github.com/Sirupsen/logrus"
)

// SDRequestID trace ID used by the registry
// TODO: does this request ID still exist, and is it still "sd"?
const SDRequestID = "sd-request-id"

// Registry client
type Registry interface {
	GetInstances(token string, url string) ([]Instance, error)
	GetToken(username, password, org, space string, url string) (string, error)
	GetServices(token string, url string) ([]string, error)
	GetService(name, token string, url string) (ServiceInfo, error)
	CheckUptime(url string) error
}

type registry struct {
	client http.Client
}

// NewRegistry client
func NewRegistry() Registry {
	return &registry{
		client: http.Client{},
	}
}

// GetInstances gets all instance registered for a given token
func (c *registry) GetInstances(token string, url string) ([]Instance, error) {
	registeredServ := struct {
		Instances []Instance `json:"instances"`
	}{}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)
	header.Set("X-Forwarded-Proto", "https")

	err := c.doRequest("GET", url, "/api/v1/instances", http.StatusOK, nil, &registeredServ, header)
	if err != nil {
		return nil, err
	}

	return registeredServ.Instances, nil
}

// GetToken generates a new token for the given organization and space
func (c *registry) GetToken(username, password, org, space string, url string) (string, error) {
	reqJSON := struct {
		Org   string `json:"org"`
		Space string `json:"space"`
	}{
		Org:   org,
		Space: space,
	}

	respJSON := struct {
		Token string `json:"token"`
	}{}

	authBytes := []byte(fmt.Sprintf("%v:%v", username, password))
	encodedAuth := base64.StdEncoding.EncodeToString(authBytes)

	header := http.Header{}
	header.Set("Authorization", "Basic "+encodedAuth)
	header.Set("Content-type", "application/json")

	err := c.doRequest("POST", url, "/api/v1/token", http.StatusOK, &reqJSON, &respJSON, header)
	if err != nil {
		return respJSON.Token, err
	}

	return respJSON.Token, nil

}

// GetServices provides all the services registered to the org and space of the given SD token
func (c *registry) GetServices(token string, url string) ([]string, error) {
	registeredServ := struct {
		Services []string `json:"services"`
	}{}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)
	header.Set("X-Forwarded-Proto", "https")

	err := c.doRequest("GET", url, "/api/v1/services", http.StatusOK, nil, &registeredServ, header)
	if err != nil {
		return nil, err
	}

	return registeredServ.Services, nil
}

// GetService TODO
func (c *registry) GetService(name, token string, url string) (ServiceInfo, error) {
	var serviceInfo ServiceInfo
	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)
	header.Set("X-Forwarded-Proto", "https")
	err := c.doRequest("GET", url, fmt.Sprintf("/api/v1/services/%v", name), http.StatusOK, nil, &serviceInfo, header)
	if err != nil {
		return serviceInfo, err
	}

	sort.Sort(ByInstance(serviceInfo.Instances))
	return serviceInfo, nil
}

// CheckUptime TODO
func (c *registry) CheckUptime(url string) error {
	err := c.doRequest("GET", url, "/uptime", http.StatusOK, nil, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// DoRequest TODO
func (c *registry) doRequest(method, url, path string, desiredCode int, reqJSON, respJSON interface{}, header http.Header) error {

	var err error
	var reader io.Reader
	if reqJSON != nil {
		bodyBytes, err := json.Marshal(reqJSON)
		if err != nil {
			log.WithFields(log.Fields{
				"err":    err,
				"url":    url + path,
				"method": method,
			}).Warn("Error marshalling JSON body")
			return err
		}

		reader = bytes.NewReader(bodyBytes)
	}
	req, err := http.NewRequest(method, url+path, reader)
	if err != nil {
		log.WithFields(log.Fields{
			"err":    err,
			"url":    url + path,
			"method": method,
		}).Error("Error creating http request")
		return err
	}
	if header != nil {
		req.Header = header
	}

	resp, err := c.client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"err":    err,
			"url":    url + path,
			"method": method,
		}).Error("Request failed")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != desiredCode {
		log.WithFields(log.Fields{
			"err":           err,
			"url":           url + path,
			"method":        method,
			"sd_request_id": resp.Header.Get(SDRequestID),
			"status_code":   resp.StatusCode,
			"expected_code": desiredCode,
		}).Error("Returned code did not match expected")
		return NewRegistryError(resp)
	}

	if respJSON != nil {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.WithFields(log.Fields{
				"err":           err,
				"url":           url + path,
				"method":        method,
				"sd_request_id": resp.Header.Get(SDRequestID),
			}).Error("Error reading response body")
			return err
		}

		if err = json.Unmarshal(b, respJSON); err != nil {
			log.WithFields(log.Fields{
				"err":           err,
				"url":           url + path,
				"method":        method,
				"sd_request_id": resp.Header.Get(SDRequestID),
			}).Error("Response body did not unmarshal into provided interface")
			return err
		}
	}

	return nil
}

// RegistryError TODO
type RegistryError struct {
	StatusCode int
	Content    string
}

// NewRegistryError TODO
func NewRegistryError(resp *http.Response) error {
	e := new(RegistryError)
	e.StatusCode = resp.StatusCode
	content, _ := ioutil.ReadAll(resp.Body)
	e.Content = string(content)

	return e
}

// Error TODO
func (e *RegistryError) Error() string {
	return fmt.Sprintf("RegistryError: status_code=%v content=%v", e.StatusCode, e.Content)
}

// LocalIP TODO
func LocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
