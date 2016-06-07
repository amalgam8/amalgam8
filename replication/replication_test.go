package replication_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/amalgam8/registry/api"
	"github.com/amalgam8/registry/api/protocol/amalgam8"
	"github.com/amalgam8/registry/cluster"
	"github.com/amalgam8/registry/replication"
	"github.com/amalgam8/registry/store"
	"github.com/amalgam8/registry/utils/network"
)

const (
	defaultSyncWaitTime = 10
)

var basePort uint16 = 5080

func setupServer(t *testing.T, rest_port, rep_port uint16, cl cluster.Cluster) (replication.Replication, api.Server) {
	var rep replication.Replication
	var server api.Server
	networkAvailable := network.WaitForPrivateNetwork()
	assert.NotNil(t, networkAvailable)

	// Configure and create the cluster module
	if cl == nil {
		return nil, nil
	}
	var err error

	// Configure and create the replication module
	self := cluster.NewMember(network.GetPrivateIP(), rep_port)
	repConfig := &replication.Config{
		Membership:  cl.Membership(),
		Registrator: cl.Registrator(self),
	}
	rep, err = replication.New(repConfig)
	assert.NoError(t, err)
	assert.NotNil(t, rep)

	regConfig := &store.Config{
		DefaultTTL:        time.Duration(30) * time.Second,
		MinimumTTL:        time.Duration(10) * time.Second,
		MaximumTTL:        time.Duration(600) * time.Second,
		SyncWaitTime:      time.Duration(defaultSyncWaitTime) * time.Second,
		NamespaceCapacity: 50,
	}

	reg := store.New(regConfig, rep)
	server, err = api.NewServer(
		&api.Config{
			HTTPAddressSpec: fmt.Sprintf(":%d", rest_port),
			Registry:        reg,
		},
	)

	assert.NoError(t, err)
	assert.NotNil(t, server)

	// Start the API server, and "wait" for it to bind
	go server.Start()
	time.Sleep(100 * time.Millisecond)

	return rep, server
}

var instances = []amalgam8.InstanceRegistration{
	{ServiceName: "http-1", Endpoint: &amalgam8.InstanceAddress{Value: "192.168.0.1:80", Type: "tcp"}, TTL: 60},
	{ServiceName: "http-2", Endpoint: &amalgam8.InstanceAddress{Value: "192.168.0.2:81", Type: "tcp"}, TTL: 60},
	{ServiceName: "http-3", Endpoint: &amalgam8.InstanceAddress{Value: "192.168.0.3:82", Type: "tcp"}, TTL: 60},
	{ServiceName: "http-4", Endpoint: &amalgam8.InstanceAddress{Value: "192.168.0.4:83", Type: "tcp"}, TTL: 60},
	{ServiceName: "http-5", Endpoint: &amalgam8.InstanceAddress{Value: "192.168.0.5:84", Type: "tcp"}, TTL: 60},
}

func register(t *testing.T, port uint16) []string {
	ids := make([]string, len(instances))

	url := fmt.Sprintf("http://localhost:%d%s", port, amalgam8.InstanceCreateURL())
	for index, inst := range instances {
		b, err := json.Marshal(inst)
		assert.NoError(t, err)

		req, err1 := http.NewRequest("POST", url, bytes.NewReader(b))
		assert.NoError(t, err1)
		assert.NotNil(t, req)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-Proto", "https")

		res, err2 := http.DefaultClient.Do(req)
		assert.NoError(t, err2)
		assert.NotNil(t, res)
		assert.EqualValues(t, 201, res.StatusCode)

		data, err3 := ioutil.ReadAll(res.Body)
		assert.NoError(t, err3)
		res.Body.Close()

		var si amalgam8.ServiceInstance
		err = json.Unmarshal(data, &si)
		assert.NoError(t, err)
		ids[index] = si.ID
	}

	return ids
}

func renew(t *testing.T, port uint16, id string) {
	url := fmt.Sprintf("http://localhost:%d%s", port, amalgam8.InstanceHeartbeatURL(id))
	req, err1 := http.NewRequest("PUT", url, nil)
	assert.NoError(t, err1)
	assert.NotNil(t, req)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-Proto", "https")

	res, err2 := http.DefaultClient.Do(req)
	assert.NoError(t, err2)
	assert.NotNil(t, res)
	assert.EqualValues(t, 200, res.StatusCode)
	res.Body.Close()
}

func lookup(t *testing.T, port uint16, sname string) int {
	url := fmt.Sprintf("http://localhost:%d%s", port, amalgam8.ServiceInstancesURL(sname))
	req, err1 := http.NewRequest("GET", url, nil)
	assert.NoError(t, err1)
	assert.NotNil(t, req)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-Proto", "https")

	res, err2 := http.DefaultClient.Do(req)
	assert.NoError(t, err2)
	assert.NotNil(t, res)
	assert.EqualValues(t, 200, res.StatusCode)

	data, err3 := ioutil.ReadAll(res.Body)
	assert.NoError(t, err3)
	res.Body.Close()

	var slist amalgam8.InstanceList
	err1 = json.Unmarshal(data, &slist)
	assert.NoError(t, err1)
	return len(slist.Instances)
}

func list(t *testing.T, port uint16) int {
	url := fmt.Sprintf("http://localhost:%d%s", port, amalgam8.ServiceNamesURL())
	req, err1 := http.NewRequest("GET", url, nil)
	assert.NoError(t, err1)
	assert.NotNil(t, req)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-Proto", "https")

	res, err2 := http.DefaultClient.Do(req)
	assert.NoError(t, err2)
	assert.NotNil(t, res)
	assert.EqualValues(t, 200, res.StatusCode)

	data, err3 := ioutil.ReadAll(res.Body)
	assert.NoError(t, err3)
	res.Body.Close()

	var slist amalgam8.ServicesList
	err1 = json.Unmarshal(data, &slist)
	assert.NoError(t, err1)
	return len(slist.Services)
}

func TestStartup(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	port1 := nextPort()

	// Configure and create the cluster module
	cfConfig := &cluster.Config{
		BackendType: cluster.MemoryBackend,
	}
	cl, err := cluster.New(cfConfig)
	assert.NoError(t, err)
	assert.NotNil(t, cl)

	rep1, server1 := setupServer(t, port1, port1+1000, cl)
	defer func() {
		server1.Stop()
		rep1.Stop()
	}()

	register(t, port1)

	assert.EqualValues(t, len(instances), list(t, port1))
	for _, inst := range instances {
		assert.EqualValues(t, 1, lookup(t, port1, inst.ServiceName))
	}
}

func TestReplication(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// Configure and create the cluster module
	cfConfig := &cluster.Config{
		BackendType: cluster.MemoryBackend,
	}
	cl, err := cluster.New(cfConfig)
	assert.NoError(t, err)
	assert.NotNil(t, cl)

	port1 := nextPort()
	port2 := nextPort()
	rep1, server1 := setupServer(t, port1, port1+1000, cl)
	defer func() {
		server1.Stop()
		rep1.Stop()
	}()

	rep2, server2 := setupServer(t, port2, port2+1000, cl)
	defer func() {
		server2.Stop()
		rep2.Stop()
	}()

	ids := register(t, port1)

	// Let a few milliseconds for the replication
	time.Sleep(time.Duration(1) * time.Second)

	assert.EqualValues(t, len(instances), list(t, port2))
	for _, inst := range instances {
		assert.EqualValues(t, 1, lookup(t, port2, inst.ServiceName))
	}

	for _, id := range ids {
		renew(t, port2, id)
	}
}

func TestSynchronization(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// Configure and create the cluster module
	cfConfig := &cluster.Config{
		BackendType: cluster.MemoryBackend,
	}
	cl, err := cluster.New(cfConfig)
	assert.NoError(t, err)
	assert.NotNil(t, cl)

	port1 := nextPort()
	port2 := nextPort()
	rep1, server1 := setupServer(t, port1, port1+1000, cl)
	defer func() {
		server1.Stop()
		rep1.Stop()
	}()

	ids := register(t, port1)

	assert.EqualValues(t, len(instances), list(t, port1))
	for _, inst := range instances {
		assert.EqualValues(t, 1, lookup(t, port1, inst.ServiceName))
	}

	rep2, server2 := setupServer(t, port2, port2+1000, cl)
	defer func() {
		server2.Stop()
		rep2.Stop()
	}()

	// Let a few milliseconds for the replication
	time.Sleep(time.Duration(1) * time.Second)

	assert.EqualValues(t, len(instances), list(t, port2))
	for _, inst := range instances {
		assert.EqualValues(t, 1, lookup(t, port2, inst.ServiceName))
	}

	for _, id := range ids {
		renew(t, port2, id)
	}
}

func nextPort() uint16 {
	// Note: must use this to get new port for each server created
	// otherwise there is race conditions causing unknown state of the system
	basePort++
	return basePort
}
