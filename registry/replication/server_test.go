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
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/cluster"
	"github.com/amalgam8/amalgam8/registry/utils/network"
	"github.com/stretchr/testify/assert"
)

func TestSameReplicatorRequest(t *testing.T) {

	var err error
	var repServer Replication
	var replicator Replicator

	cluster := createCluster()
	member := createMember(6100)
	repServer, err = New(&Config{
		Registrator: cluster.Registrator(member),
		Membership:  cluster.Membership()})
	assert.NotNil(t, repServer)
	assert.NoError(t, err)
	defer repServer.Stop()
	replicator, err = repServer.GetReplicator(auth.NamespaceFrom("ns1"))
	assert.NotNil(t, replicator)
	assert.NoError(t, err)

	replicator, err = repServer.GetReplicator(auth.NamespaceFrom("ns1"))
	assert.NotNil(t, replicator)
	assert.Error(t, err)
}

func TestPortInUse(t *testing.T) {
	cluster := createCluster()

	member1 := createMember(6101)
	repServer1, err := New(&Config{
		Registrator: cluster.Registrator(member1),
		Membership:  cluster.Membership()})
	assert.NotNil(t, repServer1)
	assert.NoError(t, err)
	defer repServer1.Stop()

	member2 := createMember(6101)
	repServer2, err := New(&Config{
		Registrator: cluster.Registrator(member2),
		Membership:  cluster.Membership()})
	assert.Nil(t, repServer2)
	assert.Error(t, err)
}

func TestBroadcastMsg(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	cluster := createCluster()

	member1 := createMember(6102)
	repServer1, err := New(&Config{
		Registrator: cluster.Registrator(member1),
		Membership:  cluster.Membership()})
	assert.NotNil(t, repServer1)
	assert.NoError(t, err)
	defer repServer1.Stop()
	<-repServer1.Sync(1)

	member2 := createMember(6202)
	repServer2, err := New(&Config{
		Registrator: cluster.Registrator(member2),
		Membership:  cluster.Membership()})
	assert.NotNil(t, repServer2)
	assert.NoError(t, err)
	defer repServer2.Stop()
	<-repServer2.Sync(1)

	member3 := createMember(6302)
	repServer3, err := New(&Config{
		Registrator: cluster.Registrator(member3),
		Membership:  cluster.Membership()})
	assert.NotNil(t, repServer3)
	assert.NoError(t, err)
	defer repServer3.Stop()
	<-repServer3.Sync(1)

	replicator1, err := repServer1.GetReplicator(auth.NamespaceFrom("ns1"))
	assert.NotNil(t, replicator1)
	assert.NoError(t, err)

	timeOut := time.Now().Add(time.Duration(15) * time.Second)
	for timeOut.After(time.Now()) {
		if len(cluster.Membership().Members()) >= 2 {
			break
		}
		time.Sleep(time.Duration(1000) * time.Millisecond)

	}
	nMbrs := len(cluster.Membership().Members())
	assert.True(t, nMbrs >= 2, "Number of members in the cluster (%d) is not as expected (>=2).", nMbrs)
	time.Sleep(time.Duration(500) * time.Millisecond)

	data := "Hello TestBroadcatSingleMsg"
	assert.NoError(t, replicator1.Broadcast([]byte(data)))

	for count2, count3 := 0, 0; count2 < 1 || count3 < 1; {
		select {
		case in2 := <-repServer2.Notification():
			assert.EqualValues(t, data, in2.Data)
			count2++
		case in3 := <-repServer3.Notification():
			assert.EqualValues(t, data, in3.Data)
			count3++
		case <-time.Tick(time.Duration(10) * time.Second):
			assert.Fail(t, fmt.Sprintf("Fail to receive broadcast message due to timeout <%d,%d>", count2, count3))
			count2 = 2
			count3 = 2
		}
	}
}

func TestSendMsg(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	cluster := createCluster()

	member1 := createMember(6103)
	repServer1, err := New(&Config{
		Registrator: cluster.Registrator(member1),
		Membership:  cluster.Membership()})
	assert.NotNil(t, repServer1)
	assert.NoError(t, err)
	defer repServer1.Stop()
	<-repServer1.Sync(1)

	member2 := createMember(6203)
	repServer2, err := New(&Config{
		Registrator: cluster.Registrator(member2),
		Membership:  cluster.Membership()})
	assert.NotNil(t, repServer2)
	assert.NoError(t, err)
	defer repServer2.Stop()
	<-repServer2.Sync(1)

	replicator2, err := repServer2.GetReplicator(auth.NamespaceFrom("ns1"))
	assert.NotNil(t, replicator2)
	assert.NoError(t, err)

	timeOut := time.Now().Add(time.Duration(15) * time.Second)
	for timeOut.After(time.Now()) {
		if len(cluster.Membership().Members()) >= 1 {
			break
		}
		time.Sleep(time.Duration(1000) * time.Millisecond)

	}
	nMbrs := len(cluster.Membership().Members())
	assert.True(t, nMbrs >= 1, "Number of members in the cluster (%d) is not as expected (>=1).", nMbrs)
	time.Sleep(time.Duration(500) * time.Millisecond)

	data := "Hello TestSendMsg"
	assert.NoError(t, replicator2.Send(member1.ID(), []byte(data)))

	select {
	case in1 := <-repServer1.Notification():
		assert.EqualValues(t, data, in1.Data)
	case <-time.Tick(time.Duration(10) * time.Second):
		assert.Fail(t, "Fail to receive unicast send message due to timeout")

	}
}

func TestSync(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	cluster := createCluster()

	member1 := createMember(6104)
	repServer1, err := New(&Config{
		Registrator: cluster.Registrator(member1),
		Membership:  cluster.Membership()})
	assert.NotNil(t, repServer1)
	assert.NoError(t, err)
	defer repServer1.Stop()
	<-repServer1.Sync(1)

	member2 := createMember(6204)
	repServer2, err := New(&Config{
		Registrator: cluster.Registrator(member2),
		Membership:  cluster.Membership()})
	assert.NotNil(t, repServer2)
	assert.NoError(t, err)
	defer repServer2.Stop()

	timeOut := time.Now().Add(time.Duration(15) * time.Second)
	for timeOut.After(time.Now()) {
		if len(cluster.Membership().Members()) >= 1 {
			break
		}
		time.Sleep(time.Duration(1000) * time.Millisecond)

	}
	nMbrs := len(cluster.Membership().Members())
	assert.True(t, nMbrs >= 1, "Number of members in the cluster (%d) is not as expected (>=1).", nMbrs)

	data := []byte("Sync Message")
	outMsg := &outMessage{Data: data}
	buff, _ := json.Marshal(outMsg)
	go func() {
		ch := <-repServer1.SyncRequest()
		ch <- buff
		close(ch)
	}()

	count := 0
	for in1 := range repServer2.Sync(time.Duration(5) * time.Second) {
		assert.EqualValues(t, data, in1.Data)
		count++
	}

	assert.Equal(t, 1, count)
}

func TestStopGracefully(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	cluster := createCluster()

	member1 := createMember(6105)
	repServer1, err := New(&Config{
		Registrator: cluster.Registrator(member1),
		Membership:  cluster.Membership()})
	assert.NotNil(t, repServer1)
	assert.NoError(t, err)
	<-repServer1.Sync(1)

	member2 := createMember(6205)
	repServer2, err := New(&Config{
		Registrator: cluster.Registrator(member2),
		Membership:  cluster.Membership()})
	assert.NotNil(t, repServer2)
	assert.NoError(t, err)
	<-repServer2.Sync(1)

	replicator2, err := repServer2.GetReplicator(auth.NamespaceFrom("ns1"))
	assert.NotNil(t, replicator2)
	assert.NoError(t, err)

	timeOut := time.Now().Add(time.Duration(15) * time.Second)
	for timeOut.After(time.Now()) {
		if len(cluster.Membership().Members()) >= 1 {
			break
		}
		time.Sleep(time.Duration(1000) * time.Millisecond)

	}
	nMbrs := len(cluster.Membership().Members())
	assert.True(t, nMbrs >= 1, "Number of members in the cluster (%d) is not as expected (>=1).", nMbrs)
	time.Sleep(time.Duration(500) * time.Millisecond)

	repServer1.Stop()
	repServer2.Stop()

	timeOut = time.Now().Add(time.Duration(15) * time.Second)
	for timeOut.After(time.Now()) {
		if len(cluster.Membership().Members()) == 0 {
			break
		}
		time.Sleep(time.Duration(1000) * time.Millisecond)

	}
	nMbrs = len(cluster.Membership().Members())
	assert.True(t, nMbrs == 0, "Number of members in the cluster (%d) is not as expected (==0).", nMbrs)
}

func TestGzippedReplication(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	cluster := createCluster()

	member1 := createMember(6106)
	repServer1, _ := New(&Config{
		Registrator: cluster.Registrator(member1),
		Membership:  cluster.Membership()})
	defer repServer1.Stop()
	<-repServer1.Sync(1)

	member2 := createMember(6206)
	repServer2, _ := New(&Config{
		Registrator: cluster.Registrator(member2),
		Membership:  cluster.Membership()})
	defer repServer2.Stop()
	<-repServer2.Sync(1)

	var resp *http.Response

	replicator2, _ := repServer2.GetReplicator(auth.NamespaceFrom("ns1"))

	client := http.Client{Transport: &http.Transport{MaxIdleConnsPerHost: 1}}

	var requestData bytes.Buffer
	for i := 0; i < 100; i++ {
		requestData.Write([]byte("9999999999999999999999999999999999"))
	}

	// Iterate over both sync and replication HTTP endpoints
	for _, serviceType := range []string{syncContext, repContext} {
		// Iterate over both content-encoding options- gzip and no zip (identity)
		for _, encoding := range []string{"gzip", "identity"} {
			var dec decoder
			var unZipper *gzip.Reader
			var respBuff bytes.Buffer

			// prepare event that triggers a replication or a sync
			if serviceType == repContext {
				time.AfterFunc(time.Duration(2)*time.Second, func() {
					replicator2.Broadcast(requestData.Bytes())
				})
			} else {
				// grab the channel that was created for us
				// insert an entity inside it and close the channel to signal sync finish
				outMsg := &outMessage{Data: requestData.Bytes()}
				buff, _ := json.Marshal(outMsg)
				time.AfterFunc(time.Duration(2)*time.Second, func() {
					ch := <-repServer2.SyncRequest()
					ch <- buff
					close(ch)
				})
			}

			url := fmt.Sprintf("http://%s:%d/%s/%s", network.GetPrivateIP(), 6206, version, serviceType)
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("Accept-Encoding", encoding)
			req.Header.Set(headerMemberID, "fake member id")
			resp, _ = client.Do(req)
			// wait 4 sec to read enough data, since ioutil.ReadAll blocks until EOF
			// and we use an indefinite connection in case of replication service endpoint
			if serviceType == repContext {
				time.AfterFunc(time.Duration(4)*time.Second, func() { resp.Body.Close() })
			}
			responseData, _ := ioutil.ReadAll(resp.Body)
			respBuff.Write(responseData)

			// if we've sent a gzipped entity, decode it before comparing
			// but also ensure that the returned payload was compressed
			// by comparing the size of the payload to the size sent
			if encoding == "gzip" {
				assert.True(t, len(requestData.Bytes()) > len(responseData),
					"Seems like the gzip hasn't shrunk the payload size enough or at all")
				unZipper, _ = gzip.NewReader(&respBuff)
				dec.Reader = bufio.NewReader(unZipper)
			} else {
				dec.Reader = bufio.NewReader(&respBuff)
			}

			event, _ := dec.Decode()
			if encoding == "gzip" {
				unZipper.Close()
			}
			var entityFromServer outMessage
			json.Unmarshal([]byte(event.Data()), &entityFromServer)
			assert.Equal(t, string(requestData.Bytes()), string(entityFromServer.Data),
				"Didn't receive the same content we sent in the case where encoding is %s", encoding)
		}
	}
}

func createCluster() cluster.Cluster {
	config := &cluster.Config{
		BackendType: cluster.MemoryBackend,
	}
	cl, _ := cluster.New(config)
	return cl
}

func createMember(port uint16) cluster.Member {
	return cluster.NewMember(network.GetPrivateIP(), port)
}
