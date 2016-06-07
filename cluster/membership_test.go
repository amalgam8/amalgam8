package cluster

import (
	"testing"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testTTL      = time.Duration(100) * time.Millisecond
	testInterval = time.Duration(5) * time.Millisecond
)

type MembershipSuite struct {
	suite.Suite
	backend    backend
	membership *membership
	listener   *mockedListener
}

func TestMembershipSuite(t *testing.T) {
	suite.Run(t, new(MembershipSuite))
}

func (suite *MembershipSuite) SetupTest() {
	metrics.DefaultRegistry.UnregisterAll()
	suite.backend = newMemoryBackend()
	suite.membership = newMembership(suite.backend, testTTL, testInterval)
	suite.listener = &mockedListener{}
	suite.membership.RegisterListener(suite.listener)
}

func (suite *MembershipSuite) TearDownTest() {
	suite.membership.DeregisterListener(suite.listener)
	suite.assertExpectations()
	suite.backend = nil
	suite.membership = nil
	suite.listener = nil
}

func (suite *MembershipSuite) assertMembership(expected ...*member) {
	actual := suite.membership.Members()
	assert.Len(suite.T(), actual, len(expected), "Expected %d members in membership, but found %d", len(expected), len(actual))
	for _, em := range expected {
		_, exists := actual[em]
		assert.True(suite.T(), exists, "Expected member %v does not exist in membership", em.ID())
	}
}

func (suite *MembershipSuite) assertBackend(expected ...*member) {
	actual, _ := suite.backend.ReadMembers()
	assert.Len(suite.T(), actual, len(expected), "Expected %d members in backend, but found %d", len(expected), len(actual))
	for _, em := range expected {
		am, exists := actual[em.ID()]
		assert.True(suite.T(), exists, "Expected member %v does not exist in backend", em.ID())
		assert.Equal(suite.T(), em, am)
	}
}

func (suite *MembershipSuite) assertExpectations() {
	suite.listener.AssertExpectations(suite.T())
}

////////////////////
// Tests start here
//

func (suite *MembershipSuite) TestNoMembers() {
	suite.assertMembership()
}

func (suite *MembershipSuite) TestJoinedMemberPickedUp() {
	suite.membership.StartMonitoring()
	time.Sleep(testTTL)

	suite.listener.expectJoin(member1)
	agent := suite.newAgent(member1)

	time.Sleep(testTTL)
	suite.assertMembership(member1)
	agent.stopRenewing(false)
}

func (suite *MembershipSuite) TestLeftMemberDropped() {
	suite.membership.StartMonitoring()
	time.Sleep(testTTL)

	suite.listener.expectJoin(member1)
	suite.listener.expectLeave(member1)
	agent := suite.newAgent(member1)

	time.Sleep(testTTL)
	agent.stopRenewing(true)

	time.Sleep(testTTL)
	suite.assertMembership()
}

func (suite *MembershipSuite) TestExpiredMemberRemoved() {
	suite.membership.StartMonitoring()
	time.Sleep(testTTL)

	suite.listener.expectJoin(member1)
	suite.listener.expectLeave(member1)
	agent := suite.newAgent(member1)

	time.Sleep(testTTL)
	agent.stopRenewing(false)

	time.Sleep(testTTL)
	suite.assertMembership()
	suite.assertBackend()
}

func (suite *MembershipSuite) TestStopMonitoring() {
	suite.membership.StartMonitoring()
	time.Sleep(testTTL)

	suite.listener.expectJoin(member1)
	suite.listener.expectLeave(member1)
	agent := suite.newAgent(member1)

	time.Sleep(testTTL)
	suite.membership.StopMonitoring()

	time.Sleep(testTTL)
	suite.assertMembership()
	agent.stopRenewing(true)
}

func (suite *MembershipSuite) TestCallbackNoDeadlock() {
	suite.membership.StartMonitoring()
	time.Sleep(testTTL)

	var members map[Member]struct{}
	suite.listener.joinFn = func() {
		members = suite.membership.Members()
	}

	suite.listener.expectJoin(member1)
	agent := suite.newAgent(member1)

	time.Sleep(testTTL)
	suite.Assertions.NotNil(members)
	agent.stopRenewing(false)
}

func (suite *MembershipSuite) TestSlowCallback() {
	suite.membership.StartMonitoring()
	time.Sleep(testTTL)

	suite.listener.joinFn = func() {
		time.Sleep(testTTL * 3)
	}

	suite.listener.expectJoin(member1)
	suite.listener.expectLeave(member1)
	agent := suite.newAgent(member1)

	time.Sleep(testTTL)
	agent.stopRenewing(false)
	time.Sleep(testTTL * 5)

}

func (suite *MembershipSuite) TestMembershipMetrics() {
	calcSize := func() int64 {
		gauge, _ := metrics.Get(membershipSizeMetricName).(metrics.Gauge)
		require.NotNil(suite.T(), gauge, "Failed to retrieve '%s' metric", membershipSizeMetricName)
		return gauge.Value()
	}
	calcChurn := func() int64 {
		meter, _ := metrics.Get(membershipChurnMetricName).(metrics.Meter)
		require.NotNil(suite.T(), meter, "Failed to retrieve '%s' metric", membershipChurnMetricName)
		return meter.Count()
	}

	suite.membership.StartMonitoring()
	time.Sleep(testTTL)

	assert.EqualValues(suite.T(), 0, calcSize())
	assert.EqualValues(suite.T(), 0, calcChurn())

	suite.listener.expectJoin(member1)
	suite.listener.expectJoin(member2)
	suite.listener.expectJoin(member3)
	suite.listener.expectLeave(member2)
	suite.listener.expectLeave(member3)

	agent1 := suite.newAgent(member1)

	time.Sleep(testTTL)
	assert.EqualValues(suite.T(), 1, calcSize())
	assert.EqualValues(suite.T(), 1, calcChurn())

	agent2 := suite.newAgent(member2)
	agent3 := suite.newAgent(member3)

	time.Sleep(testTTL)
	assert.EqualValues(suite.T(), 3, calcSize())
	assert.EqualValues(suite.T(), 3, calcChurn())

	agent2.stopRenewing(true)  // Leave
	agent3.stopRenewing(false) // Expire

	time.Sleep(testTTL)
	assert.EqualValues(suite.T(), 1, calcSize())
	assert.EqualValues(suite.T(), 5, calcChurn())

	agent1.stopRenewing(false)
}

type mockedListener struct {
	mock.Mock
	joinFn  func()
	leaveFn func()
}

func (l *mockedListener) OnJoin(m Member) {
	l.Called(m)
	if l.joinFn != nil {
		l.joinFn()
	}
}

// Invoked when a member leaves a cluster.
func (l *mockedListener) OnLeave(m Member) {
	l.Called(m)
	if l.leaveFn != nil {
		l.leaveFn()
	}
}

func (l *mockedListener) expectJoin(m Member) {
	l.On("OnJoin", m).Return().Once()
}

func (l *mockedListener) expectLeave(m Member) {
	l.On("OnLeave", m).Return().Once()
}

type agent struct {
	member  *member
	done    chan bool
	backend backend
}

func (suite *MembershipSuite) newAgent(m *member) *agent {
	a := &agent{
		member:  m,
		done:    make(chan bool),
		backend: suite.backend,
	}
	go a.continuouslyRenew()
	return a
}

func (a *agent) continuouslyRenew() {
	for {
		select {
		case delete := <-a.done:
			if delete {
				_ = a.backend.DeleteMember(a.member.ID())
			}
			close(a.done)
			return
		default:
			a.member.Timestamp = time.Now()
			_ = a.backend.WriteMember(a.member)
			time.Sleep(testTTL / 2)
		}
	}
}

func (a *agent) stopRenewing(delete bool) {
	a.done <- delete
	select {
	case <-a.done:
		// channel was closed
	}
}
