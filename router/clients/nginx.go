package clients

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
)

// NGINX client interface for updates to lua
type NGINX interface {
	UpdateHTTPUpstreams(conf NGINXJson) error
}

// NGINXJson sent to update http/https endpoints
type NGINXJson struct {
	Upstreams map[string]NGINXUpstream `json:"upstreams"`
	Services  map[string]NGINXService  `json:"services"`
	Faults    []NGINXFault             `json:"faults,omitempty"`
}

// NGINXService version info for lua
type NGINXService struct {
	Default   string `json:"default"`
	Selectors string `json:"selectors,omitempty"`
	Type      string `json:"type"`
}

// NGINXUpstream server info for lua
type NGINXUpstream struct {
	Upstreams []NGINXEndpoint `json:"servers"`
}

// NGINXEndpoint for lua
type NGINXEndpoint struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// NGINXFault for representing fault injection for lua
type NGINXFault struct {
	Source           string  `json:"source"`
	Destination      string  `json:"destination"`
	Header           string  `json:"header"`
	Pattern          string  `json:"pattern"`
	Delay            float64 `json:"delay"`
	DelayProbability float64 `json:"delay_probability"`
	AbortProbability float64 `json:"abort_probability"`
	AbortCode        int     `json:"return_code"`
}

type nginx struct {
	httpClient *http.Client
	url        string
}

// NewNGINXClient return new NGINX client
func NewNGINXClient(url string) NGINX {
	return &nginx{
		httpClient: &http.Client{},
		url:        url,
	}
}

// UpdateHTTPUpstreams updates http upstreams in lua dynamically
func (n *nginx) UpdateHTTPUpstreams(conf NGINXJson) error {

	data, err := json.Marshal(&conf)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("Could not marshal request body")
		return err
	}

	reader := bytes.NewReader(data)
	req, err := http.NewRequest("POST", n.url+"/a8-admin", reader)
	if err != nil {
		logrus.WithError(err).Error("Failed building request to NGINX server")
		return err
	}

	resp, err := n.httpClient.Do(req)
	if err != nil {
		logrus.WithError(err).Error("Failed to send request to NGINX server")
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)

		logrus.WithFields(logrus.Fields{
			"err":         err,
			"body":        string(data),
			"status_code": resp.StatusCode,
		}).Error("POST to NGINX server return failure")
		return err
	}

	return nil
}
