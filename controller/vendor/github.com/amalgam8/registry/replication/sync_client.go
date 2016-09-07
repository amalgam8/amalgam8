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

	log "github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/registry/cluster"
)

type syncClient struct {
	selfID      cluster.MemberID
	member      cluster.Member
	httpclient  *http.Client
	evChan      chan<- *InMessage
	lastEventID string
	sync.Mutex
	logger *log.Entry
}

func newSyncClient(selfID cluster.MemberID, member cluster.Member, httpclient *http.Client, eventsChan chan<- *InMessage, logger *log.Entry) error {
	client := &syncClient{
		selfID:     selfID,
		member:     member,
		httpclient: httpclient,
		evChan:     eventsChan,
		logger:     logger.WithFields(log.Fields{"peer": member.ID()}),
	}

	client.logger.Info("Creating a sync client to peer")

	r, err := client.connect()
	if err != nil {
		return err
	}
	client.readEvents(r)
	return nil
}

func (client *syncClient) connect() (io.ReadCloser, error) {
	var resp *http.Response
	var req *http.Request
	var err error

	client.logger.Info("Connecting to peer")

	url := fmt.Sprintf("http://%s:%d/%s/%s", client.member.IP(), client.member.Port(), version, syncContext)
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
	return resp.Body, nil
}

func (client *syncClient) readEvents(r io.ReadCloser) {
	defer r.Close()

	dec := newDecoder(r)
	for {
		ev, err := dec.Decode()

		if err != nil {
			if err != io.EOF {
				client.logger.WithFields(log.Fields{
					"error":       err,
					"lastEventId": client.lastEventID,
				}).Warn("Failed to decode message")
			}
			break
		}
		inSSE := ev.(*sse)
		if len(inSSE.ID()) > 0 {
			client.lastEventID = inSSE.ID()
		}

		var m outMessage
		if err := json.Unmarshal([]byte(ev.Data()), &m); err != nil {
			client.logger.WithFields(log.Fields{
				"error":       err,
				"lastEventId": client.lastEventID,
			}).Errorf("Failed to unmarshal message \"%s\"", ev.Data())
			return
		}
		client.evChan <- &InMessage{client.member.ID(), m.Namespace, m.Data}
	}
	client.logger.Info("Synchronization with peer has completed")
}
