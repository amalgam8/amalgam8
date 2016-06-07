package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidGlobalToken(t *testing.T) {
	g := globalAuth
	namespace, err := g.Authenticate("")
	assert.NoError(t, err)
	assert.EqualValues(t, globalNamespace, *namespace)
}

func TestInvalidGlobalToken(t *testing.T) {
	g := globalAuth
	_, err := g.Authenticate("invalid-token")
	assert.Error(t, err)
}
