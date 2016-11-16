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

package common_test

import (
	"github.com/amalgam8/amalgam8/cli/common"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Constants", func() {

	var test common.Const

	BeforeEach(func() {
		test = "THIS_IS_A_TEST"
	})

	Describe("Flags", func() {
		It("should format a string as `this_is_a_test` format", func() {
			Expect(test.Flag()).To(Equal("this_is_a_test"))
		})
	})

	Describe("Environment Variables", func() {
		It("should format a string as an `A8_THIS_IS_A_TEST`", func() {
			Expect(test.EnvVar()).To(Equal("A8_THIS_IS_A_TEST"))
		})
	})

})
