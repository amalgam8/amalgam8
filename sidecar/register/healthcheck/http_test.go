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

	"github.com/amalgam8/amalgam8/sidecar/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HTTP health check", func() {

	Context("When constructing a new HTTP health check", func() {

		var check Check
		var hc *HTTP
		var err error

		Context("Using an explicit configuration values", func() {
			conf := config.HealthCheck{
				Type:     "http",
				Value:    "http://localhost:8082/healthcheck",
				Method:   "POST",
				Code:     201,
				Interval: 45 * time.Second,
				Timeout:  5 * time.Second,
			}

			BeforeEach(func() {
				check, err = NewHTTP(conf)
				hc = check.(*HTTP)
			})

			It("Succeesfully creates a healthcheck", func() {
				Expect(check).ToNot(BeNil())
				Expect(hc).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})

			It("Uses values passed with configurations", func() {
				// TODO: Better, less ugly way to test this?
				Expect(hc.url).To(Equal(conf.Value))
				Expect(hc.code).To(Equal(conf.Code))
				Expect(hc.method).To(Equal(conf.Method))
				Expect(hc.client.Timeout).To(Equal(conf.Timeout))
			})
		})
		Context("Using default configuration values", func() {
			conf := config.HealthCheck{
				Type:  "http",
				Value: "http://localhost:8082/healthcheck",
			}

			BeforeEach(func() {
				check, err = NewHTTP(conf)
				hc = check.(*HTTP)
			})

			It("Succeeds to create a healthcheck", func() {
				Expect(hc).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})

			It("Sets default values for missing fields", func() {
				// TODO: Better, less ugly way to test this?
				Expect(hc.url).To(Equal(conf.Value))
				Expect(hc.code).ToNot(BeZero())
				Expect(hc.method).ToNot(BeZero())
				Expect(hc.client.Timeout).ToNot(BeZero())
			})

		})

		Context("Using invalid configuration values", func() {
			var conf config.HealthCheck

			// Set "base" good configuration
			BeforeEach(func() {
				conf = config.HealthCheck{
					Type:  "http",
					Value: "http://localhost:8082/healthcheck",
				}
			})

			It("Fails to create a healthcheck due to an invalid type", func() {
				conf.Type = "wtf"
				check, err = NewHTTP(conf)

				Expect(check).To(BeNil())
				Expect(err).To(HaveOccurred())
			})

			It("Fails to create a healthcheck due to an invalid URL", func() {
				conf.Value = "wtf"
				check, err = NewHTTP(conf)

				Expect(check).To(BeNil())
				Expect(err).To(HaveOccurred())
			})

			It("Fails to create a healthcheck due to a missing URL", func() {
				conf.Value = ""
				check, err = NewHTTP(conf)

				Expect(check).To(BeNil())
				Expect(err).To(HaveOccurred())
			})

			It("Fails to create a healthcheck due to an invalid method", func() {
				conf.Method = "PING"
				check, err = NewHTTP(conf)

				Expect(check).To(BeNil())
				Expect(err).To(HaveOccurred())
			})

			It("Fails to create a healthcheck due to an invalid method", func() {
				conf.Method = "PING"
				check, err = NewHTTP(conf)

				Expect(check).To(BeNil())
				Expect(err).To(HaveOccurred())
			})

			It("Fails to create a healthcheck due to an invalid code", func() {
				conf.Code = 1
				check, err = NewHTTP(conf)

				Expect(check).To(BeNil())
				Expect(err).To(HaveOccurred())
			})

			It("Fails to create a healthcheck using an empty configuration", func() {
				conf.Type = ""
				conf.Value = ""
				check, err = NewHTTP(conf)

				Expect(check).To(BeNil())
				Expect(err).To(HaveOccurred())
			})

		})

	})

})
