package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type RegistratorSuite struct {
	suite.Suite
	backend     backend
	membership  *mockedMembership
	registrator *registrator
}

func TestRegistratorSuite(t *testing.T) {
	suite.Run(t, new(RegistratorSuite))
}

func (suite *RegistratorSuite) SetupTest() {
	suite.backend = newMemoryBackend()
	suite.membership = &mockedMembership{}
	suite.membership.On("RegisterListener", mock.Anything).Return()
	suite.membership.On("DeregisterListener", mock.Anything).Return()
	suite.registrator = newRegistrator(suite.backend, member1, suite.membership, testInterval)
}

func (suite *RegistratorSuite) TearDownTest() {
	suite.backend = nil
	suite.membership = nil
	suite.registrator = nil
}

func (suite *RegistratorSuite) TestSelf() {
	member := suite.registrator.Self()
	assert.Equal(suite.T(), member1, member)
}

func (suite *RegistratorSuite) TestJoin() {
	err := suite.registrator.Join()
	ids, _ := suite.backend.ReadMemberIDs()

	assert.NoError(suite.T(), err)
	_, exists := ids[member1.ID()]
	assert.True(suite.T(), exists, "Expected member id %s doesn't exist", member1.ID())
	assert.Len(suite.T(), ids, 1)
}

func (suite *RegistratorSuite) TestJoinAlreadyJoined() {
	_ = suite.registrator.Join()
	err := suite.registrator.Join()

	assert.Error(suite.T(), err)
}

func (suite *RegistratorSuite) TestJoinAfterLeave() {
	_ = suite.registrator.Join()
	_ = suite.registrator.Leave()
	err := suite.registrator.Join()
	ids, _ := suite.backend.ReadMemberIDs()

	assert.NoError(suite.T(), err)
	_, exists := ids[member1.ID()]
	assert.True(suite.T(), exists, "Expected member id %s doesn't exist", member1.ID())
	assert.Len(suite.T(), ids, 1)
}

func (suite *RegistratorSuite) TestLeave() {
	_ = suite.registrator.Join()
	err := suite.registrator.Leave()
	ids, _ := suite.backend.ReadMemberIDs()

	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), ids)
}

func (suite *RegistratorSuite) TestLeaveNotJoined() {
	err := suite.registrator.Leave()

	assert.Error(suite.T(), err)
}

func (suite *RegistratorSuite) TestLeaveAlreadyLeft() {
	_ = suite.registrator.Join()
	_ = suite.registrator.Leave()
	err := suite.registrator.Leave()

	assert.Error(suite.T(), err)
}

type mockedMembership struct {
	mock.Mock
}

func (m *mockedMembership) Members() map[Member]struct{} {
	args := m.Called()
	return args.Get(0).(map[Member]struct{})
}

func (m *mockedMembership) RegisterListener(l Listener) {
	m.Called(l)
}

func (m *mockedMembership) DeregisterListener(l Listener) {
	m.Called(l)
}
