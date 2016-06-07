package register

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/amalgam8/sidecar/config"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

// DoServiceRegistrationAndHeartbeat perform service registration if requested
func DoServiceRegistrationAndHeartbeat(conf *config.Config, startThread bool) {
	var sd *SvcRegistryClient
	sd = NewSvcRegistryClient(conf.Registry.URL)
	//TODO make ttl and interval configurable
	sd.RegisterAndHeartbeat(conf.ServiceName, conf.EndpointHost, conf.EndpointPort, conf.Registry.Token, 45, 30, startThread)
}

// // SvcRegistryClient Interface for communicating with Service Registry
// type SvcRegistryClient interface {
// 	RegisterService(regJSON *RegisteredService, token string, headers *headers.Headers) (string, error)
// 	SendHeartbeat(regJSON *RegisteredService, heartbeatURL, token string) (string, error)
// }

// SvcRegistryClient TODO
type SvcRegistryClient struct {
	url    string
	path   string
	client http.Client
}

// Endpoint TODO
type Endpoint struct {
	Type  string `json:"type" encrypt:"true"`
	Value string `json:"value" encrypt:"true"`
}

// MetaData TODO
type MetaData struct {
	Version string `json:"version"`
}

// Instance TODO
type Instance struct {
	Endpoint    Endpoint `json:"endpoint"`
	ServiceName string   `json:"service_name,omitempty"`
	MetaData    MetaData `json:"metadata"`
	// Also has TTL and last_heartbeat, but we don't use them
}

// ServiceInfo TODO
type ServiceInfo struct {
	ServiceName string     `json:"service_name" encrypt:"true"`
	Instances   []Instance `json:"instances"`
}

// Links TODO
type Links struct {
	Heartbeat string `json:"heartbeat"`
}

// Response JSON
type Response struct {
	Links Links `json:"links"`
}

// RegisteredService JSON
type RegisteredService struct {
	ServiceName string   `json:"service_name"`
	Endpoint    Endpoint `json:"endpoint"`
	TTL         int      `json:"ttl"`
	MetaData    MetaData `json:"metadata"`
}

// SvcRegistryError error
type SvcRegistryError struct {
	StatusCode int
	Content    string
}

// NewSvcRegistryClient new Registry client
func NewSvcRegistryClient(url string) *SvcRegistryClient {
	c := new(SvcRegistryClient)
	c.url = url
	c.path = "/api/v1/instances"
	c.client = http.Client{}
	return c
}

// RegisterService registers service instance
func (c *SvcRegistryClient) RegisterService(reqJSON *RegisteredService, token string) (string, error) {
	respJSON := Response{}

	var err error
	var reader io.Reader
	if reqJSON != nil {
		bodyBytes, err := json.Marshal(reqJSON)
		if err != nil {
			log.WithFields(log.Fields{
				"err":    err,
				"url":    c.url + c.path,
				"method": "POST",
			}).Warn("Error marshalling JSON body")
			return "", err
		}
		reader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest("POST", c.url+c.path, reader)
	if err != nil {
		log.WithFields(log.Fields{
			"err":    err,
			"url":    c.url + c.path,
			"method": "POST",
		}).Error("Error creating http request")
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("Content-type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"err":    err,
			"url":    c.url + c.path,
			"method": "POST",
		}).Error("Request failed")
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		log.WithFields(log.Fields{
			"err":           err,
			"url":           c.url + c.path,
			"method":        "POST",
			"request_id":    resp.Header.Get("request-id"),
			"sd_request_id": resp.Header.Get("sd-request-id"),
			"status_code":   resp.StatusCode,
			"expected_code": http.StatusCreated,
		}).Error("Returned code did not match expected")
		return "", NewSvcRegistryError(resp)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err":           err,
			"url":           c.url + c.path,
			"method":        "POST",
			"request_id":    resp.Header.Get("request-id"),
			"sd_request_id": resp.Header.Get("sd-request-id"),
		}).Error("Error reading response body")
		return "", err
	}
	err = json.Unmarshal(b, &respJSON)
	if err != nil {
		// shouldn't happen unless the tenantJSON doesn't match the struct returned
		return "", err
	}

	// Set our new heartbeat URL
	heartbeatURL := respJSON.Links.Heartbeat
	if strings.Contains(c.url, "http:") {
		heartbeatURL = strings.Replace(heartbeatURL, "https:/", "http:/", -1)
	} else if strings.Contains(c.url, "https:") {
		heartbeatURL = strings.Replace(heartbeatURL, "http:/", "https:/", -1)
	}

	log.WithFields(log.Fields{
		"heartbeat_url": heartbeatURL,
		"request_id":    resp.Header.Get("request-id"),
	}).Info("Successfully registered with Registry")
	return heartbeatURL, nil
}

// SendHeartbeat sends service instance heartbeat
func (c *SvcRegistryClient) SendHeartbeat(regJSON *RegisteredService, heartbeatURL, token string) (string, error) {
	// Heartbeat using our special heartbeat URL
	req, err := http.NewRequest("PUT", heartbeatURL, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Error creating heartbeat http request")
		return heartbeatURL, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Forwarded-Proto", "https")

	resp, err := c.client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Heartbeat to Registry failed")
		return heartbeatURL, fmt.Errorf("Heartbeat to Registry failed: %v", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return heartbeatURL, err
		}

		// Check if our service has become unregistered
		if strings.Contains(string(b), "Instance not found") {
			// Try to re-register the service
			log.Info("Heartbeat to Regsitry failed, attempting to re-register")
			newHeartbeat, err := c.RegisterService(regJSON, token)
			if err != nil {
				return heartbeatURL, err
			}
			heartbeatURL = newHeartbeat
		} else {
			return heartbeatURL, fmt.Errorf("Expected HTTP status 200 OK, got %v", resp.StatusCode)
		}
	}
	return heartbeatURL, nil
}

// NewSvcRegistryError new Registry error
func NewSvcRegistryError(resp *http.Response) error {
	e := new(SvcRegistryError)
	e.StatusCode = resp.StatusCode
	content, _ := ioutil.ReadAll(resp.Body)
	e.Content = string(content)

	return e
}

// Error formatted error from Service Discovery
func (e *SvcRegistryError) Error() string {
	return fmt.Sprintf("SvcRegistryError: status_code=%v content=%v", e.StatusCode, e.Content)
}

// LocalIP gets local IP of system
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

// RegisterAndHeartbeat registers and sends heartbeats
func (c *SvcRegistryClient) RegisterAndHeartbeat(service, endpointHost string, endpointPort int, sdToken string, ttl, interval int, startThread bool) {
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	ip := endpointHost
	if ip == "" {
		for {
			ip = LocalIP()
			if ip != "" {
				break
			} else {
				time.Sleep(5 * time.Second)
				log.Warn("Unable to obtain local IP, retrying . . .")
			}
		}
		log.Info("Retrieved local IP")
	}

	s := strings.Split(service, ":")
	serviceName := s[0]

	regJSON := &RegisteredService{
		ServiceName: serviceName,
		Endpoint: Endpoint{
			// TODO: This needs to change to user-defined option
			Type:  "http",
			Value: fmt.Sprintf("%v:%v", ip, endpointPort),
		},
		TTL: ttl,
	}

	if len(s) > 1 {
		regJSON.MetaData.Version = s[1]
	}

	heartbeatURL := ""
	for {
		var err error
		heartbeatURL, err = c.RegisterService(regJSON, sdToken)
		if err == nil {
			break
		} else {
			time.Sleep(5 * time.Second)
			log.Warn("Unable to register with Registry, retrying . . .")
		}
	}
	log.WithFields(log.Fields{
		"heartbeat_url": heartbeatURL,
	}).Info("Successfully registered with Registry")

	heartbeat := func() {
		var err error
		for tick := range ticker.C {
			log.Debug("Regsitry heartbeat at ", tick)
			heartbeatURL, err = c.SendHeartbeat(regJSON, heartbeatURL, sdToken)
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("Couldn't send heartbeat to Registry")
			}
		}
	}

	if startThread {
		go heartbeat()
	} else {
		heartbeat()
	}

}
