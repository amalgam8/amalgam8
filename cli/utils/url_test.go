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

package utils_test

import (
	"github.com/amalgam8/amalgam8/cli/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Url", func() {
	Describe("Validating URL", func() {
		Context("when calling IsURL()", func() {

			Context("when the string passed is empty", func() {
				It("FALSE is returned", func() {
					Expect(utils.IsURL("rawURL")).To(BeFalse())
				})
			})

			Context("when the string passed is an invalid URL", func() {
				It("FALSE is returned", func() {
					Expect(utils.IsURL("rawURL")).To(BeFalse())
				})
			})

			Context("when the string passed is a valid URL", func() {
				It("TRUE is returned", func() {
					Expect(utils.IsURL("http://localhost.com")).To(BeTrue())
				})
			})

		})
	})
})
