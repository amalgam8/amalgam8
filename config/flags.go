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
		EnvVar: strings.ToUpper(apiPort),
		Value:  6379,
		Usage:  "API port",
	},

	cli.StringFlag{
		Name:  statsdHost,
		Value: "127.0.0.1:8125",
		Usage: "Statsd host",
	},

	cli.StringFlag{
		Name:   controlToken,
		Value:  "ABCDEFGHIJKLMNOP",
		EnvVar: strings.ToUpper(controlToken),
		Usage:  "Token for control plane API authentication",
	},

	cli.StringFlag{
		Name:   secretKey,
		Value:  "abcdefghijklmnop",
		EnvVar: strings.ToUpper(secretKey),
		Usage:  "Secret key",
	},

	cli.DurationFlag{
		Name:   pollInterval,
		EnvVar: strings.ToUpper(pollInterval),
		Usage:  "Poll interval (optional)",
	},

	// Database
	cli.StringFlag{
		Name:   dbType,
		EnvVar: strings.ToUpper(dbType),
		Value:  "memory",
		Usage:  "Database type: memory or cloudant",
	},
	cli.StringFlag{
		Name:   dbUser,
		EnvVar: strings.ToUpper(dbUser),
		Usage:  "Username for Database",
	},
	cli.StringFlag{
		Name:   dbPassword,
		EnvVar: strings.ToUpper(dbPassword),
		Usage:  "Password for Database",
	},
	cli.StringFlag{
		Name:   dbHost,
		EnvVar: strings.ToUpper(dbHost),
		Usage:  "Host for Database",
	},

	cli.StringFlag{
		Name:   logLevel,
		EnvVar: strings.ToUpper(logLevel),
		Value:  "info",
		Usage:  "Logging level (debug, info, warn, error, fatal, panic)",
	},
}
