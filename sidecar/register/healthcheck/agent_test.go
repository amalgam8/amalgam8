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

package healthcheck

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type MockCheck struct {
	ExecuteError error
}

func (m *MockCheck) Execute() error {
	return m.ExecuteError
}

var _ = Describe("Health check agent", func() {

	Context("When constructing a new health check agent", func() {

		var check Check
		var agent *Agent

		Context("Using an explicit configuration values", func() {

			BeforeEach(func() {
				check = &MockCheck{}
				agent = NewAgent(check, 5*time.Second)
			})

			It("Successfully create an agent", func() {
				Expect(agent).ToNot(BeNil())
				Expect(agent.interval).To(Equal(5 * time.Second))
			})
		})

		Context("Using default configuration values", func() {
			BeforeEach(func() {
				check = &MockCheck{}
				agent = NewAgent(check, 0)
				Expect(agent).ToNot(BeNil())
			})

			It("Sets default values for missing fields", func() {
				Expect(agent).ToNot(BeNil())
				Expect(agent.interval).To(Equal(defaultHealthCheckInterval))
			})

		})

	})

})
