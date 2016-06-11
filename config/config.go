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
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

// Tenant stores tenant configuration
type Tenant struct {
	ID        string
	Token     string
	TTL       time.Duration
	Heartbeat time.Duration
	Port      int
}

// Registry configuration
type Registry struct {
	URL   string
	Token string
}

// Kafka configuration
type Kafka struct {
	Brokers  []string
	Username string
	Password string
	APIKey   string
	RestURL  string
	SASL     bool
}

// Nginx stores NGINX configuration
type Nginx struct {
	Port    int
	Logging bool
}

// Controller configuration
type Controller struct {
	URL  string
	Poll time.Duration
}

// Config TODO
type Config struct {
	ServiceName    string
	EndpointHost   string
	EndpointPort   int
	LogstashServer string
	Register       bool
	Proxy          bool
	Log            bool
	Supervise      bool
	Tenant         Tenant
	Controller     Controller
	Registry       Registry
	Kafka          Kafka
	Nginx          Nginx
	LogLevel       logrus.Level
	AppArgs        []string
}

// New TODO
func New(context *cli.Context) *Config {
	// TODO
	// Read from VCAP JSON object if present?

	mhBrokers := []string{}
	count := 0
	for {
		broker := os.Getenv(fmt.Sprintf("%v%v", "VCAP_SERVICES_MESSAGEHUB_0_CREDENTIALS_KAFKA_BROKERS_SASL_", count))
		if broker == "" {
			break
		}
		mhBrokers = append(mhBrokers, broker)
		count++
	}

	// TODO: parse this more gracefully
	loggingLevel := logrus.DebugLevel
	logLevelArg := context.String(logLevel)
	var err error
	loggingLevel, err = logrus.ParseLevel(logLevelArg)
	if err != nil {
		loggingLevel = logrus.DebugLevel
	}

	return &Config{
		ServiceName:    context.String(serviceName),
		EndpointHost:   context.String(endpointHost),
		EndpointPort:   context.Int(endpointPort),
		LogstashServer: context.String(logstashServer),
		Register:       context.BoolT(register),
		Proxy:          context.BoolT(proxy),
		Log:            context.BoolT(log),
		Supervise:      context.Bool(supervise),
		Controller: Controller{
			URL:  context.String(controllerURL),
			Poll: context.Duration(controllerPoll),
		},
		Tenant: Tenant{
			ID:        context.String(tenantID),
			Token:     context.String(tenantToken),
			TTL:       context.Duration(tenantTTL),
			Heartbeat: context.Duration(tenantHeartbeat),
			Port:      context.Int(tenantPort),
		},
		Registry: Registry{
			URL:   context.String(registryURL),
			Token: context.String(registryToken),
		},
		Kafka: Kafka{
			Username: context.String(kafkaUsername),
			Password: context.String(kafkaPassword),
			APIKey:   context.String(kafkaToken),
			RestURL:  context.String(kafkaRestURL),
			// FIXME brokers handled properly?
			Brokers: mhBrokers,
			SASL:    context.Bool(kafkaSASL),
		},
		Nginx: Nginx{
			Port: context.Int(nginxPort),
		},
		LogLevel: loggingLevel,
		AppArgs: context.Args(),
	}
}

// Validate the configuration
func (c *Config) Validate() error {

	if !c.Register && !c.Proxy {
		return errors.New("Sidecar serves no purpose. Please enable either proxy or registry or both")
	}

	// Create list of validation checks
	validators := []ValidatorFunc{}

	if c.Log {
		validators = append(validators,
			IsNotEmpty("Logstash Host", c.LogstashServer),
		)
	}

	if c.Register {
		validators = append(validators,
			func() error {
				if c.Tenant.TTL.Seconds() < c.Tenant.Heartbeat.Seconds() {
					return fmt.Errorf("Tenant TTL (%v) is less than heartbeat interval (%v)", c.Tenant.TTL, c.Tenant.Heartbeat)
				}
				return nil
			},
			IsNotEmpty("Service Name", c.ServiceName),
			IsNotEmpty("Registry token", c.Registry.Token),
			IsValidURL("Regsitry URL", c.Registry.URL),
			IsInRange("Tenant port", c.Nginx.Port, 1, 65535),
			IsNotEmpty("Service Endpoint Host", c.EndpointHost),
			IsInRange("Service Endpoint Port", c.EndpointPort, 1, 65535),
			IsInRangeDuration("Tenant TTL", c.Tenant.TTL, 5*time.Second, 1*time.Hour),
			IsInRangeDuration("Tenant heartbeat interval", c.Tenant.TTL, 5*time.Second, 1*time.Hour),
		)
	}

	if c.Proxy {
		validators = append(validators,
			IsNotEmpty("Tenant ID", c.Tenant.ID),
			IsNotEmpty("Tenant token", c.Tenant.Token),
			IsInRange("Tenant port", c.Tenant.Port, 1, 65535),
			IsValidURL("Controller URL", c.Controller.URL),
			IsInRangeDuration("Controller polling interval", c.Controller.Poll, 5*time.Second, 1*time.Hour),
		)
	}

	// If any of the Message Hub config is present validate the Message Hub config
	if len(c.Kafka.Brokers) > 0 || c.Kafka.Username != "" || c.Kafka.Password != "" {
		validators = append(validators,
			func() error {
				if len(c.Kafka.Brokers) == 0 {
					return errors.New("Kafka requires at least one broker")
				}

				for _, broker := range c.Kafka.Brokers {
					if err := IsNotEmpty("Kafka broker", broker)(); err != nil {
						return err
					}
				}
				return nil
			},
		)
		if c.Kafka.SASL {
			validators = append(validators,
				IsNotEmpty("Kafka username", c.Kafka.Username),
				IsNotEmpty("Kafka password", c.Kafka.Password),
				IsNotEmpty("Kafka token", c.Kafka.APIKey),
				IsValidURL("Kafka Rest URL", c.Kafka.RestURL),
			)
		} else {
			validators = append(validators,
				func() error {
					if len(c.Kafka.Brokers) != 0 {
						if c.Kafka.Username != "" || c.Kafka.Password != "" ||
							c.Kafka.RestURL != "" || c.Kafka.APIKey != "" {
							return errors.New("Kafka credentials provided when SASL authentication disabled")
						}
					}

					return nil
				},
			)
		}
	}

	return Validate(validators)
}
