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
	"strings"
	"time"

	"fmt"

	"github.com/codegangsta/cli"
)

const (
	register        = "register"
	proxy           = "proxy"
	log             = "log"
	supervise       = "supervise"
	kafkaToken      = "kafka_token"
	kafkaUsername   = "kafka_user"
	kafkaPassword   = "kafka_pass"
	kafkaBrokers    = "kafka_broker"
	kafkaRestURL    = "kafka_rest_url"
	kafkaAdminURL   = "kafka_admin_url"
	kafkaSASL       = "kafka_sasl"
	registryToken   = "registry_token"
	registryURL     = "registry_url"
	nginxPort       = "nginx_port"
	controllerURL   = "controller_url"
	controllerPoll  = "controller_poll"
	controllerToken = "controller_token"
	tenantTTL       = "tenant_ttl"
	tenantHeartbeat = "tenant_heartbeat"
	endpointHost    = "endpoint_host"
	endpointPort    = "endpoint_port"
	endpointType    = "endpoint_type"
	serviceName     = "service"
	logLevel        = "log_level"
	logstashServer  = "logstash_server"
)

// TenantFlags defines all expected args for Tenant
var TenantFlags = []cli.Flag{
	cli.StringFlag{
		Name:   logLevel,
		EnvVar: envVar(logLevel),
		Value:  "info",
		Usage:  "Logging level (debug, info, warn, error, fatal, panic)",
	},

	cli.StringFlag{
		Name:   serviceName,
		EnvVar: envVar(serviceName),
		Usage:  "Service name to register with",
	},
	cli.StringFlag{
		Name:   endpointHost,
		EnvVar: envVar(endpointHost),
		Usage:  "Service endpoint host name (local IP is used if none specified)",
	},
	cli.IntFlag{
		Name:   endpointPort,
		EnvVar: envVar(endpointPort),
		Usage:  "Service endpoint port",
	},
	cli.StringFlag{
		Name:   endpointType,
		EnvVar: envVar(endpointType),
		Usage:  "Service endpoint type (http, https, tcp, udp, user)",
		Value:  "http",
	},

	cli.BoolTFlag{
		Name:   register,
		EnvVar: envVar(register),
		Usage:  "Enable automatic service registration and heartbeat",
	},

	cli.BoolTFlag{
		Name:   proxy,
		EnvVar: envVar(proxy),
		Usage:  "Enable automatic service discovery and load balancing across services using NGINX",
	},

	cli.BoolTFlag{
		Name:   log,
		EnvVar: envVar(log),
		Usage:  "Enable logging of outgoing requests through proxy using FileBeat",
	},

	cli.BoolFlag{
		Name:   supervise,
		EnvVar: envVar(supervise),
		Usage:  "Enable monitoring of application process. If application dies, container is killed as well. This has to be the last flag. All arguments provided after this flag will considered as part of the application invocation",
	},

	// Tenant
	cli.DurationFlag{
		Name:   tenantTTL,
		EnvVar: envVar(tenantTTL),
		Value:  time.Duration(time.Minute),
		Usage:  "Tenant TTL for Registry",
	},
	cli.DurationFlag{
		Name:   tenantHeartbeat,
		EnvVar: envVar(tenantHeartbeat),
		Value:  time.Duration(time.Second * 45),
		Usage:  "Tenant heartbeat interval to Registry",
	},

	// Registry
	cli.StringFlag{
		Name:   registryURL,
		EnvVar: envVar(registryURL),
		Usage:  "URL for Registry",
	},
	cli.StringFlag{
		Name:   registryToken,
		EnvVar: envVar(registryToken),
		Usage:  "API token for Regsitry",
	},

	// NGINX
	cli.IntFlag{
		Name:   nginxPort,
		EnvVar: envVar(nginxPort),
		Value:  6379,
		Usage:  "Port for NGINX",
	},

	// Controller
	cli.StringFlag{
		Name:   controllerURL,
		EnvVar: envVar(controllerURL),
		Usage:  "URL for Controller service",
	},
	cli.DurationFlag{
		Name:   controllerPoll,
		EnvVar: envVar(controllerPoll),
		Value:  time.Duration(15 * time.Second),
		Usage:  "Interval for polling Controller",
	},
	cli.StringFlag{
		Name:   controllerToken,
		EnvVar: envVar(controllerToken),
		Usage:  "Amalgam8 controller token",
	},

	// Logserver
	cli.StringFlag{
		Name:   logstashServer,
		EnvVar: envVar(logstashServer),
		Usage:  "Logstash target for nginx logs",
	},

	// Kafka
	cli.StringFlag{
		Name:   kafkaUsername,
		EnvVar: envVar(kafkaUsername),
		Usage:  "Username for Kafka service",
	},
	cli.StringFlag{
		Name:   kafkaPassword,
		EnvVar: envVar(kafkaPassword),
		Usage:  "Password for Kafka service",
	},
	cli.StringFlag{
		Name:   kafkaToken,
		EnvVar: envVar(kafkaToken),
		Usage:  "Token for Kafka service",
	},
	cli.StringFlag{
		Name:   kafkaAdminURL,
		EnvVar: envVar(kafkaAdminURL),
		Usage:  "Admin URL for Kafka service",
	},
	cli.StringFlag{
		Name:   kafkaRestURL,
		EnvVar: envVar(kafkaRestURL),
		Usage:  "REST URL for Kafka service",
	},
	cli.BoolFlag{
		Name:   kafkaSASL,
		EnvVar: envVar(kafkaSASL),
		Usage:  "Use SASL/PLAIN authentication for Kafka",
	},
	cli.StringSliceFlag{
		Name:   kafkaBrokers,
		EnvVar: envVar(kafkaBrokers),
		Usage:  "Kafka broker",
	},
}

func envVar(name string) string {
	return strings.ToUpper(fmt.Sprintf("%v%v", "A8_", name))
}
