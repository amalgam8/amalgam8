package clients

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/resources"
)

type NGINX interface {
	UpdateHttpUpstreams(conf resources.NGINXJson) error
}

type nginx struct {
	httpClient *http.Client
	url        string
}

func NewNGINXClient(url string) NGINX {
	return &nginx{
		httpClient: &http.Client{},
		url:        url,
	}
}

func (n *nginx) UpdateHttpUpstreams(conf resources.NGINXJson) error {

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
