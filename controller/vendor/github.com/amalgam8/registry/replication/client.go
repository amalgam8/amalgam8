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

package replication

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/registry/cluster"
)

// clientConnection encapsulates client-side connection with a peer-member replication server.
type clientConnection interface {

	// IsConnected returns whether this client is connected to a peer member.
	isConnected() bool

	// Connect establishes a connection with the peer member replication server, and returns a ready-to-use io.ReadCloser.
	// If conenction could not be established, a non-nil error is returned.
	connect() (io.ReadCloser, error)

	// Close disconnect from the peer member replication server.
	close()
}

type client struct {
	selfID      cluster.MemberID
	member      cluster.Member
	httpclient  *http.Client
	evChan      chan<- *InMessage
	lastEventID string
	retry       time.Duration
	connected   bool
	keep        bool
	body        io.ReadCloser
	sync.Mutex
	logger *log.Entry
}

func newClient(selfID cluster.MemberID, member cluster.Member, httpclient *http.Client, eventsChan chan<- *InMessage, logger *log.Entry) (*client, error) {
	client := &client{
		selfID:     selfID,
		member:     member,
		httpclient: httpclient,
		evChan:     eventsChan,
		retry:      (time.Millisecond * 3000),
		connected:  false,
		keep:       true,
		logger:     logger.WithFields(log.Fields{"peer": member.ID()}),
	}

	client.logger.Info("Creating a replication client to peer")
	go client.connectAndReadEvents(client.retry)

	return client, nil
}

func (client *client) connect() (io.ReadCloser, error) {
	var resp *http.Response
	var req *http.Request
	var err error

	client.logger.Info("Connecting to peer")
	client.setConnected(false)

	url := fmt.Sprintf("http://%s:%d/%s/%s", client.member.IP(), client.member.Port(), version, repContext)
	if req, err = http.NewRequest("GET", url, nil); err != nil {
		client.logger.WithFields(log.Fields{
			"error": err,
		}).Warn("Failed to connect to peer")

		return nil, err
	}
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/event-stream")
	if len(client.lastEventID) > 0 {
		req.Header.Set("Last-Event-ID", client.lastEventID)
	}
	req.Header.Set(headerMemberID, string(client.selfID))

	if resp, err = client.httpclient.Do(req); err != nil {
		client.logger.WithFields(log.Fields{
			"error": err,
		}).Warn("Failed to connect to peer")

		return nil, err
	}
	if resp.StatusCode != 200 {
		message, _ := ioutil.ReadAll(resp.Body)
		client.logger.WithFields(log.Fields{
			"error": message,
		}).Warn("Failed to connect to peer")

		return nil, fmt.Errorf("Code:%d , Msg:%s", resp.StatusCode, string(message))
	}
	client.setConnected(true)
	client.Lock()
	defer client.Unlock()
	client.body = resp.Body
	return resp.Body, nil
}

func (client *client) readEvents(r io.ReadCloser) {
	defer r.Close()

	client.logger.Info("Start reading events from peer")
	dec := newDecoder(r)
	for client.goOn() {
		ev, err := dec.Decode()

		if err != nil {
			if err != io.EOF && client.goOn() {
				client.logger.WithFields(log.Fields{
					"error": err,
				}).Warn("Failed to decode message")
			}
			break
		}
		inSSE := ev.(*sse)
		if inSSE.Retry() > 0 {
			client.retry = time.Duration(inSSE.Retry()) * time.Millisecond
		}
		if len(inSSE.ID()) > 0 {
			client.lastEventID = inSSE.ID()
		}

		var m outMessage
		if err := json.Unmarshal([]byte(ev.Data()), &m); err != nil {
			client.logger.WithFields(log.Fields{
				"error": err,
			}).Errorf("Failed to unmarshal message \"%s\"", ev.Data())
			break
		}
		client.evChan <- &InMessage{client.member.ID(), m.Namespace, m.Data}
	}

	client.setConnected(false)

	// NOTE: because of the defer we're opening the new connection
	// before closing the old one. Shouldn't be a problem in practice,
	// but something to be aware of.
	if client.goOn() {
		go client.connectAndReadEvents(client.retry)
	}

}

func (client *client) connectAndReadEvents(retryDelay time.Duration) {

	for {

		next, err := client.connect()
		if err == nil {
			client.readEvents(next)
			return
		}

		client.logger.Infof("Reconnecting to peer in %0.4fs", retryDelay.Seconds())
		time.Sleep(retryDelay)
		if !client.goOn() {
			client.logger.Info("Peer connection attempts has stopped")
			return
		}

		retryDelay *= 2
	}

}

func (client *client) close() {
	client.Lock()
	defer client.Unlock()

	client.logger.Info("Closing peer connection")
	client.keep = false
	if client.body != nil {
		if err := client.body.Close(); err != nil {
			client.logger.WithFields(log.Fields{
				"error": err,
			}).Warn("Failed to close response body")
		}
	}

	// Avoid calling setConnected() as the mutex is already locked
	client.connected = false
}

func (client *client) goOn() bool {
	client.Lock()
	defer client.Unlock()

	return client.keep
}

// IsConnected sets whether this client is connected to a peer member.
func (client *client) setConnected(connected bool) {
	client.Lock()
	defer client.Unlock()

	client.connected = connected
}

// IsConnected returns whether this client is connected to a peer member.
func (client *client) isConnected() bool {
	client.Lock()
	defer client.Unlock()

	return client.connected
}
