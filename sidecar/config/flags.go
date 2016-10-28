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
	configFlag          = "config"
	registerFlag        = "register"
	proxyFlag           = "proxy"
	serviceFlag         = "service"
	endpointHostFlag    = "endpoint_host"
	endpointPortFlag    = "endpoint_port"
	endpointTypeFlag    = "endpoint_type"
	registryTokenFlag   = "registry_token"
	registryURLFlag     = "registry_url"
	registryPollFlag    = "registry_poll"
	controllerURLFlag   = "controller_url"
	controllerTokenFlag = "controller_token"
	controllerPollFlag  = "controller_poll"
	healthchecksFlag    = "healthchecks"
	logLevelFlag        = "log_level"
	debugFlag           = "debug"
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
		Name:   registryURLFlag,
		EnvVar: envVar(registryURLFlag),
		Usage:  "URL for Registry",
	},
	cli.StringFlag{
		Name:   registryTokenFlag,
		EnvVar: envVar(registryTokenFlag),
		Usage:  "API token for Registry",
	},
	cli.DurationFlag{
		Name:   registryPollFlag,
		EnvVar: envVar(registryPollFlag),
		Usage:  "Interval for polling Controller",
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
