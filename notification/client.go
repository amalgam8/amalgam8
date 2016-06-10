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

package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
)

// Topic for Message Hub.
type Topic struct {
	Name              string `json:"name"`
	MarkedForDeletion bool   `json:"markedForDeletion"`
}

// MessageHubClient for the Message Hub REST API.
type MessageHubClient interface {
	Topics() ([]string, error)
	CreateTopic(topic string, partitions int) error
	DeleteTopic(topic string) error
}

type messageHubClient struct {
	client   http.Client
	clientID string
	url      string
}

// NewMessageHubClient creates new instance
func NewMessageHubClient(clientID, url string) MessageHubClient {
	return &messageHubClient{
		client:   http.Client{},
		clientID: clientID,
		url:      url,
	}
}

// Topics lists topics.
func (c *messageHubClient) Topics() ([]string, error) {
	var topicsJSON []Topic
	var topics []string

	req, err := http.NewRequest("GET", c.url+"/admin/topics", nil)
	if err != nil {
		return topics, err
	}

	req.Header.Set("X-Auth-Token", c.clientID)

	resp, err := c.client.Do(req)
	if err != nil {
		return topics, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return topics, fmt.Errorf("Non-200 response returned getting topics: %v", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return topics, err
	}

	err = json.Unmarshal(bodyBytes, &topicsJSON)
	if err != nil {
		return topics, err
	}

	for _, topic := range topicsJSON {
		topics = append(topics, topic.Name)
	}

	logrus.WithFields(logrus.Fields{
		"topics": topics,
	}).Info("Obtained topics from Messagehub")

	return topics, nil
}

// CreateTopic creates a topic.
func (c *messageHubClient) CreateTopic(topic string, partitions int) error {
	logrus.WithFields(logrus.Fields{
		"topic":      topic,
		"partitions": partitions,
	}).Info("Creating topic")

	body := struct {
		Name       string `json:"name"`
		Partitions int    `json:"partitions"`
	}{
		Name:       topic,
		Partitions: partitions,
	}

	bodyBytes, err := json.Marshal(&body)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(bodyBytes)

	req, err := http.NewRequest("POST", c.url+"/admin/topics", reader)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", c.clientID)
	req.Header.Set("Content-type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Non-202 status code returned: %v", resp.StatusCode)
	}

	return nil
}

// DeleteTopic deletes a topic.
func (c *messageHubClient) DeleteTopic(topic string) error {
	logrus.WithFields(logrus.Fields{
		"topic": topic,
	}).Info("Deleting topic")

	req, err := http.NewRequest("DELETE", c.url+"/admin/topics/"+topic, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", c.clientID)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Non-202 status code returned: %v", resp.StatusCode)
	}

	return nil
}
