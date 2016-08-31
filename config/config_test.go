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

package config

import (
	"os"
	"time"

	"github.com/codegangsta/cli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {

	var (
		c *Config
	)

	Context("config loaded with default values", func() {

		BeforeEach(func() {
			app := cli.NewApp()

			app.Name = "sidecar"
			app.Usage = "Amalgam8 Sidecar"
			app.Flags = Flags
			app.Action = func(context *cli.Context) {
				c, _ = New(context)
			}

			Expect(app.Run(os.Args[:1])).NotTo(HaveOccurred())
		})

	})

	Context("config validation", func() {

		BeforeEach(func() {
			c = &Config{
				Registry: Registry{
					URL:   "http://registry",
					Token: "sd_token",
				},
				Controller: Controller{
					Token: "token",
					URL:   "http://controller",
					Poll:  60 * time.Second,
				},
				Proxy:    true,
				Register: true,
				Service: Service{
					Name: "mock",
				},
				Endpoint: Endpoint{
					Host: "mockhost",
					Port: 9090,
					Type: "http",
				},
			}
		})

		It("accepts a valid config", func() {
			Expect(c.Validate()).ToNot(HaveOccurred())
		})

		It("rejects an invalid URL", func() {
			c.Controller.URL = "123456"
			Expect(c.Validate()).To(HaveOccurred())
		})

		It("rejects an excessively large poll interval", func() {
			c.Controller.Poll = 48 * time.Hour
			Expect(c.Validate()).To(HaveOccurred())
		})
	})

})
