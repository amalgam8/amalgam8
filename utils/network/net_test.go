package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPrivateIP(t *testing.T) {
	ip := GetPrivateIP()
	assert.False(t, ip.IsUnspecified())
	assert.False(t, ip.IsLoopback())
}

func TestWaitForPrivateIP(t *testing.T) {
	assert.True(t, WaitForPrivateNetwork())
}
