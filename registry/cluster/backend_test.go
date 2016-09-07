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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Generic test suite for the backend interface.
// Usage guidelines:
//
// Create a derived test suite, embedding BackendSuite:
//
// 		type MyBackendSuite struct {
//			BackendSuite
//		}
//
// Create a 'go test'-compatible function to invoke the suite:
//
//		func TestMyBackendSuite(t *testing.T) {
//			suite.Run(t, new(MyBackendSuite))
//		}
//
// Implement any suite- or testcase-level setup/teardown methods
// according to the interfaces defined in the testify/suite package.
// Particularly, make sure to setup the BackendSuite.backend field
// with an instance of your backend of choice;
//
//
//		func (suite *MyBackendSuite) SetupTest() {
//			suite.backend = &MyBackend{}
//		}
//
//		func (suite *MyBackendSuite) TearDownTest() {
//			// Do cleanup
//		}
//

type BackendSuite struct {
	suite.Suite
	backend backend
}

func (suite *BackendSuite) TestBackendEmpty() {
	suite.assertMembers()
}

func (suite *BackendSuite) TestBackendWriteMember() {
	err := suite.backend.WriteMember(member1)
	assert.NoError(suite.T(), err)
	suite.assertMembers(member1)
}

func (suite *BackendSuite) TestBackendReWriteMember() {
	_ = suite.backend.WriteMember(member1)
	err := suite.backend.WriteMember(member1)
	assert.NoError(suite.T(), err)
	suite.assertMembers(member1)
}

func (suite *BackendSuite) TestBackendWriteMultipleMembers() {
	for _, m := range allMembers {
		err := suite.backend.WriteMember(m)
		assert.NoError(suite.T(), err, "Error writing member %v", m)
	}
	suite.assertMembers(allMembers...)
}

func (suite *BackendSuite) TestBackendDeleteMember() {
	_ = suite.backend.WriteMember(member1)
	err := suite.backend.DeleteMember(member1.ID())
	assert.NoError(suite.T(), err)
	suite.assertMembers()
}

func (suite *BackendSuite) TestBackendDeleteMemberNotWritten() {
	err := suite.backend.DeleteMember(member1.ID())
	assert.Error(suite.T(), err)
	suite.assertMembers()
}

func (suite *BackendSuite) TestBackendDeleteMemberAlreadyDeleted() {
	_ = suite.backend.WriteMember(member1)
	_ = suite.backend.DeleteMember(member1.ID())
	err := suite.backend.DeleteMember(member1.ID())

	assert.Error(suite.T(), err)
	suite.assertMembers()
}

// assertMembers asserts that the actual set of members
// reflected by the backend equal the expected set of members
func (suite *BackendSuite) assertMembers(expected ...*member) {

	// Fetch actual members and IDs
	members, err1 := suite.backend.ReadMembers()
	ids, err2 := suite.backend.ReadMemberIDs()

	// Assert no error occur
	assert.NoError(suite.T(), err1)
	assert.NoError(suite.T(), err2)

	// Assert length are as expected
	assert.Len(suite.T(), members, len(expected))
	assert.Len(suite.T(), ids, len(expected))

	// Check that all expected members exist
	for _, em := range expected {

		id := em.ID()
		_, exists := ids[id]
		assert.True(suite.T(), exists, "Expected member id %s doesn't exist", id)

		am, exists := members[id]
		assert.True(suite.T(), exists, "Expected member %v doesn't exist", em)
		suite.assertMembersEqual(em, am)

		rm, err := suite.backend.ReadMember(id)
		assert.NoError(suite.T(), err)
		suite.assertMembersEqual(em, rm)

	}

}

// assertMembersEqual asserts that the actual member is logically equivalent to the expected member
func (suite *BackendSuite) assertMembersEqual(expected *member, actual *member) bool {
	equal := true
	equal = equal && assert.Equal(suite.T(), expected.ID(), actual.ID())
	equal = equal && assert.Equal(suite.T(), expected.IP(), actual.IP())
	equal = equal && assert.Equal(suite.T(), expected.Port(), actual.Port())
	equal = equal && assert.True(suite.T(), expected.Timestamp.Equal(actual.Timestamp))
	if !equal {
		assert.Fail(suite.T(), "Actual member [%+v] is not equal to expected member [%+v]", actual, expected)
	}
	return equal
}
