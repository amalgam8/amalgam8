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

package store

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/amalgam8/registry/auth"
	"github.com/amalgam8/registry/cluster"
	"github.com/amalgam8/registry/replication"
)

const (
	testProtocol = 1
)

type mockupReplicator struct {
	namespace auth.Namespace
	broadcast chan<- []byte
	sendChan  chan<- []byte
}

func (mrc *mockupReplicator) Broadcast(data []byte) error {
	return nil
}

func (mrc *mockupReplicator) Send(memberID cluster.MemberID, data []byte) error {
	return nil
}

type mockupReplication struct {
	// Receive channel for incoming messages used to notify external listeners (e.g. Registry)
	NotifyChannel chan *replication.InMessage

	// Sync channel for incoming sync requests
	SyncReqChannel chan chan []byte

	syncChan chan *replication.InMessage
}

func (mr *mockupReplication) GetReplicator(namespace auth.Namespace) (replication.Replicator, error) {
	return createMockupReplicator(namespace), nil
}
func (mr *mockupReplication) Notification() <-chan *replication.InMessage {
	return mr.NotifyChannel
}
func (mr *mockupReplication) Sync(waitTime time.Duration) <-chan *replication.InMessage {
	return mr.syncChan
}
func (mr *mockupReplication) SyncRequest() <-chan chan []byte {
	return mr.SyncReqChannel
}
func (mr *mockupReplication) Stop() {
}

func TestNewCatalogMap(t *testing.T) {

	r := New(nil)
	assert.NotNil(t, r)
}

func TestFactory(t *testing.T) {

	cm := New(nil)
	assert.NotNil(t, cm)
}

func TestNonEmptyCreationCatalogMap(t *testing.T) {
	cm := New(nil)
	catalog, err := cm.GetCatalog(auth.NamespaceFrom("ns1"))
	assert.NoError(t, err)
	assert.NotNil(t, catalog)
	assert.NotEmpty(t, catalog)

	otherCatalog, err2 := cm.GetCatalog(auth.NamespaceFrom("ns1"))
	assert.NoError(t, err2)
	assert.NotNil(t, otherCatalog)
	assert.NotEmpty(t, otherCatalog)
	assert.Equal(t, otherCatalog, catalog)
}

func TestGetCatalogMultipleRegistrations(t *testing.T) {
	cm := New(nil)
	catalog1, err1 := cm.GetCatalog(auth.NamespaceFrom("ns1"))
	assert.NoError(t, err1)
	assert.NotNil(t, catalog1)
	assert.NotEmpty(t, catalog1)

	catalog2, err2 := cm.GetCatalog(auth.NamespaceFrom("ns2"))
	assert.NoError(t, err2)
	assert.NotNil(t, catalog2)
	assert.NotEmpty(t, catalog2)

	catalog3, err3 := cm.GetCatalog(auth.NamespaceFrom("ns3"))
	assert.NoError(t, err3)
	assert.NotNil(t, catalog3)
	assert.NotEmpty(t, catalog3)

}

func TestRPCCreationOfDiffInstanceDiffCatalog(t *testing.T) {
	rep := createMockupReplication()
	close(rep.(*mockupReplication).syncChan) //Make sure no deadlock occur

	var conf = *DefaultConfig
	conf.Replication = rep
	cm := New(&conf)

	catalog, err := cm.GetCatalog(auth.NamespaceFrom("ns1"))
	assert.NoError(t, err)
	assert.NotNil(t, catalog)
	assert.NotEmpty(t, catalog)

	instance1 := newServiceInstance("Calc", "192.168.0.1", 9080)
	var err2 error
	instance1, err2 = catalog.Register(instance1)
	assert.NoError(t, err2)
	assert.NotNil(t, instance1)

	otherCatalog, err1 := cm.GetCatalog(auth.NamespaceFrom("ns1"))
	assert.NoError(t, err1)
	assert.NotNil(t, otherCatalog)
	assert.NotEmpty(t, otherCatalog)
	assert.Equal(t, otherCatalog, catalog)

	instance2 := newServiceInstance("Calc", "192.168.0.2", 9082)
	var err3 error
	instance2, err3 = otherCatalog.Register(instance2)
	assert.NoError(t, err3)
	assert.NotNil(t, instance2)

	instances, err4 := catalog.List("Calc", nil)
	assert.NoError(t, err4)
	assert.Len(t, instances, 2)
	assertContainsInstance(t, instances, instance1)
	assertContainsInstance(t, instances, instance2)
}

func TestIncomingReplication(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	rep := createMockupReplication()
	close(rep.(*mockupReplication).syncChan) //Make sure no deadlock occur

	ns := auth.NamespaceFrom("ns1")

	var conf = *DefaultConfig
	conf.Replication = rep
	cm := New(&conf)
	assert.NotContains(t, cm.(*catalogMap).catalogs, ns)

	inst := newServiceInstance("Calc1", "192.168.0.1", 9080)
	payload, _ := json.Marshal(inst)
	data, _ := json.Marshal(&replicatedMsg{RepType: REGISTER, Payload: payload})
	rep.(*mockupReplication).NotifyChannel <- &replication.InMessage{cluster.MemberID("192.1.1.3:6100"), ns, data}

	catalog, err := cm.GetCatalog(auth.NamespaceFrom("ns1"))

	// NOTICE, it may fail, since a race between the registry and the test...
	time.Sleep(time.Duration(5) * time.Second)

	assert.NoError(t, err)

	instances1, err1 := catalog.List("Calc1", nil)
	assert.NoError(t, err1)
	assert.Len(t, instances1, 1)
}

func createMockupReplication() replication.Replication {
	return &mockupReplication{
		NotifyChannel:  make(chan *replication.InMessage, 2),
		SyncReqChannel: make(chan chan []byte, 2),
		syncChan:       make(chan *replication.InMessage, 1),
	}
}

func createMockupReplicator(namespace auth.Namespace) replication.Replicator {
	return &mockupReplicator{
		namespace: namespace,
		broadcast: make(chan []byte, 2),
		sendChan:  make(chan []byte, 2),
	}
}
