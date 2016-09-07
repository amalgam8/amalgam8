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
	apiPortFlag      = "api_port"
	dbTypeFlag       = "database_type"
	dbUserFlag       = "database_username"
	dbPasswordFlag   = "database_password"
	dbHostFlag       = "database_host"
	secretKeyFlag    = "encryption_key"
	logLevelFlag     = "log_level"
	authModeFlag     = "auth_mode"
	jwtSecretFlag    = "jwt_secret"
	requireHTTPSFlag = "require_https"
)

const apiPort = 6379

// Flags command line args for Controller
var Flags = []cli.Flag{

	cli.IntFlag{
		Name:   apiPortFlag,
		EnvVar: envVar(apiPortFlag),
		Value:  apiPort,
		Usage:  "API port",
	},

	cli.StringFlag{
		Name:   secretKeyFlag,
		Value:  "abcdefghijklmnop",
		EnvVar: envVar(secretKeyFlag),
		Usage:  "secret key",
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
		Name:   dbTypeFlag,
		EnvVar: envVar(dbTypeFlag),
		Value:  "memory",
		Usage:  "database type",
	},
	cli.StringFlag{
		Name:   dbUserFlag,
		EnvVar: envVar(dbUserFlag),
		Usage:  "database username",
	},
	cli.StringFlag{
		Name:   dbPasswordFlag,
		EnvVar: envVar(dbPasswordFlag),
		Usage:  "database password",
	},
	cli.StringFlag{
		Name:   dbHostFlag,
		EnvVar: envVar(dbHostFlag),
		Usage:  "database host",
	},

	cli.StringFlag{
		Name:   logLevelFlag,
		EnvVar: envVar(logLevelFlag),
		Value:  "info",
		Usage:  "logging level (debug, info, warn, error, fatal, panic)",
	},
}

func envVar(name string) string {
	return "A8_" + strings.ToUpper(name)
}
