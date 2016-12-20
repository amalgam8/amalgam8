package config_test

import (
	. "github.com/amalgam8/amalgam8/cli/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GlobalCommands", func() {
	var _ = Describe("Commands", func() {
		Describe("Load Commands function", func() {
			It("should return an array of commands", func() {
				cmds := GlobalCommands(nil)
				Expect(len(cmds)).NotTo(BeZero())
			})
		})
	})

})
