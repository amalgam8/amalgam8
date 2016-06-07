package cluster

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MemoryBackendSuite struct {
	BackendSuite
}

func TestMemoryBackendSuite(t *testing.T) {
	suite.Run(t, new(MemoryBackendSuite))
}

func (suite *MemoryBackendSuite) SetupTest() {
	suite.backend = newMemoryBackend()
}

func (suite *MemoryBackendSuite) TearDownTest() {
	suite.backend = nil
}
