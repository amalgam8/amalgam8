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
			Expect(c.ProxyAdapter).To(Equal(DefaultConfig.ProxyAdapter))
			Expect(c.DNS).To(Equal(DefaultConfig.DNS))
			Expect(c.Service).To(Equal(DefaultConfig.Service))
			Expect(c.Endpoint.Port).To(Equal(DefaultConfig.Endpoint.Port))
			Expect(c.Endpoint.Type).To(Equal(DefaultConfig.Endpoint.Type))
			Expect(c.DiscoveryAdapter).To(Equal(DefaultConfig.DiscoveryAdapter))
			Expect(c.RulesAdapter).To(Equal(DefaultConfig.RulesAdapter))
			Expect(c.A8Registry).To(Equal(DefaultConfig.A8Registry))
			Expect(c.A8Controller).To(Equal(DefaultConfig.A8Controller))
			Expect(c.Kubernetes).To(Equal(DefaultConfig.Kubernetes))
			Expect(c.Eureka).To(Equal(DefaultConfig.Eureka))
			Expect(c.Dnsconfig).To(Equal(DefaultConfig.Dnsconfig))
			Expect(c.HealthChecks).To(Equal(DefaultConfig.HealthChecks))
			Expect(c.LogLevel).To(Equal(DefaultConfig.LogLevel))
			Expect(c.DiscoveryPort).To(Equal(DefaultConfig.DiscoveryPort))
			Expect(c.Commands).To(HaveLen(0))
		})

	})

	Context("config overidden with command line flags", func() {

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
				"--proxy_adapter=envoy",
				"--dns=true",
				"--proxy_tls=false",
				"--proxy_cert_path=/etc/certs/server.pem",
				"--proxy_cert_key_path=/etc/certs/server_key.pem",
				"--proxy_ca_cert_path=/etc/certs/ca.pem",
				"--service=helloworld:v1,somethingelse",
				"--endpoint_host=localhost",
				"--endpoint_port=9080",
				"--endpoint_type=https",
				"--discovery_adapter=kubernetes",
				"--rules_adapter=kubernetes",
				"--registry_url=http://registry:8080",
				"--registry_token=local",
				"--registry_poll=5s",
				"--kubernetes_url=http://kubernetes:8080",
				"--kubernetes_token=12345",
				"--kubernetes_namespace=default",
				"--kubernetes_pod_name=mypod-a74n3",
				"--eureka_url=http://eureka1:9001",
				"--eureka_url=http://eureka2:9002",
				"--controller_url=http://controller:8080",
				"--controller_token=local",
				"--controller_poll=5s",
				"--dns_port=4056",
				"--dns_domain=someServer",
				"--healthchecks=http://localhost:8082/health1",
				"--healthchecks=http://localhost:8082/health2",
				"--log_level=debug",
				"--discovery_port=9080",
				"python", "productpage.py",
			}...)

			Expect(app.Run(args)).NotTo(HaveOccurred())
		})

		It("uses config values from command line flags", func() {
			Expect(c.Register).To(Equal(true))
			Expect(c.Proxy).To(Equal(true))
			Expect(c.ProxyAdapter).To(Equal(EnvoyAdapter))
			Expect(c.DNS).To(Equal(true))
			Expect(c.ProxyConfig.TLS).To(Equal(false))
			Expect(c.ProxyConfig.CertPath).To(Equal("/etc/certs/server.pem"))
			Expect(c.ProxyConfig.CertKeyPath).To(Equal("/etc/certs/server_key.pem"))
			Expect(c.ProxyConfig.CACertPath).To(Equal("/etc/certs/ca.pem"))
			Expect(c.Service.Name).To(Equal("helloworld"))
			Expect(c.Service.Tags).To(Equal([]string{"v1", "somethingelse"}))
			Expect(c.Endpoint.Host).To(Equal("localhost"))
			Expect(c.Endpoint.Port).To(Equal(9080))
			Expect(c.Endpoint.Type).To(Equal("https"))
			Expect(c.DiscoveryAdapter).To(Equal("kubernetes"))
			Expect(c.RulesAdapter).To(Equal("kubernetes"))
			Expect(c.A8Registry.URL).To(Equal("http://registry:8080"))
			Expect(c.A8Registry.Token).To(Equal("local"))
			Expect(c.A8Registry.Poll).To(Equal(time.Duration(5) * time.Second))
			Expect(c.Kubernetes.URL).To(Equal("http://kubernetes:8080"))
			Expect(c.Kubernetes.Token).To(Equal("12345"))
			Expect(c.Kubernetes.Namespace).To(Equal("default"))
			Expect(c.Kubernetes.PodName).To(Equal("mypod-a74n3"))
			Expect(c.Eureka.URLs).To(And(ContainElement("http://eureka1:9001"), ContainElement("http://eureka2:9002")))
			Expect(c.A8Controller.URL).To(Equal("http://controller:8080"))
			Expect(c.A8Controller.Token).To(Equal("local"))
			Expect(c.A8Controller.Poll).To(Equal(time.Duration(5) * time.Second))
			Expect(c.Dnsconfig.Port).To(Equal(4056))
			Expect(c.Dnsconfig.Domain).To(Equal("someServer"))
			Expect(c.HealthChecks[0].Value).To(Equal("http://localhost:8082/health1"))
			Expect(c.HealthChecks[0].CACertPath).To(Equal(""))
			Expect(c.HealthChecks[1].Value).To(Equal("http://localhost:8082/health2"))
			Expect(c.HealthChecks[1].CACertPath).To(Equal(""))
			Expect(c.LogLevel).To(Equal("debug"))
			Expect(c.DiscoveryPort).To(Equal(9080))
			Expect(c.Commands).To(HaveLen(1))
			Expect(c.Commands[0].OnExit).To(Equal(TerminateProcess))
			Expect(c.Commands[0].Cmd).To(Equal([]string{"python", "productpage.py"}))
		})
	})

	Context("config overidden with environment variables", func() {

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
			os.Setenv("A8_PROXY_ADAPTER", "envoy")
			os.Setenv("A8_DNS", "true")
			os.Setenv("A8_PROXY_TLS", "false")
			os.Setenv("A8_PROXY_CERT_PATH", "/etc/certs/server.pem")
			os.Setenv("A8_PROXY_CERT_KEY_PATH", "/etc/certs/server_key.pem")
			os.Setenv("A8_PROXY_CA_CERT_PATH", "/etc/certs/ca.pem")
			os.Setenv("A8_SERVICE", "helloworld:v1,somethingelse")
			os.Setenv("A8_ENDPOINT_HOST", "localhost")
			os.Setenv("A8_ENDPOINT_PORT", "9080")
			os.Setenv("A8_ENDPOINT_TYPE", "https")
			os.Setenv("A8_DISCOVERY_ADAPTER", "kubernetes")
			os.Setenv("A8_RULES_ADAPTER", "kubernetes")
			os.Setenv("A8_REGISTRY_URL", "http://registry:8080")
			os.Setenv("A8_REGISTRY_TOKEN", "local")
			os.Setenv("A8_REGISTRY_POLL", "5s")
			os.Setenv("A8_KUBERNETES_URL", "http://kubernetes:8080")
			os.Setenv("A8_KUBERNETES_TOKEN", "12345")
			os.Setenv("A8_KUBERNETES_NAMESPACE", "default")
			os.Setenv("A8_KUBERNETES_POD_NAME", "mypod-a74n3")
			os.Setenv("A8_EUREKA_URL", "http://eureka1:9001,http://eureka2:9002")
			os.Setenv("A8_CONTROLLER_URL", "http://controller:8080")
			os.Setenv("A8_CONTROLLER_TOKEN", "local")
			os.Setenv("A8_CONTROLLER_POLL", "5s")
			os.Setenv("A8_DNS_PORT", "4056")
			os.Setenv("A8_DNS_DOMAIN", "someServer")
			os.Setenv("A8_HEALTHCHECKS", "http://localhost:8082/health1,http://localhost:8082/health2")
			os.Setenv("A8_LOG_LEVEL", "debug")
			args := append(os.Args[:1], []string{
				"python", "productpage.py",
			}...)
			Expect(app.Run(args)).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.Unsetenv("A8_REGISTER")
			os.Unsetenv("A8_PROXY")
			os.Unsetenv("A8_PROXY_ADAPTER")
			os.Unsetenv("A8_DNS")
			os.Unsetenv("A8_PROXY_TLS")
			os.Unsetenv("A8_PROXY_CERT_PATH")
			os.Unsetenv("A8_PROXY_CERT_KEY_PATH")
			os.Unsetenv("A8_PROXY_CA_CERT_PATH")
			os.Unsetenv("A8_SERVICE")
			os.Unsetenv("A8_ENDPOINT_HOST")
			os.Unsetenv("A8_ENDPOINT_PORT")
			os.Unsetenv("A8_ENDPOINT_TYPE")
			os.Unsetenv("A8_REGISTRY_URL")
			os.Unsetenv("A8_DISCOVERY_ADAPTER")
			os.Unsetenv("A8_RULES_ADAPTER")
			os.Unsetenv("A8_REGISTRY_TOKEN")
			os.Unsetenv("A8_REGISTRY_POLL")
			os.Unsetenv("A8_KUBERNETES_URL")
			os.Unsetenv("A8_KUBERNETES_TOKEN")
			os.Unsetenv("A8_KUBERNETES_NAMESPACE")
			os.Unsetenv("A8_KUBERNETES_POD_NAME")
			os.Unsetenv("A8_EUREKA_URL")
			os.Unsetenv("A8_CONTROLLER_URL")
			os.Unsetenv("A8_CONTROLLER_TOKEN")
			os.Unsetenv("A8_CONTROLLER_POLL")
			os.Unsetenv("A8_DNS_PORT")
			os.Unsetenv("A8_DNS_DOMAIN")
			os.Unsetenv("A8_HEALTHCHECKS")
			os.Unsetenv("A8_LOG_LEVEL")
		})

		It("uses config values from environment variables", func() {
			Expect(c.Register).To(Equal(true))
			Expect(c.Proxy).To(Equal(true))
			Expect(c.ProxyAdapter).To(Equal(EnvoyAdapter))
			Expect(c.DNS).To(Equal(true))
			Expect(c.ProxyConfig.TLS).To(Equal(false))
			Expect(c.ProxyConfig.CertPath).To(Equal("/etc/certs/server.pem"))
			Expect(c.ProxyConfig.CertKeyPath).To(Equal("/etc/certs/server_key.pem"))
			Expect(c.ProxyConfig.CACertPath).To(Equal("/etc/certs/ca.pem"))
			Expect(c.Service.Name).To(Equal("helloworld"))
			Expect(c.Service.Tags).To(Equal([]string{"v1", "somethingelse"}))
			Expect(c.Endpoint.Host).To(Equal("localhost"))
			Expect(c.Endpoint.Port).To(Equal(9080))
			Expect(c.Endpoint.Type).To(Equal("https"))
			Expect(c.DiscoveryAdapter).To(Equal("kubernetes"))
			Expect(c.RulesAdapter).To(Equal("kubernetes"))
			Expect(c.A8Registry.URL).To(Equal("http://registry:8080"))
			Expect(c.A8Registry.Token).To(Equal("local"))
			Expect(c.A8Registry.Poll).To(Equal(time.Duration(5) * time.Second))
			Expect(c.Kubernetes.URL).To(Equal("http://kubernetes:8080"))
			Expect(c.Kubernetes.Token).To(Equal("12345"))
			Expect(c.Kubernetes.Namespace).To(Equal("default"))
			Expect(c.Kubernetes.PodName).To(Equal("mypod-a74n3"))
			Expect(c.Eureka.URLs).To(And(ContainElement("http://eureka1:9001"), ContainElement("http://eureka2:9002")))
			Expect(c.A8Controller.URL).To(Equal("http://controller:8080"))
			Expect(c.A8Controller.Token).To(Equal("local"))
			Expect(c.A8Controller.Poll).To(Equal(time.Duration(5) * time.Second))
			Expect(c.Dnsconfig.Port).To(Equal(4056))
			Expect(c.Dnsconfig.Domain).To(Equal("someServer"))
			Expect(c.HealthChecks[0].Value).To(Equal("http://localhost:8082/health1"))
			Expect(c.HealthChecks[0].CACertPath).To(Equal(""))
			Expect(c.HealthChecks[1].Value).To(Equal("http://localhost:8082/health2"))
			Expect(c.HealthChecks[1].CACertPath).To(Equal(""))
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
proxy_adapter: envoy
dns: true

proxy_config:
  tls:           true
  cert_path:     /etc/certs/server.pem
  cert_key_path: /etc/certs/server_key.pem
  ca_cert_path:  /etc/certs/ca.pem

service:
  name: helloworld
  tags:
    - v1
    - somethingelse

endpoint:
  host: localhost
  port: 9080
  type: https

discovery_adapter: kubernetes
rules_adapter: kubernetes

registry:
  url:   http://registry:8080
  token: local
  poll:  5s

controller:
  url:   http://controller:8080
  token: local
  poll:  5s

kubernetes:
  url:   http://kubernetes:8080
  token: 12345
  namespace: default
  pod_name: mypod-a74n3

eureka:
  urls:
    - http://eureka1:9001
    - http://eureka2:9002

dnsconfig:
  port:   4056
  domain: someServer

healthchecks:
  - type: http
    value: http://localhost:8082/health1
    interval: 15s
    timeout: 5s
    method: GET
    code: 200
    ca_cert_path: /etc/certs/ca1.pem
  - type: http
    value: http://localhost:8082/health2
    interval: 30s
    timeout: 3s
    method: POST
    code: 201
    ca_cert_path: /etc/certs/ca2.pem

commands:
  - cmd: [ "sleep", "720" ]
    env: [ "GODEBUG=netdns=go" ]
    on_exit: terminate
  - cmd: [ "ls" ]
    on_exit: ignore

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
			Expect(c.ProxyAdapter).To(Equal(EnvoyAdapter))
			Expect(c.DNS).To(Equal(true))
			Expect(c.ProxyConfig.TLS).To(Equal(true))
			Expect(c.ProxyConfig.CertPath).To(Equal("/etc/certs/server.pem"))
			Expect(c.ProxyConfig.CertKeyPath).To(Equal("/etc/certs/server_key.pem"))
			Expect(c.ProxyConfig.CACertPath).To(Equal("/etc/certs/ca.pem"))
			Expect(c.Service.Name).To(Equal("helloworld"))
			Expect(c.Service.Tags).To(Equal([]string{"v1", "somethingelse"}))
			Expect(c.Endpoint.Host).To(Equal("localhost"))
			Expect(c.Endpoint.Port).To(Equal(9080))
			Expect(c.Endpoint.Type).To(Equal("https"))
			Expect(c.DiscoveryAdapter).To(Equal("kubernetes"))
			Expect(c.RulesAdapter).To(Equal("kubernetes"))
			Expect(c.A8Registry.URL).To(Equal("http://registry:8080"))
			Expect(c.A8Registry.Token).To(Equal("local"))
			Expect(c.A8Registry.Poll).To(Equal(time.Duration(5) * time.Second))
			Expect(c.Kubernetes.URL).To(Equal("http://kubernetes:8080"))
			Expect(c.Kubernetes.Token).To(Equal("12345"))
			Expect(c.Kubernetes.Namespace).To(Equal("default"))
			Expect(c.Kubernetes.PodName).To(Equal("mypod-a74n3"))
			Expect(c.Eureka.URLs).To(And(ContainElement("http://eureka1:9001"), ContainElement("http://eureka2:9002")))
			Expect(c.A8Controller.URL).To(Equal("http://controller:8080"))
			Expect(c.A8Controller.Token).To(Equal("local"))
			Expect(c.A8Controller.Poll).To(Equal(time.Duration(5) * time.Second))
			Expect(c.Dnsconfig.Port).To(Equal(4056))
			Expect(c.Dnsconfig.Domain).To(Equal("someServer"))
			Expect(c.HealthChecks[0].Type).To(Equal("http"))
			Expect(c.HealthChecks[0].Value).To(Equal("http://localhost:8082/health1"))
			Expect(c.HealthChecks[0].Interval).To(Equal(time.Duration(15) * time.Second))
			Expect(c.HealthChecks[0].Timeout).To(Equal(time.Duration(5) * time.Second))
			Expect(c.HealthChecks[0].Method).To(Equal("GET"))
			Expect(c.HealthChecks[0].Code).To(Equal(200))
			Expect(c.HealthChecks[0].CACertPath).To(Equal("/etc/certs/ca1.pem"))
			Expect(c.HealthChecks[1].Type).To(Equal("http"))
			Expect(c.HealthChecks[1].Value).To(Equal("http://localhost:8082/health2"))
			Expect(c.HealthChecks[1].Interval).To(Equal(time.Duration(30) * time.Second))
			Expect(c.HealthChecks[1].Timeout).To(Equal(time.Duration(3) * time.Second))
			Expect(c.HealthChecks[1].Method).To(Equal("POST"))
			Expect(c.HealthChecks[1].Code).To(Equal(201))
			Expect(c.HealthChecks[1].CACertPath).To(Equal("/etc/certs/ca2.pem"))
			Expect(c.LogLevel).To(Equal("debug"))
			Expect(c.Commands).To(HaveLen(2))
			Expect(c.Commands[0].OnExit).To(Equal(TerminateProcess))
			Expect(c.Commands[0].Cmd).To(Equal([]string{"sleep", "720"}))
			Expect(c.Commands[0].Env).To(Equal([]string{"GODEBUG=netdns=go"}))
		})
	})

	Context("config validation", func() {

		BeforeEach(func() {
			c = &Config{
				DiscoveryAdapter: Amalgam8Adapter,
				RulesAdapter:     Amalgam8Adapter,
				A8Registry: Amalgam8Registry{
					URL:   "http://registry",
					Token: "sd_token",
					Poll:  60 * time.Second,
				},
				A8Controller: Amalgam8Controller{
					Token: "token",
					URL:   "http://controller",
					Poll:  60 * time.Second,
				},
				Dnsconfig: Dnsconfig{
					Port:   8053,
					Domain: "amalgam8",
				},
				Proxy:        true,
				ProxyAdapter: NGINXAdapter,
				Register:     true,
				DNS:          true,
				Service: Service{
					Name: "mock",
				},
				Endpoint: Endpoint{
					Host: "mockhost",
					Port: 9090,
					Type: "http",
				},
				Commands: []Command{Command{
					Cmd:    []string{"ls"},
					Env:    []string{},
					OnExit: TerminateProcess,
				}},
			}
		})

		It("accepts a valid config", func() {
			Expect(c.Validate()).ToNot(HaveOccurred())
		})

		It("rejects an invalid URL", func() {
			c.A8Controller.URL = "123456"
			Expect(c.Validate()).To(HaveOccurred())
		})

		It("rejects an excessively large poll interval", func() {
			c.A8Controller.Poll = 48 * time.Hour
			Expect(c.Validate()).To(HaveOccurred())
		})

		It("rejects invalid OnExit parameter", func() {
			c.Commands[0].OnExit = "unknown_param"
			Expect(c.Validate()).To(HaveOccurred())
		})

		It("accepts empty OnExit parameter", func() {
			c.Commands[0].OnExit = ""
			Expect(c.Validate()).ToNot(HaveOccurred())
		})

		It("rejects Command with empty command", func() {
			c.Commands[0].Cmd = []string{}
			Expect(c.Validate()).To(HaveOccurred())
		})
	})

})
