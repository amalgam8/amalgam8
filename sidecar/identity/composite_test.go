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

package identity

import (
	"testing"

	"errors"

	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/stretchr/testify/suite"
)

type CompositeProviderSuite struct {
	suite.Suite
}

func TestCompositeProviderSuite(t *testing.T) {
	suite.Run(t, new(CompositeProviderSuite))
}

func (s *CompositeProviderSuite) TestNoIdentity() {
	provider, err := newCompositeProvider()

	s.Require().NoError(err, "Error creating Composite Identity Provider")
	s.Require().NotNil(provider, "Composite Identity Provider should not be nil")

	si, err := provider.GetIdentity()
	s.Require().NoError(err, "Expected GetIdentity() to not fail")
	s.Require().Nil(si, "Expected a nil service instance when no sub-providers exist")
}

func (s *CompositeProviderSuite) TestNilIdentity() {
	provider, err := newCompositeProvider(&mockProvider{})

	s.Require().NoError(err, "Error creating Composite Identity Provider")
	s.Require().NotNil(provider, "Composite Identity Provider should not be nil")

	si, err := provider.GetIdentity()
	s.Require().NoError(err, "Expected GetIdentity() to not fail")
	s.Require().Nil(si, "Expected a nil service instance when sub-providers return a nil identity")
}

func (s *CompositeProviderSuite) TestErrorIdentity() {
	provider, err := newCompositeProvider(&mockProvider{err: errors.New("error")})

	s.Require().NoError(err, "Error creating Composite Identity Provider")
	s.Require().NotNil(provider, "Composite Identity Provider should not be nil")

	si, err := provider.GetIdentity()
	s.Require().Error(err, "Expected GetIdentity() to not fail")
	s.Require().Nil(si, "Expected a nil service instance when sub-providers return an error")
}

func (s *CompositeProviderSuite) TestSingleIdentity() {
	expected := &api.ServiceInstance{
		ID:          "12345",
		ServiceName: "my-service",
		Tags:        []string{"tag"},
		TTL:         10,
	}

	provider, err := newCompositeProvider(&mockProvider{
		si: expected,
	})

	s.Require().NoError(err, "Error creating Composite Identity Provider")
	s.Require().NotNil(provider, "Composite Identity Provider should not be nil")

	si, err := provider.GetIdentity()
	s.Require().NoError(err, "Expected GetIdentity() to not fail")
	s.Require().NotNil(si, "Expected a non-nil service instance")

	s.Require().Equal(expected.ID, si.ID, "Unexpected ID")
	s.Require().Equal(expected.ServiceName, si.ServiceName, "Unexpected service name")
	s.Require().Equal(expected.TTL, si.TTL, "Unexpected TTL")
	s.Require().Len(si.Tags, 1, "Unexpected number of tags")
	s.Require().Contains(si.Tags, "tag", "Unexpected tag")
}

func (s *CompositeProviderSuite) TestFirstIdentityWins() {
	expected := &api.ServiceInstance{
		ID:          "12345",
		ServiceName: "my-service",
		Tags:        []string{"tag"},
		TTL:         10,
	}
	unexpected := &api.ServiceInstance{
		ID:          "abcde",
		ServiceName: "other-service",
		Tags:        []string{"other-tag"},
		TTL:         100,
	}
	provider, err := newCompositeProvider(&mockProvider{si: expected}, &mockProvider{si: unexpected})

	s.Require().NoError(err, "Error creating Composite Identity Provider")
	s.Require().NotNil(provider, "Composite Identity Provider should not be nil")

	si, err := provider.GetIdentity()
	s.Require().NoError(err, "Expected GetIdentity() to not fail")
	s.Require().NotNil(si, "Expected a non-nil service instance")

	s.Require().Equal(expected.ID, si.ID, "Unexpected ID")
	s.Require().Equal(expected.ServiceName, si.ServiceName, "Unexpected service name")
	s.Require().Equal(expected.TTL, si.TTL, "Unexpected TTL")
	s.Require().Len(si.Tags, 1, "Unexpected number of tags")
	s.Require().Contains(si.Tags, "tag", "Unexpected tag")
}

func (s *CompositeProviderSuite) TestSecondIdentityComplements() {
	expected := &api.ServiceInstance{
		ID:          "12345",
		ServiceName: "my-service",
		Tags:        []string{"tag"},
		TTL:         10,
	}
	si1 := &api.ServiceInstance{
		ID:          expected.ID,
		ServiceName: expected.ServiceName,
	}
	si2 := &api.ServiceInstance{
		Tags: expected.Tags,
		TTL:  expected.TTL,
	}
	provider, err := newCompositeProvider(&mockProvider{si: si1}, &mockProvider{si: si2})

	s.Require().NoError(err, "Error creating Composite Identity Provider")
	s.Require().NotNil(provider, "Composite Identity Provider should not be nil")

	si, err := provider.GetIdentity()
	s.Require().NoError(err, "Expected GetIdentity() to not fail")
	s.Require().NotNil(si, "Expected a non-nil service instance")

	s.Require().Equal(expected.ID, si.ID, "Unexpected ID")
	s.Require().Equal(expected.ServiceName, si.ServiceName, "Unexpected service name")
	s.Require().Equal(expected.TTL, si.TTL, "Unexpected TTL")
	s.Require().Len(si.Tags, 1, "Unexpected number of tags")
	s.Require().Contains(si.Tags, "tag", "Unexpected tag")
}

type mockProvider struct {
	si  *api.ServiceInstance
	err error
}

func (mp *mockProvider) GetIdentity() (*api.ServiceInstance, error) {
	return mp.si, mp.err
}
