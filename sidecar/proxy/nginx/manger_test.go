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

package nginx

import (
	"errors"

	"github.com/amalgam8/amalgam8/controller/rules"
	"github.com/amalgam8/amalgam8/pkg/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manager", func() {

	var (
		c         *MockClient
		s         *MockService
		m         Manager
		instances []api.ServiceInstance
		a8Rules   []rules.Rule
	)

	BeforeEach(func() {
		c = &MockClient{}
		s = &MockService{}
		m = NewManager(Config{
			Service: s,
			Client:  c,
		})
		instances = []api.ServiceInstance{}
		a8Rules = []rules.Rule{}
	})

	It("Manager successfully Updates", func() {
		Expect(m.Update(instances, a8Rules)).ToNot(HaveOccurred())
		Expect(s.RunningCount).To(Equal(1))
		Expect(s.StartCount).To(Equal(1))
		Expect(s.ReloadCount).To(Equal(0))
		Expect(c.UpdateCount).To(Equal(1))
	})

	Context("calls the correct corresponding nginx service command", func() {

		It("can't tell if NGINX is running", func() {
			s.RunningError = errors.New("can't tell if nginx is running or not")
			Expect(m.Update(instances, a8Rules)).To(HaveOccurred())
			Expect(s.RunningCount).To(Equal(1))
			Expect(s.StartCount).To(Equal(0))
			Expect(s.ReloadCount).To(Equal(0))
		})

		It("NGINX is running", func() {
			s.RunningBool = true
			Expect(m.Update(instances, a8Rules)).ToNot(HaveOccurred())
			Expect(s.RunningCount).To(Equal(1))
			Expect(s.StartCount).To(Equal(0))
			Expect(s.ReloadCount).To(Equal(0))
			Expect(c.UpdateCount).To(Equal(1))

		})

		It("NGINX fails to start", func() {
			s.StartError = errors.New("could not start NGINX")
			Expect(m.Update(instances, a8Rules)).To(HaveOccurred())
			Expect(s.RunningCount).To(Equal(1))
			Expect(s.StartCount).To(Equal(1))
			Expect(s.ReloadCount).To(Equal(0))
			Expect(c.UpdateCount).To(Equal(0))
		})

		It("NGINX client fails to update", func() {
			c.UpdateError = errors.New("could not update NGINX")
			Expect(m.Update(instances, a8Rules)).To(HaveOccurred())
			Expect(s.RunningCount).To(Equal(1))
			Expect(s.StartCount).To(Equal(1))
			Expect(s.ReloadCount).To(Equal(0))
			Expect(c.UpdateCount).To(Equal(1))
		})
	})
})
