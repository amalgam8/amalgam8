package config

import (
	"github.com/codegangsta/cli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"time"
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
			Expect(c.Tenant.Port).To(Equal(8080))
			Expect(c.Nginx.Port).To(Equal(6379))
		})

	})

	Context("config validation", func() {

		BeforeEach(func() {
			c = &Config{
				Tenant: Tenant{
					ID:        "id",
					Token:     "token",
					TTL:       60 * time.Second,
					Heartbeat: 30 * time.Second,
					Port:      8080,
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
					URL:  "http://controller",
					Poll: 60 * time.Second,
				},
				Proxy:        true,
				Register:     true,
				ServiceName:  "mock",
				EndpointHost: "mockhost",
				EndpointPort: 9090,
			}
		})

		It("accepts a valid config", func() {
			Expect(c.Validate()).ToNot(HaveOccurred())
		})

		It("accepts a valid config without Kafka", func() {
			c.Kafka = Kafka{}
			Expect(c.Validate()).ToNot(HaveOccurred())
		})

		It("rejects an invalid URL", func() {
			c.Controller.URL = "123456"
			Expect(c.Validate()).To(HaveOccurred())
		})

		It("rejects an empty tenant ID", func() {
			c.Tenant.ID = ""
			Expect(c.Validate()).To(HaveOccurred())
		})

		It("rejects an invalid port", func() {
			c.Tenant.Port = 0
			Expect(c.Validate()).To(HaveOccurred())
		})

		It("rejects an excessively large poll interval", func() {
			c.Controller.Poll = 48 * time.Hour
			Expect(c.Validate()).To(HaveOccurred())
		})

		It("rejects a TTL that is less than the heartbeat", func() {
			c.Tenant.Heartbeat = 5 * time.Minute
			c.Tenant.TTL = 2 * time.Minute
			Expect(c.Validate()).To(HaveOccurred())
		})

		It("rejects empty brokers", func() {
			c.Kafka.Brokers = []string{}
			Expect(c.Validate()).To(HaveOccurred())
		})

		It("rejects invalid brokers", func() {
			c.Kafka.Brokers = []string{
				"",
				"",
				"",
			}
			Expect(c.Validate()).To(HaveOccurred())
		})

		It("rejects partial config", func() {
			c.Kafka.Username = ""
			Expect(c.Validate()).To(HaveOccurred())
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
			Expect(c.Validate()).ToNot(HaveOccurred())
		})

	})

})
