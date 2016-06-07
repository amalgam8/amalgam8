package config

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/util"
	"github.com/codegangsta/cli"
	"time"
)

// Database config
type Database struct {
	Type     string
	Username string
	Password string
	Host     string
	//URL string
}

// Config for the controller
type Config struct {
	Database     Database
	APIPort      int
	ControlToken string
	SecretKey    string
	StatsdHost   string
	PollInterval time.Duration
	LogLevel     logrus.Level
}

// New config instance
func New(context *cli.Context) *Config {
	// TODO: add HTTPS to database host if not present?

	// TODO: parse this more gracefully
	loggingLevel := logrus.DebugLevel
	logLevelArg := context.String(logLevel)
	var err error
	loggingLevel, err = logrus.ParseLevel(logLevelArg)
	if err != nil {
		loggingLevel = logrus.DebugLevel
	}

	return &Config{
		Database: Database{
			Type:     context.String(dbType),
			Username: context.String(dbUser),
			Password: context.String(dbPassword),
			Host:     "https://" + context.String(dbHost), // FIXME: conditionally add HTTPS
		},
		APIPort:      context.Int(apiPort),
		ControlToken: context.String(controlToken),
		SecretKey:    context.String(secretKey),
		StatsdHost:   context.String(statsdHost),
		PollInterval: context.Duration(pollInterval),
		LogLevel:     loggingLevel,
	}
}

// Validate the config
func (c *Config) Validate() error {
	// Create list of validation checks
	validators := []util.ValidatorFunc{
		util.IsNotEmpty("Database type", c.Database.Type),
		util.IsInRange("API port", c.APIPort, 1, 65535),
		util.IsNotEmpty("Control token", c.ControlToken),
		util.IsNotEmpty("Secret key", c.SecretKey),
		util.IsNotEmpty("Statsd host", c.StatsdHost),
	}

	if c.Database.Type == "cloudant" {
		additionalValidators := []util.ValidatorFunc{
			util.IsNotEmpty("Database password", c.Database.Password),
			util.IsNotEmpty("Database username", c.Database.Username),
			util.IsNotEmpty("Database host name", c.Database.Host),
		}
		validators = append(validators, additionalValidators...)
	} else if c.Database.Type != "memory" {
		return fmt.Errorf("Invalid database type %v", c.Database.Type)
	}

	if len(c.SecretKey) != 16 {
		return fmt.Errorf("Secret must have a length of 16 characters")
	}

	return util.Validate(validators)
}
