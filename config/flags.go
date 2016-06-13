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
	"github.com/codegangsta/cli"
)

const (
	register        = "register"
	proxy           = "proxy"
	log             = "log"
	supervise       = "supervise"
	tenantID        = "tenant_id"
	tenantToken     = "tenant_token"
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
	tenantTTL       = "tenant_ttl"
	tenantHeartbeat = "tenant_heartbeat"
	endpointHost    = "endpoint_host"
	endpointPort    = "endpoint_port"
	serviceName     = "service"
	serviceVersion  = "service_version"
	logLevel        = "log_level"
	logstashServer  = "logstash_server"
)

// TenantFlags defines all expected args for Tenant
var TenantFlags = []cli.Flag{
	cli.StringFlag{
		Name:   logLevel,
		EnvVar: strings.ToUpper(logLevel),
		Value:  "info",
		Usage:  "Logging level (debug, info, warn, error, fatal, panic)",
	},

	cli.StringFlag{
		Name:   serviceName,
		EnvVar: strings.ToUpper(serviceName),
		Usage:  "Service name to register with",
	},
	cli.StringFlag{
		Name:   serviceVersion,
		EnvVar: strings.ToUpper(serviceVersion),
		Usage:  "Service version to register with",
	},
	cli.StringFlag{
		Name:   endpointHost,
		EnvVar: strings.ToUpper(endpointHost),
		Usage:  "Service endpoint host name",
	},
	cli.IntFlag{
		Name:   endpointPort,
		EnvVar: strings.ToUpper(endpointPort),
		Usage:  "Service endpoint port",
	},

	cli.BoolTFlag{
		Name:   register,
		EnvVar: strings.ToUpper(register),
		Usage:  "Enable automatic service registration and heartbeat",
	},

	cli.BoolTFlag{
		Name:   proxy,
		EnvVar: strings.ToUpper(proxy),
		Usage:  "Enable automatic service discovery and load balancing across services using NGINX",
	},

	cli.BoolTFlag{
		Name:   log,
		EnvVar: strings.ToUpper(log),
		Usage:  "Enable logging of outgoing requests through proxy using FileBeat",
	},

	cli.BoolFlag{
		Name:   supervise,
		EnvVar: strings.ToUpper(supervise),
		Usage:  "Enable monitoring of application process. If application dies, container is killed as well. This has to be the last flag. All arguments provided after this flag will considered as part of the application invocation",
	},

	// Tenant
	cli.StringFlag{
		Name:   tenantID,
		EnvVar: strings.ToUpper(tenantID),
		Usage:  "Service Proxy instance GUID",
	},
	cli.StringFlag{
		Name:   tenantToken,
		EnvVar: strings.ToUpper(tenantToken),
		Usage:  "Token for Service Proxy instance",
	},
	cli.DurationFlag{
		Name:   tenantTTL,
		EnvVar: strings.ToUpper(tenantTTL),
		Value:  time.Duration(time.Minute),
		Usage:  "Tenant TTL for Registry",
	},
	cli.DurationFlag{
		Name:   tenantHeartbeat,
		EnvVar: strings.ToUpper(tenantHeartbeat),
		Value:  time.Duration(time.Second * 45),
		Usage:  "Tenant heartbeat interval to Registry",
	},

	// Registry
	cli.StringFlag{
		Name:   registryURL,
		EnvVar: strings.ToUpper(registryURL),
		Usage:  "URL for Registry",
	},
	cli.StringFlag{
		Name:   registryToken,
		EnvVar: strings.ToUpper(registryToken),
		Usage:  "API token for Regsitry",
	},

	// NGINX
	cli.IntFlag{
		Name:   nginxPort,
		EnvVar: strings.ToUpper(nginxPort),
		Value:  6379,
		Usage:  "Port for NGINX",
	},

	// Controller
	cli.StringFlag{
		Name:   controllerURL,
		EnvVar: strings.ToUpper(controllerURL),
		Usage:  "URL for Controller service",
	},
	cli.DurationFlag{
		Name:   controllerPoll,
		EnvVar: strings.ToUpper(controllerPoll),
		Value:  time.Duration(15 * time.Second),
		Usage:  "Interval for polling Controller",
	},

	// Logserver
	cli.StringFlag{
		Name:   logstashServer,
		EnvVar: strings.ToUpper(logstashServer),
		Usage:  "Logstash target for nginx logs",
	},

	// Kafka
	cli.StringFlag{
		Name:   kafkaUsername,
		EnvVar: strings.ToUpper(kafkaUsername),
		Usage:  "Username for Kafka service",
	},
	cli.StringFlag{
		Name:   kafkaPassword,
		EnvVar: strings.ToUpper(kafkaPassword),
		Usage:  "Password for Kafka service",
	},
	cli.StringFlag{
		Name:   kafkaToken,
		EnvVar: strings.ToUpper(kafkaToken),
		Usage:  "Token for Kafka service",
	},
	cli.StringFlag{
		Name:   kafkaAdminURL,
		EnvVar: strings.ToUpper(kafkaAdminURL),
		Usage:  "Admin URL for Kafka service",
	},
	cli.StringFlag{
		Name:   kafkaRestURL,
		EnvVar: strings.ToUpper(kafkaRestURL),
		Usage:  "REST URL for Kafka service",
	},
	cli.BoolFlag{
		Name:   kafkaSASL,
		EnvVar: strings.ToUpper(kafkaSASL),
		Usage:  "Use SASL/PLAIN authentication for Kafka",
	},
	cli.StringSliceFlag{
		Name: kafkaBrokers,
		EnvVar: strings.ToUpper(kafkaBrokers),
		Usage: "Kafka broker",
	},
}
