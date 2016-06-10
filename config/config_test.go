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

			app.Name = "controller"
			app.Usage = "Amalgam8 Controller"
			app.Flags = Flags
			app.Action = func(context *cli.Context) {
				c = New(context)
			}

			Expect(app.Run([]string{"cmd"})).NotTo(HaveOccurred())
		})

		It("has expected default", func() {
			// Expected defaults specified in documentation
			Expect(c.APIPort).To(Equal(6379))
			Expect(c.Database.Type).To(Equal("memory"))
			Expect(c.StatsdHost).To(Equal("127.0.0.1:8125"))
		})

	})

	Context("config validation", func() {

		BeforeEach(func() {

			c = &Config{
				APIPort:      6379,
				ControlToken: "abcdefghijklmnop",
				SecretKey:    "ABCEDFGHIJKLMNOP",
				StatsdHost:   "127.0.0.1:8125",
				Database: Database{
					Type: "memory",
				},
			}
		})

		It("accepts a valid config", func() {
			Expect(c.Validate()).ToNot(HaveOccurred())
		})

		It("does not accept empty control token", func() {
			c.ControlToken = ""
			Expect(c.Validate()).To(HaveOccurred())
		})

		It("does not accept empty secret key", func() {
			c.SecretKey = ""
			Expect(c.Validate()).To(HaveOccurred())
		})

		It("does not accept secret key that does not have 16 chars", func() {
			c.SecretKey = "abcd"
			Expect(c.Validate()).To(HaveOccurred())
			c.SecretKey = "abcdefghijklmnopq"
			Expect(c.Validate()).To(HaveOccurred())
		})

		It("does not accept empty statsd host", func() {
			c.StatsdHost = ""
			Expect(c.Validate()).To(HaveOccurred())
		})

		Context("Invalid database fields", func() {

			It("does not accept empty database type", func() {
				c.Database.Type = ""
				Expect(c.Validate()).To(HaveOccurred())
			})

			It("does not accept database type other than memory or cloudant", func() {
				c.Database.Type = "gihanson"
				Expect(c.Validate()).To(HaveOccurred())
			})

			It("does not accept empty username if cloudant type provided", func() {
				c.Database.Type = "cloudant"
				c.Database.Password = "password"
				c.Database.Host = "dbhost"
				Expect(c.Validate()).To(HaveOccurred())
			})

			It("does not accept empty password if cloudant type provided", func() {
				c.Database.Type = "cloudant"
				c.Database.Username = "username"
				c.Database.Host = "dbhost"
				Expect(c.Validate()).To(HaveOccurred())
			})

			It("does not accept empty host if cloudant type provided", func() {
				c.Database.Type = "cloudant"
				c.Database.Password = "password"
				c.Database.Username = "username"
				Expect(c.Validate()).To(HaveOccurred())
			})
		})
	})

})
