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
	apiPort          = "api_port"
	dbType           = "database_type"
	dbUser           = "database_username"
	dbPassword       = "database_password"
	dbHost           = "database_host"
	secretKey        = "encryption_key"
	pollInterval     = "poll_interval"
	logLevel         = "log_level"
	authModeFlag     = "auth_mode"
	jwtSecretFlag    = "jwt_secret"
	requireHTTPSFlag = "require_https"
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
		Name:   secretKey,
		Value:  "abcdefghijklmnop",
		EnvVar: envVar(secretKey),
		Usage:  "secret key",
	},

	cli.DurationFlag{
		Name:   pollInterval,
		Value:  10 * time.Second,
		EnvVar: envVar(pollInterval),
		Usage:  "poll interval",
	},

	cli.StringFlag{
		Name:   jwtSecretFlag,
		EnvVar: envVar(jwtSecretFlag),
		Usage:  "Secret key for JWT authentication",
	},

	cli.StringSliceFlag{
		Name:   authModeFlag,
		EnvVar: envVar(authModeFlag),
		Usage:  "Authentication modes. Supported values are: 'trusted', 'jwt'",
	},

	cli.BoolFlag{
		Name:   requireHTTPSFlag,
		EnvVar: envVar(requireHTTPSFlag),
		Usage:  "Require clients to use HTTPS for API calls",
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
