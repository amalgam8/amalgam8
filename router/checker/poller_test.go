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

package checker

import (
	"errors"
	"time"

	"github.com/amalgam8/controller/resources"
	"github.com/amalgam8/sidecar/config"
	"github.com/amalgam8/sidecar/router/clients"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type mockNginx struct {
	UpdateFunc func(clients.NGINXJson) error
}

func (m *mockNginx) Update(nginxJson clients.NGINXJson) error {
	return m.UpdateFunc(nginxJson)
}

var _ = Describe("Tenant Poller", func() {

	var (
		rc       *clients.MockController
		n        *mockNginx
		c        *config.Config
		p        *poller
		listener *MockListener

		updateCount int
	)

	BeforeEach(func() {
		updateCount = 0

		n = &mockNginx{
			UpdateFunc: func(nginxJson clients.NGINXJson) error {
				updateCount++
				return nil
			},
		}
		listener = &MockListener{
			mockNginx: n,
		}

		rc = &clients.MockController{
			ConfigTemplate: resources.ProxyConfig{},
		}

		c = &config.Config{
			Tenant: config.Tenant{
				Token:     "token",
				TTL:       60 * time.Second,
				Heartbeat: 30 * time.Second,
			},
			Registry: config.Registry{
				URL:   "http://registry",
				Token: "sd_token",
			},
			Nginx: config.Nginx{
				Port:    6379,
				Logging: false,
			},
			Controller: config.Controller{
				URL:  "http://controller",
				Poll: 60 * time.Second,
			},
		}

		p = &poller{
			controller: rc,
			config:     c,
			listener:   listener,
		}
	})

	It("polls successfully", func() {
		Expect(p.poll()).ToNot(HaveOccurred())
		Expect(updateCount).To(Equal(1))
	})

	It("reports NGINX update failure", func() {
		n.UpdateFunc = func(nginxJson clients.NGINXJson) error {
			return errors.New("Update NGINX failed")
		}

		Expect(p.poll()).To(HaveOccurred())
	})

	It("does not update NGINX if unable to obtain config from Controller", func() {
		rc.ConfigError = errors.New("Get rules failed")

		Expect(p.poll()).To(HaveOccurred())
		Expect(updateCount).To(Equal(0))
	})

})
