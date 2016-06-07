package checker

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestTenant(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tenant Suite")
}
