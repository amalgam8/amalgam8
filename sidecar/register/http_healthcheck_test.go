package register

import (
	"time"

	"github.com/amalgam8/amalgam8/sidecar/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HTTPHealthCheck", func() {

	Context("When constructing a new HTTPHealthCheck", func() {

		var hc *HTTPHealthCheck
		var err error

		Context("Using an explicit configuraiton values", func() {
			conf := config.HealthCheck{
				Type:     "http",
				Value:    "http://localhost:8082/healthcheck",
				Method:   "POST",
				Code:     201,
				Interval: 45 * time.Second,
				Timeout:  5 * time.Second,
			}

			BeforeEach(func() {
				hc, err = NewHTTPHealthCheck(conf)
			})

			It("Succeeds to create a healthcheck", func() {
				Expect(hc).To(Not(BeNil()))
				Expect(err).To(Not(HaveOccurred()))
			})

			It("Uses values passed with configurations", func() {
				// TODO: Better, less ugly way to test this?
				Expect(hc.url).To(Equal(conf.Value))
				Expect(hc.code).To(Equal(conf.Code))
				Expect(hc.method).To(Equal(conf.Method))
				Expect(hc.interval).To(Equal(conf.Interval))
				Expect(hc.client.Timeout).To(Equal(conf.Timeout))
			})
		})
		Context("Using default configuration values", func() {
			conf := config.HealthCheck{
				Type:  "http",
				Value: "http://localhost:8082/healthcheck",
			}

			BeforeEach(func() {
				hc, err = NewHTTPHealthCheck(conf)
			})

			It("Succeeds to create a healthcheck", func() {
				Expect(hc).To(Not(BeNil()))
				Expect(err).To(Not(HaveOccurred()))
			})

			It("Sets default values for missing fields", func() {
				// TODO: Better, less ugly way to test this?
				Expect(hc.url).To(Equal(conf.Value))
				Expect(hc.code).To(Not(BeZero()))
				Expect(hc.method).To(Not(BeZero()))
				Expect(hc.interval).To(Not(BeZero()))
				Expect(hc.client.Timeout).To(Not(BeZero()))
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
				hc, err = NewHTTPHealthCheck(conf)

				Expect(hc).To(BeNil())
				Expect(err).To(HaveOccurred())
			})

			It("Fails to create a healthcheck due to an invalid URL", func() {
				conf.Value = "wtf"
				hc, err = NewHTTPHealthCheck(conf)

				Expect(hc).To(BeNil())
				Expect(err).To(HaveOccurred())
			})

			It("Fails to create a healthcheck due to a missing URL", func() {
				conf.Value = ""
				hc, err = NewHTTPHealthCheck(conf)

				Expect(hc).To(BeNil())
				Expect(err).To(HaveOccurred())
			})

			It("Fails to create a healthcheck due to an invalid method", func() {
				conf.Method = "PING"
				hc, err = NewHTTPHealthCheck(conf)

				Expect(hc).To(BeNil())
				Expect(err).To(HaveOccurred())
			})

			It("Fails to create a healthcheck due to an invalid method", func() {
				conf.Method = "PING"
				hc, err = NewHTTPHealthCheck(conf)

				Expect(hc).To(BeNil())
				Expect(err).To(HaveOccurred())
			})

			It("Fails to create a healthcheck due to an invalid code", func() {
				conf.Code = 1
				hc, err = NewHTTPHealthCheck(conf)

				Expect(hc).To(BeNil())
				Expect(err).To(HaveOccurred())
			})

			It("Fails to create a healthcheck using an empty configuration", func() {
				conf.Type = ""
				conf.Value = ""
				hc, err = NewHTTPHealthCheck(conf)

				Expect(hc).To(BeNil())
				Expect(err).To(HaveOccurred())
			})

		})

	})

})
