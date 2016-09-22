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
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/urfave/cli"
)

var _ = Describe("Config", func() {

	var (
		c    *Config
		cErr error
	)

	Context("config loaded with default values", func() {

		BeforeEach(func() {
			app := cli.NewApp()

			app.Name = "sidecar"
			app.Usage = "Amalgam8 Sidecar"
			app.Flags = Flags
			app.Action = func(context *cli.Context) error {
				c, cErr = New(context)
				return cErr
			}

			Expect(app.Run(os.Args[:1])).NotTo(HaveOccurred())
		})

		It("uses default config values", func() {
			Expect(c.Register).To(Equal(DefaultConfig.Register))
			Expect(c.Proxy).To(Equal(DefaultConfig.Proxy))
			Expect(c.Service).To(Equal(DefaultConfig.Service))
			Expect(c.Endpoint.Port).To(Equal(DefaultConfig.Endpoint.Port))
			Expect(c.Endpoint.Type).To(Equal(DefaultConfig.Endpoint.Type))
			Expect(c.Registry).To(Equal(DefaultConfig.Registry))
			Expect(c.Controller).To(Equal(DefaultConfig.Controller))
			Expect(c.Supervise).To(Equal(DefaultConfig.Supervise))
			Expect(c.App).To(Equal(DefaultConfig.App))
			Expect(c.Log).To(Equal(DefaultConfig.Log))
			Expect(c.LogstashServer).To(Equal(DefaultConfig.LogstashServer))
			Expect(c.LogLevel).To(Equal(DefaultConfig.LogLevel))
		})

		It("falls back to local IP when no hostname is specified", func() {
			Expect(c.Endpoint.Host).To(Not(BeNil()))
			Expect(net.ParseIP(c.Endpoint.Host)).To(Not(BeNil()))
		})

	})

	Context("config overiden with command line flags", func() {

		BeforeEach(func() {
			app := cli.NewApp()

			app.Name = "sidecar"
			app.Usage = "Amalgam8 Sidecar"
			app.Flags = Flags
			app.Action = func(context *cli.Context) error {
				c, cErr = New(context)
				return cErr
			}

			args := append(os.Args[:1], []string{
				"--register=true",
				"--proxy=true",
				"--service=helloworld:v1,somethingelse",
				"--endpoint_host=localhost",
				"--endpoint_port=9080",
				"--endpoint_type=https",
				"--registry_url=http://registry:8080",
				"--registry_token=local",
				"--registry_poll=5s",
				"--controller_url=http://controller:8080",
				"--controller_token=local",
				"--controller_poll=5s",
				"--supervise=true",
				"--log=true",
				"--logstash_server=logstash:8092",
				"--log_level=debug",
				"python", "productpage.py",
			}...)

			Expect(app.Run(args)).NotTo(HaveOccurred())
		})

		It("uses config values from command line flags", func() {
			Expect(c.Register).To(Equal(true))
			Expect(c.Proxy).To(Equal(true))
			Expect(c.Service.Name).To(Equal("helloworld"))
			Expect(c.Service.Tags).To(Equal([]string{"v1", "somethingelse"}))
			Expect(c.Endpoint.Host).To(Equal("localhost"))
			Expect(c.Endpoint.Port).To(Equal(9080))
			Expect(c.Endpoint.Type).To(Equal("https"))
			Expect(c.Registry.URL).To(Equal("http://registry:8080"))
			Expect(c.Registry.Token).To(Equal("local"))
			Expect(c.Registry.Poll).To(Equal(time.Duration(5) * time.Second))
			Expect(c.Controller.URL).To(Equal("http://controller:8080"))
			Expect(c.Controller.Token).To(Equal("local"))
			Expect(c.Controller.Poll).To(Equal(time.Duration(5) * time.Second))
			Expect(c.Supervise).To(Equal(true))
			Expect(c.App).To(Equal([]string{"python", "productpage.py"}))
			Expect(c.Log).To(Equal(true))
			Expect(c.LogstashServer).To(Equal("logstash:8092"))
			Expect(c.LogLevel).To(Equal("debug"))
		})
	})

	Context("config overiden with environment variables", func() {

		BeforeEach(func() {
			app := cli.NewApp()

			app.Name = "sidecar"
			app.Usage = "Amalgam8 Sidecar"
			app.Flags = Flags
			app.Action = func(context *cli.Context) error {
				c, cErr = New(context)
				return cErr
			}

			os.Setenv("A8_REGISTER", "true")
			os.Setenv("A8_PROXY", "true")
			os.Setenv("A8_SERVICE", "helloworld:v1,somethingelse")
			os.Setenv("A8_ENDPOINT_HOST", "localhost")
			os.Setenv("A8_ENDPOINT_PORT", "9080")
			os.Setenv("A8_ENDPOINT_TYPE", "https")
			os.Setenv("A8_REGISTRY_URL", "http://registry:8080")
			os.Setenv("A8_REGISTRY_TOKEN", "local")
			os.Setenv("A8_REGISTRY_POLL", "5s")
			os.Setenv("A8_CONTROLLER_URL", "http://controller:8080")
			os.Setenv("A8_CONTROLLER_TOKEN", "local")
			os.Setenv("A8_CONTROLLER_POLL", "5s")
			os.Setenv("A8_SUPERVISE", "true")
			os.Setenv("A8_LOG", "true")
			os.Setenv("A8_LOGSTASH_SERVER", "logstash:8092")
			os.Setenv("A8_LOG_LEVEL", "debug")

			args := append(os.Args[:1], []string{
				"python", "productpage.py",
			}...)
			Expect(app.Run(args)).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.Unsetenv("A8_REGISTER")
			os.Unsetenv("A8_PROXY")
			os.Unsetenv("A8_SERVICE")
			os.Unsetenv("A8_ENDPOINT_HOST")
			os.Unsetenv("A8_ENDPOINT_PORT")
			os.Unsetenv("A8_ENDPOINT_TYPE")
			os.Unsetenv("A8_REGISTRY_URL")
			os.Unsetenv("A8_REGISTRY_TOKEN")
			os.Unsetenv("A8_REGISTRY_POLL")
			os.Unsetenv("A8_CONTROLLER_URL")
			os.Unsetenv("A8_CONTROLLER_TOKEN")
			os.Unsetenv("A8_CONTROLLER_POLL")
			os.Unsetenv("A8_SUPERVISE")
			os.Unsetenv("A8_LOG")
			os.Unsetenv("A8_LOGSTASH_SERVER")
			os.Unsetenv("A8_LOG_LEVEL")
		})

		It("uses config values from environment variables", func() {
			Expect(c.Register).To(Equal(true))
			Expect(c.Proxy).To(Equal(true))
			Expect(c.Service.Name).To(Equal("helloworld"))
			Expect(c.Service.Tags).To(Equal([]string{"v1", "somethingelse"}))
			Expect(c.Endpoint.Host).To(Equal("localhost"))
			Expect(c.Endpoint.Port).To(Equal(9080))
			Expect(c.Endpoint.Type).To(Equal("https"))
			Expect(c.Registry.URL).To(Equal("http://registry:8080"))
			Expect(c.Registry.Token).To(Equal("local"))
			Expect(c.Registry.Poll).To(Equal(time.Duration(5) * time.Second))
			Expect(c.Controller.URL).To(Equal("http://controller:8080"))
			Expect(c.Controller.Token).To(Equal("local"))
			Expect(c.Controller.Poll).To(Equal(time.Duration(5) * time.Second))
			Expect(c.Supervise).To(Equal(true))
			Expect(c.App).To(Equal([]string{"python", "productpage.py"}))
			Expect(c.Log).To(Equal(true))
			Expect(c.LogstashServer).To(Equal("logstash:8092"))
			Expect(c.LogLevel).To(Equal("debug"))
		})
	})

	Context("config overiden with configuration file", func() {

		configFile := fmt.Sprintf("%s/%s", os.TempDir(), "sidecar-config.yaml")

		BeforeEach(func() {
			app := cli.NewApp()

			app.Name = "sidecar"
			app.Usage = "Amalgam8 Sidecar"
			app.Flags = Flags
			app.Action = func(context *cli.Context) error {
				c, cErr = New(context)
				return cErr
			}

			configYaml := `
register: true
proxy: true

service:
  name: helloworld
  tags:
    - v1
    - somethingelse

endpoint:
  host: localhost
  port: 9080
  type: https

registry:
  url:   http://registry:8080
  token: local
  poll:  5s

controller:
  url:   http://controller:8080
  token: local
  poll:  5s

supervise: true
app: [ "python", "productpage.py" ]

log: true
logstash_server: logstash:8092

log_level: debug
`
			err := ioutil.WriteFile(configFile, []byte(configYaml), 0777)
			Expect(err).NotTo(HaveOccurred())

			args := append(os.Args[:1], []string{
				"--config=" + configFile,
			}...)

			Expect(app.Run(args)).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.Remove(configFile)
		})

		It("uses config values from configuration file", func() {
			Expect(c.Register).To(Equal(true))
			Expect(c.Proxy).To(Equal(true))
			Expect(c.Service.Name).To(Equal("helloworld"))
			Expect(c.Service.Tags).To(Equal([]string{"v1", "somethingelse"}))
			Expect(c.Endpoint.Host).To(Equal("localhost"))
			Expect(c.Endpoint.Port).To(Equal(9080))
			Expect(c.Endpoint.Type).To(Equal("https"))
			Expect(c.Registry.URL).To(Equal("http://registry:8080"))
			Expect(c.Registry.Token).To(Equal("local"))
			Expect(c.Registry.Poll).To(Equal(time.Duration(5) * time.Second))
			Expect(c.Controller.URL).To(Equal("http://controller:8080"))
			Expect(c.Controller.Token).To(Equal("local"))
			Expect(c.Controller.Poll).To(Equal(time.Duration(5) * time.Second))
			Expect(c.Supervise).To(Equal(true))
			Expect(c.App).To(Equal([]string{"python", "productpage.py"}))
			Expect(c.Log).To(Equal(true))
			Expect(c.LogstashServer).To(Equal("logstash:8092"))
			Expect(c.LogLevel).To(Equal("debug"))
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
