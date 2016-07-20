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

	"github.com/codegangsta/cli"
)

const (
	apiPort      = "api_port"
	dbType       = "database_type"
	dbUser       = "database_username"
	dbPassword   = "database_password"
	dbHost       = "database_host"
	secretKey    = "encryption_key"
	controlToken = "control_token"
	statsdHost   = "statsd_host"
	pollInterval = "poll_interval"
	logLevel     = "log_level"
)

// Flags command line args for Controller
var Flags = []cli.Flag{

	cli.IntFlag{
		Name:   apiPort,
		EnvVar: envVar(apiPort),
		Value:  6379,
		Usage:  "API port",
	},

	cli.StringFlag{
		Name:   controlToken,
		Value:  "ABCDEFGHIJKLMNOP",
		EnvVar: envVar(controlToken),
		Usage:  "controller API authentication token",
	},

	cli.StringFlag{
		Name:   secretKey,
		Value:  "abcdefghijklmnop",
		EnvVar: envVar(secretKey),
		Usage:  "secret key",
	},

	cli.DurationFlag{
		Name:   pollInterval,
		EnvVar: envVar(pollInterval),
		Usage:  "poll interval",
	},

	// Database
	cli.StringFlag{
		Name:   dbType,
		EnvVar: envVar(dbType),
		Value:  "memory",
		Usage:  "database type",
	},
	cli.StringFlag{
		Name:   dbUser,
		EnvVar: envVar(dbUser),
		Usage:  "database username",
	},
	cli.StringFlag{
		Name:   dbPassword,
		EnvVar: envVar(dbPassword),
		Usage:  "database password",
	},
	cli.StringFlag{
		Name:   dbHost,
		EnvVar: envVar(dbHost),
		Usage:  "database host",
	},

	cli.StringFlag{
		Name:   logLevel,
		EnvVar: envVar(logLevel),
		Value:  "info",
		Usage:  "logging level (debug, info, warn, error, fatal, panic)",
	},
}

func envVar(name string) string {
	return "A8_" + strings.ToUpper(name)
}
