package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func extractErrorCode(err error) ErrorCode {
	return err.(*Error).Code
}

func TestNewRegistry(t *testing.T) {

	r := New(nil, nil)
	assert.NotNil(t, r)
}
