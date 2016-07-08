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
	"testing"
	"time"

	"io"

	"github.com/amalgam8/registry/cluster"
	"github.com/stretchr/testify/assert"
)

const (
	testDisconnectedThreshold = 100 * time.Millisecond
)

func TestHealthyNoClients(t *testing.T) {

	health := newHealthChecker()
	health.disconnectedThreshold = testDisconnectedThreshold

	status := health.Check()
	assert.True(t, status.Healthy, "Expected healthy but got %v", status)

}

func TestHealthySingleClient(t *testing.T) {

	health := newHealthChecker()
	health.disconnectedThreshold = testDisconnectedThreshold

	id := cluster.MemberID("member-id")

	health.AddClient(id, &mockClient{connected: true})

	status := health.Check()
	assert.True(t, status.Healthy, "Expected healthy but got %v", status)

}

func TestUnhealthySingleClient(t *testing.T) {

	health := newHealthChecker()
	health.disconnectedThreshold = testDisconnectedThreshold

	id := cluster.MemberID("member-id")

	health.AddClient(id, &mockClient{connected: false})

	status := health.Check()
	assert.True(t, status.Healthy, "Expected healthy but got %v", status)

	time.Sleep(testDisconnectedThreshold)

	status = health.Check()
	assert.False(t, status.Healthy, "Expected unhealthy but got %v", status)

}

func TestMultipleClientsBecomeHealthy(t *testing.T) {

	health := newHealthChecker()
	health.disconnectedThreshold = testDisconnectedThreshold

	id1 := cluster.MemberID("member-id-1")
	id2 := cluster.MemberID("member-id-2")
	id3 := cluster.MemberID("member-id-3")

	cl1 := &mockClient{connected: true}
	cl2 := &mockClient{connected: true}
	cl3 := &mockClient{connected: false}

	health.AddClient(id1, cl1)
	health.AddClient(id2, cl2)
	health.AddClient(id3, cl3)

	status := health.Check()
	assert.True(t, status.Healthy, "Expected healthy but got %v", status)

	time.Sleep(testDisconnectedThreshold)

	status = health.Check()
	assert.False(t, status.Healthy, "Expected unhealthy but got %v", status)

	_, err := cl3.connect()
	if err != nil {
		// errcheck bypass
	}

	status = health.Check()
	assert.True(t, status.Healthy, "Expected healthy but got %v", status)

}

func TestMultipleClientsBecomeUnhealthy(t *testing.T) {

	health := newHealthChecker()
	health.disconnectedThreshold = testDisconnectedThreshold

	id1 := cluster.MemberID("member-id-1")
	id2 := cluster.MemberID("member-id-2")
	id3 := cluster.MemberID("member-id-3")

	cl1 := &mockClient{connected: true}
	cl2 := &mockClient{connected: true}
	cl3 := &mockClient{connected: true}

	health.AddClient(id1, cl1)
	health.AddClient(id2, cl2)
	health.AddClient(id3, cl3)

	cl1.close()
	status := health.Check()
	assert.True(t, status.Healthy, "Expected healthy but got %v", status)

	time.Sleep(testDisconnectedThreshold)

	status = health.Check()
	assert.False(t, status.Healthy, "Expected unhealthy but got %v", status)

}

type mockClient struct {
	connected bool
}

func (mc *mockClient) connect() (io.ReadCloser, error) {
	mc.connected = true
	return nil, nil
}

func (mc *mockClient) close() {
	mc.connected = false
}

func (mc *mockClient) isConnected() bool {
	return mc.connected
}
