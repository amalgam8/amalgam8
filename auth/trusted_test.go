package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidTrustedToken(t *testing.T) {
	ta := NewTrustedAuthenticator()
	namespace, err := ta.Authenticate("valid-token")
	assert.NoError(t, err)
	assert.EqualValues(t, "valid-token", *namespace)
}

func TestInvalidTrustedToken(t *testing.T) {
	ta := NewTrustedAuthenticator()
	_, err := ta.Authenticate("")
	assert.Error(t, err)
}
