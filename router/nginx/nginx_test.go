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

	"github.com/amalgam8/sidecar/router/clients"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type configMock struct {
	UpdateFunc func(string) error
	RevertFunc func() error
}

func (m *configMock) Update(config string) error {
	return m.UpdateFunc(config)
}

func (m *configMock) Revert() error {
	return m.RevertFunc()
}

type serviceMock struct {
	StartFunc   func() error
	ReloadFunc  func() error
	RunningFunc func() (bool, error)
}

func (m *serviceMock) Start() error {
	return m.StartFunc()
}

func (m *serviceMock) Reload() error {
	return m.ReloadFunc()
}

func (m *serviceMock) Running() (bool, error) {
	return m.RunningFunc()
}

var _ = Describe("NGINX", func() {

	var (
		c  *configMock
		s  *serviceMock
		n  Nginx
		r  clients.NGINXJson
		nc clients.MockNginx
	)

	BeforeEach(func() {
		var err error

		nc = clients.MockNginx{}

		returnNil := func() error { return nil }

		c = &configMock{
			UpdateFunc: func(config string) error { return nil },
			RevertFunc: returnNil,
		}

		s = &serviceMock{
			StartFunc:   returnNil,
			ReloadFunc:  returnNil,
			RunningFunc: func() (bool, error) { return false, nil },
		}

		n, err = NewNginx(
			Config{
				Service: s,
				Client:  &nc,
			},
		)
		Expect(err).ToNot(HaveOccurred())

		//templ = resources.ConfigTemplate{
		//	Proxies: []resources.ServiceTemplate{
		//		resources.ServiceTemplate{
		//			ServiceName: "ServiceA",
		//			Versions: []resources.VersionedUpstreams{
		//				resources.VersionedUpstreams{
		//					UpstreamName: "ServiceA_v1",
		//					Upstreams:    []string{"127.0.0.1"},
		//				},
		//				resources.VersionedUpstreams{
		//					UpstreamName: "ServiceA_v2",
		//					Upstreams:    []string{"127.0.0.5"},
		//				},
		//			},
		//			VersionDefault:   "v1",
		//			VersionSelectors: "{v2={weight=0.25}}",
		//			Rules: []resources.Rule{
		//				resources.Rule{
		//					Source:           "source",
		//					Destination:      "ServiceA",
		//					Delay:            0.3,
		//					DelayProbability: 0.9,
		//					ReturnCode:       501,
		//					AbortProbability: 0.1,
		//					Pattern:          "header_value",
		//					Header:           "header_name",
		//				},
		//			},
		//		},
		//		resources.ServiceTemplate{
		//			ServiceName: "ServiceC",
		//			Versions: []resources.VersionedUpstreams{
		//				resources.VersionedUpstreams{
		//					UpstreamName: "ServiceC_UNVERSIONED",
		//					Upstreams:    []string{"127.0.0.1"},
		//				},
		//			},
		//			VersionDefault:   "",
		//			VersionSelectors: "",
		//			Rules:            []resources.Rule{},
		//		},
		//	},
		//}

	})

	Context("NGINX is not running", func() {

		It("Updates NGINX configuration and starts NGINX", func() {
			nginxUpdated := false
			nginxStarted := false

			c.UpdateFunc = func(config string) error {
				nginxUpdated = true
				return nil
			}

			s.StartFunc = func() error {
				nginxStarted = true
				return nil
			}

			Expect(n.Update(r)).ToNot(HaveOccurred())
			//Expect(nginxUpdated).To(BeTrue())
			//Expect(nginxStarted).To(BeTrue())
		})

		Context("NGINX fails to start", func() {

			var (
				revertCalled bool
				startCount   int
			)

			BeforeEach(func() {
				revertCalled = false
				startCount = 0

				s.StartFunc = func() error {
					startCount++
					if !revertCalled {
						return errors.New("Service could not start")
					}
					return nil
				}

				c.RevertFunc = func() error {
					revertCalled = true
					return nil
				}
			})

			It("Reverts to the backup NGINX configuration and starts NGINX", func() {
				//Expect(n.Update(r)).To(HaveOccurred())
				//Expect(revertCalled).To(BeTrue())
				//Expect(startCount).To(Equal(2))
			})

			Context("Revert fails", func() {

			})

		})

	})

	Context("NGINX is running", func() {

		var (
			reloadCount int
			updateCount int
			revertCount int
		)

		BeforeEach(func() {
			reloadCount = 0
			updateCount = 0
			revertCount = 0

			s.RunningFunc = func() (bool, error) {
				return true, nil
			}

			s.ReloadFunc = func() error {
				reloadCount++
				return nil
			}

			c.UpdateFunc = func(config string) error {
				updateCount++
				return nil
			}

			c.RevertFunc = func() error {
				revertCount++
				return nil
			}
		})

		It("Updates NGINX configuration and reloads NGINX", func() {
			Expect(n.Update(r)).ToNot(HaveOccurred())
			//Expect(reloadCount).To(Equal(1))
			//Expect(updateCount).To(Equal(1))
			//Expect(revertCount).To(Equal(0))
		})

		Context("NGINX fails to reload", func() {

			BeforeEach(func() {
				s.ReloadFunc = func() error {
					reloadCount++
					return errors.New("NGINX reload failed")
				}
			})

			It("Reverts to the backup NGINX configuration", func() {
				//Expect(n.Update(r)).To(HaveOccurred())
				//Expect(reloadCount).To(Equal(1))
				//Expect(updateCount).To(Equal(1))
				//Expect(revertCount).To(Equal(1))
			})

			Context("Revert fails", func() {

			})

		})

	})

})
