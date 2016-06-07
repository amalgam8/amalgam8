package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	namespace1 = Namespace("namespace1")
	namespace2 = Namespace("namespace2")
)

func TestInvalidChain(t *testing.T) {
	cases := []struct {
		auths []Authenticator
	}{
		{auths: nil},
		{auths: []Authenticator{}},
		{auths: []Authenticator{nil}},
	}

	for _, c := range cases {
		chainAuth, err := NewChainAuthenticator(c.auths)
		assert.Error(t, err)
		assert.Nil(t, chainAuth)
	}
}

func TestSingleAuthenticatorChain(t *testing.T) {
	ma := &mockAuthenticator{namespace1, nil}
	ca, err := NewChainAuthenticator([]Authenticator{ma})
	assert.NoError(t, err)
	assert.NotNil(t, ca)

	// Case 1 - token is authorized
	ma.err = nil
	ns, err := ca.Authenticate("token")
	assert.NoError(t, err)
	assert.Equal(t, namespace1, *ns)

	// Case 2 - token is unauthorized
	ma.err = ErrUnauthorized
	ns, err = ca.Authenticate("token")
	assert.Equal(t, ErrUnauthorized, err)
	assert.Nil(t, ns)

	// Case 3 - token is unrecognized
	ma.err = ErrUnrecognizedToken
	ns, err = ca.Authenticate("token")
	assert.Equal(t, ErrUnauthorized, err)
	assert.Nil(t, ns)
}

func TestMultipleAuthenticatorChain(t *testing.T) {
	ma1 := &mockAuthenticator{namespace1, nil}
	ma2 := &mockAuthenticator{namespace2, nil}
	ca, err := NewChainAuthenticator([]Authenticator{ma1, ma2})
	assert.NoError(t, err)
	assert.NotNil(t, ca)

	// Case 1 - first authenticator authorizes the token
	ma1.err = nil
	ma2.err = ErrUnrecognizedToken
	ns, err := ca.Authenticate("token")
	assert.NoError(t, err)
	assert.Equal(t, namespace1, *ns)

	// Case 2 - first authenticator unauthorizes the token, the second authorizes
	ma1.err = ErrUnauthorized
	ma2.err = nil
	ns, err = ca.Authenticate("token")
	assert.Equal(t, ErrUnauthorized, err)
	assert.Nil(t, ns)

	// Case 3 - first authenticator unrecognizes the token, the second authorizes
	ma1.err = ErrUnrecognizedToken
	ma2.err = nil
	ns, err = ca.Authenticate("token")
	assert.NoError(t, err)
	assert.Equal(t, namespace2, *ns)
}

type mockAuthenticator struct {
	namespace Namespace
	err       error
}

func (ma mockAuthenticator) Authenticate(token string) (*Namespace, error) {
	if ma.err != nil {
		return nil, ma.err
	}
	return &ma.namespace, nil
}
