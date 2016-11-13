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
	"strings"

	"github.com/urfave/cli"
)

const (
	configFlag              = "config"
	registerFlag            = "register"
	proxyFlag               = "proxy"
	serviceFlag             = "service"
	endpointHostFlag        = "endpoint_host"
	endpointPortFlag        = "endpoint_port"
	endpointTypeFlag        = "endpoint_type"
	registryBackendFlag     = "registry_backend"
	registryURLFlag         = "registry_url"
	registryTokenFlag       = "registry_token"
	registryPollFlag        = "registry_poll"
	kubernetesURLFlag       = "kubernetes_url"
	kubernetesTokenFlag     = "kubernetes_token"
	kubernetesNamespaceFlag = "kubernetes_namespace"
	eurekaURLFlag           = "eureka_url"
	controllerURLFlag       = "controller_url"
	controllerTokenFlag     = "controller_token"
	controllerPollFlag      = "controller_poll"
	superviseFlag           = "supervise"
	healthchecksFlag        = "healthchecks"
	logLevelFlag            = "log_level"
	dnsFlag                 = "dns"
	dnsConfigPortFlag       = "dns_port"
	dnsConfigDomainFlag     = "dns_domain"
	debugFlag               = "debug"
)

// Flags is the set of supported flags
var Flags = []cli.Flag{
	cli.StringFlag{
		Name:  debugFlag,
		Usage: "Check current sidecar state via CLI command",
	},
	cli.StringFlag{
		Name:   configFlag,
		EnvVar: envVar(configFlag),
		Usage:  "Load configuration from file",
	},
	cli.BoolFlag{
		Name:   registerFlag,
		EnvVar: envVar(registerFlag),
		Usage:  "Enable automatic service registration and heartbeat",
	},
	cli.BoolFlag{
		Name:   proxyFlag,
		EnvVar: envVar(proxyFlag),
		Usage:  "Enable automatic service discovery and load balancing across services using NGINX",
	},
	cli.BoolFlag{
		Name:   dnsFlag,
		EnvVar: envVar(dnsFlag),
		Usage:  "Enable DNS server",
	},
	cli.StringFlag{
		Name:   serviceFlag,
		EnvVar: envVar(serviceFlag),
		Usage:  "Service name to register with",
	},
	cli.StringFlag{
		Name:   endpointHostFlag,
		EnvVar: envVar(endpointHostFlag),
		Usage:  "Service endpoint host name (local IP is used if none specified)",
	},
	cli.IntFlag{
		Name:   endpointPortFlag,
		EnvVar: envVar(endpointPortFlag),
		Usage:  "Service endpoint port",
	},
	cli.StringFlag{
		Name:   endpointTypeFlag,
		EnvVar: envVar(endpointTypeFlag),
		Usage:  "Service endpoint type (http, https, tcp, udp, user)",
	},
	cli.StringFlag{
		Name:   registryBackendFlag,
		EnvVar: envVar(registryBackendFlag),
		Usage:  "Registry backend type (amalgam8, kubernetes, eureka)",
	},
	cli.StringFlag{
		Name:   registryURLFlag,
		EnvVar: envVar(registryURLFlag),
		Usage:  "URL for Amalgam8 Registry",
	},
	cli.StringFlag{
		Name:   registryTokenFlag,
		EnvVar: envVar(registryTokenFlag),
		Usage:  "API token for Amalgam8 Registry",
	},
	cli.DurationFlag{
		Name:   registryPollFlag,
		EnvVar: envVar(registryPollFlag),
		Usage:  "Interval for polling Amalgam8 Registry",
	},
	cli.StringFlag{
		Name:   kubernetesURLFlag,
		EnvVar: envVar(kubernetesURLFlag),
		Usage:  "URL for Kubernetes API server",
	},
	cli.StringFlag{
		Name:   kubernetesTokenFlag,
		EnvVar: envVar(kubernetesTokenFlag),
		Usage:  "API token for Kubernetes API server",
	},
	cli.StringFlag{
		Name:   kubernetesNamespaceFlag,
		EnvVar: envVar(kubernetesNamespaceFlag),
		Usage:  "Kubernetes API namespace",
	},
	cli.StringSliceFlag{
		Name:   eurekaURLFlag,
		EnvVar: envVar(eurekaURLFlag),
		Usage:  "List of Eureka server URLs",
	},
	cli.StringFlag{
		Name:   controllerURLFlag,
		EnvVar: envVar(controllerURLFlag),
		Usage:  "URL for Controller service",
	},
	cli.StringFlag{
		Name:   controllerTokenFlag,
		EnvVar: envVar(controllerTokenFlag),
		Usage:  "Amalgam8 controller token",
	},
	cli.DurationFlag{
		Name:   controllerPollFlag,
		EnvVar: envVar(controllerPollFlag),
		Usage:  "Interval for polling Controller",
	},

	cli.StringFlag{
		Name:   dnsConfigPortFlag,
		EnvVar: envVar(dnsConfigPortFlag),
		Usage:  "DNS server port number",
	},
	cli.StringFlag{
		Name:   dnsConfigDomainFlag,
		EnvVar: envVar(dnsConfigDomainFlag),
		Usage:  "DNS server authorization domain name",
	},
	cli.BoolFlag{
		Name:   superviseFlag,
		EnvVar: envVar(superviseFlag),
		Usage:  "Deprecated - this flag is no longer required and will be ignored",
	},
	cli.StringSliceFlag{
		Name:   healthchecksFlag,
		EnvVar: envVar(healthchecksFlag),
		Usage:  "List of health check URLs",
	},
	cli.StringFlag{
		Name:   logLevelFlag,
		EnvVar: envVar(logLevelFlag),
		Usage:  "Logging level (debug, info, warn, error, fatal, panic)",
	},
}

func envVar(name string) string {
	return strings.ToUpper(fmt.Sprintf("%v%v", "A8_", name))
}
