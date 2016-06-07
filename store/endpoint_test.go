package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloneEndpoint(t *testing.T) {

	original := &Endpoint{
		Value: "127.0.0.1:9080",
		Type:  "tcp",
	}

	cloned := original.DeepClone()

	assert.Equal(t, original, cloned)
	assert.False(t, &original == &cloned)

	cloned.Value = "127.0.0.1:9443"
	assert.NotEqual(t, original, cloned)

}
