package cluster

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testDirectory         = "/tmp/sd-test"
	testUnreadableTimeout = time.Duration(100) * time.Millisecond
)

type FilesystemBackendSuite struct {
	BackendSuite
}

func TestFilesystemBackendSuite(t *testing.T) {
	suite.Run(t, new(FilesystemBackendSuite))
}

func (suite *FilesystemBackendSuite) SetupTest() {
	err := os.RemoveAll(testDirectory)
	if err != nil {
		panic(err)
	}
	backend, err := newFilesystemBackend(testDirectory, testUnreadableTimeout)
	if err != nil {
		panic(err)
	}
	suite.backend = backend
}

func (suite *FilesystemBackendSuite) TearDownTest() {
	_ = os.RemoveAll(testDirectory) // Best-effort :)
	suite.backend = nil
}

func (suite *FilesystemBackendSuite) TestUnreadableFileDeleted() {

	memberID := MemberID("test-id")
	filename := testDirectory + "/" + string(memberID)

	// Create an empty (unreadble) member file
	_, err := os.Create(filename)
	require.NoError(suite.T(), err)

	// Assert file is unreadable but not yet removed
	_, err = suite.BackendSuite.backend.ReadMember(memberID)
	assert.Error(suite.T(), err)
	_, err = os.Stat(filename)
	assert.NoError(suite.T(), err)

	time.Sleep(testUnreadableTimeout * 2)

	// Assert file is unreadable and gets removed
	_, err = suite.BackendSuite.backend.ReadMember(memberID)
	assert.Error(suite.T(), err)
	_, err = os.Stat(filename)
	assert.True(suite.T(), os.IsNotExist(err))

}

func (suite *FilesystemBackendSuite) TestUnreadableFile() {

	err := suite.backend.WriteMember(member1)
	assert.NoError(suite.T(), err)

	// Add the member into the cache
	_, err = suite.backend.ReadMembers()
	assert.NoError(suite.T(), err)
	m, err := suite.backend.ReadMember(member1.ID())
	assert.NoError(suite.T(), err)
	suite.assertMembersEqual(member1, m)

	filename := testDirectory + "/" + string(member1.ID())

	// The sleep here is to make sure that the timestamp of the
	// new empty file will be greater than the timestamp of member1
	time.Sleep(testUnreadableTimeout / 2)

	// Create an empty (unreadble) member file
	_, err = os.Create(filename)
	require.NoError(suite.T(), err)

	// Assert file is unreadable but not yet removed
	m, err = suite.BackendSuite.backend.ReadMember(member1.ID())
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), m)
	assert.EqualValues(suite.T(), member1.MemberIP, m.MemberIP)
	assert.EqualValues(suite.T(), member1.MemberPort, m.MemberPort)
	assert.True(suite.T(), m.Timestamp.After(member1.Timestamp))
	_, err = os.Stat(filename)
	assert.NoError(suite.T(), err)

	time.Sleep(testUnreadableTimeout * 2)

	// Assert file is unreadable and gets removed
	m, err = suite.BackendSuite.backend.ReadMember(member1.ID())
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), m)
	_, err = os.Stat(filename)
	assert.True(suite.T(), os.IsNotExist(err))

}
