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
			app.Flags = TenantFlags
			app.Action = func(context *cli.Context) {
				c = New(context)
			}

			Expect(app.Run(os.Args[:1])).NotTo(HaveOccurred())
		})

		It("has expected default ports", func() {
			// Expected defaults specified in documentation
			Expect(c.Nginx.Port).To(Equal(6379))
		})

	})

	Context("config validation", func() {

		BeforeEach(func() {
			c = &Config{
				Tenant: Tenant{
					TTL:       60 * time.Second,
					Heartbeat: 30 * time.Second,
				},
				Registry: Registry{
					URL:   "http://registry",
					Token: "sd_token",
				},
				Kafka: Kafka{
					Brokers: []string{
						"http://broker1",
						"http://broker2",
						"http://broker3",
					},
					Username: "username",
					Password: "password",
					APIKey:   "apitoken",
					RestURL:  "http://resturl",
					SASL:     true,
				},
				Nginx: Nginx{
					Port:    6379,
					Logging: false,
				},
				Controller: Controller{
					Token: "token",
					URL:   "http://controller",
					Poll:  60 * time.Second,
				},
				Proxy:        true,
				Register:     true,
				ServiceName:  "mock",
				EndpointHost: "mockhost",
				EndpointPort: 9090,
				EndpointType: "http",
			}
		})

		It("accepts a valid config", func() {
			Expect(c.Validate(true)).ToNot(HaveOccurred())
		})

		It("accepts a valid config without Kafka", func() {
			c.Kafka = Kafka{}
			Expect(c.Validate(true)).ToNot(HaveOccurred())
		})

		It("rejects an invalid URL", func() {
			c.Controller.URL = "123456"
			Expect(c.Validate(true)).To(HaveOccurred())
		})

		It("rejects an invalid port", func() {
			c.Nginx.Port = 0
			Expect(c.Validate(true)).To(HaveOccurred())
		})

		It("rejects an excessively large poll interval", func() {
			c.Controller.Poll = 48 * time.Hour
			Expect(c.Validate(true)).To(HaveOccurred())
		})

		It("rejects a TTL that is less than the heartbeat", func() {
			c.Tenant.Heartbeat = 5 * time.Minute
			c.Tenant.TTL = 2 * time.Minute
			Expect(c.Validate(true)).To(HaveOccurred())
		})

		It("rejects empty brokers", func() {
			c.Kafka.Brokers = []string{}
			Expect(c.Validate(true)).To(HaveOccurred())
		})

		It("rejects invalid brokers", func() {
			c.Kafka.Brokers = []string{
				"",
				"",
				"",
			}
			Expect(c.Validate(true)).To(HaveOccurred())
		})

		It("rejects partial config", func() {
			c.Kafka.Username = ""
			Expect(c.Validate(true)).To(HaveOccurred())
		})

		It("accepts local kafka config", func() {
			c.Kafka = Kafka{
				Brokers: []string{
					"http://broker1",
					"http://broker2",
					"http://broker3",
				},
				SASL: false,
			}
			Expect(c.Validate(true)).ToNot(HaveOccurred())
		})

	})

})
