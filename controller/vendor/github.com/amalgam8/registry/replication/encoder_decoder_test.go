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
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/amalgam8/registry/auth"
	"github.com/stretchr/testify/assert"
)

type replicatedMsgMockup struct {
	RepType int
	Payload []byte
}

type RegistryServiceInstanceMockup struct {
	ID          string
	ServiceName string
	Endpoint    *RegistryEndpointMockup
	TTL         time.Duration
}

type RegistryEndpointMockup struct {
	Host string
	Port uint32
}

var ev *sse

func TestMain(m *testing.M) {

	ev = newSSEEvent("101", "REP")
	os.Exit(m.Run())
}

func TestEncodingEvent(t *testing.T) {
	var expected = `id: 101` +
		`event: REP` +
		`data: {"Namespace":"ns1","Data":"eyJSZXBUeXBlIjoxLCJQYXlsb2FkIjoiZXlKSlJDSTZJakVpTENKVFpYSjJhV05sVG1GdFpTSTZJa05oYkdOc0lpd2lSVzVrY0c5cGJuUWlPbnNpU0c5emRDSTZJakU1TWk0eE5DNHhOUzR4TmlJc0lsQnZjblFpT2pNek16TjlMQ0pVVkV3aU9qQjkifQ=="}`

	r, w := io.Pipe()

	enc := newEncoder(w)
	assert.NotNil(t, enc)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer w.Close()

		err := enc.Encode(ev)
		assert.Nil(t, err)
	}()

	go func() {
		defer wg.Done()
		defer r.Close()

		scanner := bufio.NewScanner(r)

		var buffer bytes.Buffer
		for scanner.Scan() {
			_, err := buffer.WriteString(scanner.Text())
			assert.NoError(t, err)

		}

		assert.Equal(t, expected, buffer.String())
	}()

	wg.Wait()
}

func TestSingleEncodingDecodeEvent(t *testing.T) {
	r, w := io.Pipe()

	enc := newEncoder(w)
	assert.NotNil(t, enc)

	dec := newDecoder(r)
	assert.NotNil(t, dec)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer w.Close()

		err := enc.Encode(ev)
		assert.Nil(t, err)
	}()

	go func() {
		defer wg.Done()
		defer r.Close()

		inputEv, err := dec.Decode()
		assert.NoError(t, err)
		assert.Equal(t, ev, inputEv)
	}()

	wg.Wait()
}

func TestMultipleEncodingDecodeEvent(t *testing.T) {
	cases := []struct {
		ev *sse
	}{
		{ev},
		{newSSEEvent("102", "DUP")},
		{newSSEEvent("103", "FUN")},
	}

	r, w := io.Pipe()

	enc := newEncoder(w)
	assert.NotNil(t, enc)

	dec := newDecoder(r)
	assert.NotNil(t, dec)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer w.Close()

		for _, tc := range cases {
			err := enc.Encode(tc.ev)
			assert.Nil(t, err)
		}
	}()

	go func() {
		defer wg.Done()
		defer r.Close()

		for {
			inputEv, err := dec.Decode()
			if err == io.EOF {
				break
			}

			exists := false
			for _, tc := range cases {
				exists = exists || assert.ObjectsAreEqual(tc.ev, inputEv)
			}
			assert.NoError(t, err)
			assert.True(t, exists)
		}
	}()

	wg.Wait()

}

//
// Utility Functions
//

func newServiceInstance(id string, name string, host string, port uint32) *RegistryServiceInstanceMockup {
	return &RegistryServiceInstanceMockup{
		ID:          id,
		ServiceName: name,
		Endpoint: &RegistryEndpointMockup{
			Host: host,
			Port: port,
		},
	}
}

func newSSEEvent(id, event string) *sse {
	// Prepare Registry/Catalog data message
	si := newServiceInstance("1", "Calcl", "192.14.15.16", 3333)
	payload, _ := json.Marshal(si)
	catalogMsg, _ := json.Marshal(&replicatedMsgMockup{RepType: 1, Payload: payload})
	// Prepare Replicator event
	ns := auth.NamespaceFrom("ns1")
	msg := &outMessage{Namespace: ns, Data: catalogMsg}
	data, _ := json.Marshal(msg)
	return &sse{id: id, event: event, data: string(data)}
}
