// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package flags_test

import (
	"github.com/amalgam8/amalgam8/cli/flags"
	"github.com/amalgam8/amalgam8/cli/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFlags(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flags Suite")

}

var _ = Describe("Flags", func() {

	utils.LoadLocales()

	Describe("Load Flags function", func() {
		It("should return an array of flags", func() {
			f := flags.GlobalFlags()
			Expect(len(f)).NotTo(BeZero())
		})
	})
})
