package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"encoding/json"

	"github.com/amalgam8/registry/auth"
	"github.com/amalgam8/registry/cluster"
	"github.com/amalgam8/registry/replication"
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

func TestFactory(t *testing.T) {

	r := newInMemoryRegistry(nil, nil)
	assert.NotNil(t, r)
}

func TestNonEmptyCreationRegistry(t *testing.T) {
	r := newInMemoryRegistry(nil, nil).(*inMemoryRegistry)
	catalog, err := r.create(auth.NamespaceFrom("ns1"))
	assert.NoError(t, err)
	assert.NotNil(t, catalog)
	assert.NotEmpty(t, catalog)

	otherCatalog, err2 := r.GetCatalog(auth.NamespaceFrom("ns1"))
	assert.NoError(t, err2)
	assert.NotNil(t, otherCatalog)
	assert.NotEmpty(t, otherCatalog)
	assert.Equal(t, otherCatalog, catalog)
}

func TestGetCatalogMultipleRegistrations(t *testing.T) {
	r := newInMemoryRegistry(nil, nil)
	catalog1, err1 := r.GetCatalog(auth.NamespaceFrom("ns1"))
	assert.NoError(t, err1)
	assert.NotNil(t, catalog1)
	assert.NotEmpty(t, catalog1)

	catalog2, err2 := r.GetCatalog(auth.NamespaceFrom("ns2"))
	assert.NoError(t, err2)
	assert.NotNil(t, catalog2)
	assert.NotEmpty(t, catalog2)

	catalog3, err3 := r.GetCatalog(auth.NamespaceFrom("ns3"))
	assert.NoError(t, err3)
	assert.NotNil(t, catalog3)
	assert.NotEmpty(t, catalog3)

}

func TestRPCCreationOfDiffInstanceDiffCatalog(t *testing.T) {
	rep := createMockupReplication()
	close(rep.(*mockupReplication).syncChan) //Make sure no deadlock occur

	r := newInMemoryRegistry(nil, rep)

	catalog, err := r.GetCatalog(auth.NamespaceFrom("ns1"))
	assert.NoError(t, err)
	assert.NotNil(t, catalog)
	assert.NotEmpty(t, catalog)

	instance1 := newServiceInstance("Calc", "192.168.0.1", 9080)
	var err2 error
	instance1, err2 = catalog.Register(instance1)
	assert.NoError(t, err2)
	assert.NotNil(t, instance1)

	otherCatalog, err1 := r.GetCatalog(auth.NamespaceFrom("ns1"))
	assert.NoError(t, err1)
	assert.NotNil(t, otherCatalog)
	assert.NotEmpty(t, otherCatalog)
	assert.Equal(t, otherCatalog, catalog)

	instance2 := newServiceInstance("Calc", "192.168.0.2", 9082)
	var err3 error
	instance2, err3 = otherCatalog.Register(instance2)
	assert.NoError(t, err3)
	assert.NotNil(t, instance2)

	instances, err4 := catalog.List("Calc", protocolPredicate)
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

	r := newInMemoryRegistry(nil, rep)
	assert.NotContains(t, r.(*inMemoryRegistry).namespaces, ns)

	inst := newServiceInstance("Calc1", "192.168.0.1", 9080)
	payload, _ := json.Marshal(inst)
	data, _ := json.Marshal(&replicatedMsg{RepType: REGISTER, Payload: payload})
	rep.(*mockupReplication).NotifyChannel <- &replication.InMessage{cluster.MemberID("192.1.1.3:6100"), ns, data}

	catalog, err := r.GetCatalog(auth.NamespaceFrom("ns1"))

	// NOTICE, it may fail, since a race between the registry and the test...
	time.Sleep(time.Duration(5) * time.Second)

	assert.NoError(t, err)

	instances1, err1 := catalog.List("Calc1", protocolPredicate)
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
