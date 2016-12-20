package config_test

import (
	"github.com/amalgam8/amalgam8/cli/config"
	"github.com/amalgam8/amalgam8/cli/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GlobalFlags", func() {

	utils.LoadLocales("../locales")

	Describe("Load Flags function", func() {
		It("should return an array of flags", func() {
			f := config.GlobalFlags()
			Expect(len(f)).NotTo(BeZero())
		})
	})
})
