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

package cluster

import (
	"testing"

	"time"

	"math"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testClusterSize      = 2
	testSubsizeThreshold = 100 * time.Millisecond
)

var (
	zeroMembers  = memberSet([]Member{})
	oneMember    = memberSet([]Member{member1})
	twoMembers   = memberSet([]Member{member1, member2})
	threeMembers = memberSet([]Member{member1, member2, member3})
)

func TestHealthyCluster(t *testing.T) {

	membership := &mockedMembership{}
	membership.On("RegisterListener", mock.Anything).Return()
	membership.On("DeregisterListener", mock.Anything).Return()

	membership.On("Members").Return(twoMembers).Times(math.MaxInt32)

	health := newHealthChecker(membership, testClusterSize)
	health.subsizeGracePeriod = testSubsizeThreshold

	status := health.Check()
	assert.True(t, status.Healthy, "Expected healthy cluster but got %v", status)

}

func TestHealthyClusterBringup(t *testing.T) {

	membership := &mockedMembership{}
	membership.On("RegisterListener", mock.Anything).Return()
	membership.On("DeregisterListener", mock.Anything).Return()

	membership.On("Members").Return(zeroMembers).Once()
	membership.On("Members").Return(oneMember).Once()
	membership.On("Members").Return(twoMembers).Once()
	membership.On("Members").Return(threeMembers).Once()
	membership.On("Members").Return(twoMembers).Times(math.MaxInt32)

	health := newHealthChecker(membership, testClusterSize)
	health.subsizeGracePeriod = testSubsizeThreshold

	// 0 members
	status := health.Check()
	assert.True(t, status.Healthy, "Expected healthy cluster but got %v", status)

	// 1 member
	health.OnJoin(member1)
	status = health.Check()
	assert.True(t, status.Healthy, "Expected healthy cluster but got %v", status)

	// 2 members
	health.OnJoin(member2)
	status = health.Check()
	assert.True(t, status.Healthy, "Expected healthy cluster but got %v", status)

	// 3 members
	health.OnJoin(member3)
	status = health.Check()
	assert.True(t, status.Healthy, "Expected healthy cluster but got %v", status)

	// 2 members
	health.OnLeave(member1)
	status = health.Check()
	assert.True(t, status.Healthy, "Expected healthy cluster but got %v", status)

	time.Sleep(testSubsizeThreshold)
	status = health.Check()
	assert.True(t, status.Healthy, "Expected healthy cluster but got %v", status)

}

func TestUnhealthyCluster(t *testing.T) {

	membership := &mockedMembership{}
	membership.On("RegisterListener", mock.Anything).Return()
	membership.On("DeregisterListener", mock.Anything).Return()

	membership.On("Members").Return(oneMember).Times(math.MaxInt32)

	health := newHealthChecker(membership, testClusterSize)
	health.subsizeGracePeriod = testSubsizeThreshold

	status := health.Check()
	assert.True(t, status.Healthy, "Expected healthy cluster but got %v", status)

	time.Sleep(testSubsizeThreshold)
	status = health.Check()
	// Cluster health check will report that the cluster size is below the threshold
	// but will not flag it as unhealthy
	assert.True(t, status.Healthy, "Expected healthy cluster but got %v", status)

}

func TestUnhealthyClusterBringdown(t *testing.T) {

	membership := &mockedMembership{}
	membership.On("RegisterListener", mock.Anything).Return()
	membership.On("DeregisterListener", mock.Anything).Return()

	membership.On("Members").Return(twoMembers).Once()
	membership.On("Members").Return(oneMember).Times(math.MaxInt32)

	health := newHealthChecker(membership, testClusterSize)
	health.subsizeGracePeriod = testSubsizeThreshold

	// 2 members
	status := health.Check()
	assert.True(t, status.Healthy, "Expected healthy cluster but got %v", status)

	// 1 member
	health.OnLeave(member1)
	status = health.Check()
	assert.True(t, status.Healthy, "Expected healthy cluster but got %v", status)

	time.Sleep(testSubsizeThreshold)
	status = health.Check()
	// Cluster health check will report that the cluster size is below the threshold
	// but will not flag it as unhealthy
	assert.True(t, status.Healthy, "Expected healthy cluster but got %v", status)

}

// memberSet converts a Member slice to a Member set
func memberSet(slice []Member) map[Member]struct{} {
	set := make(map[Member]struct{}, len(slice))
	for _, m := range slice {
		set[m] = struct{}{}
	}
	return set
}
