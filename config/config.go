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
	"time"

	"net"

	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

// Service configuration
type Service struct {
	Name string
	Tags []string
}

// Endpoint configuration
type Endpoint struct {
	Host string
	Port int
	Type string
}

// Registry configuration
type Registry struct {
	URL   string
	Token string
	Poll  time.Duration
}

// Controller configuration
type Controller struct {
	URL   string
	Token string
	Poll  time.Duration
}

// Config TODO
type Config struct {
	Register bool
	Proxy    bool

	Service  Service
	Endpoint Endpoint

	Registry   Registry
	Controller Controller

	Supervise bool
	App       []string

	Log            bool
	LogstashServer string

	LogLevel logrus.Level
}

// New TODO
func New(context *cli.Context) *Config {

	// TODO: parse this more gracefully
	loggingLevel := logrus.DebugLevel
	logLevelArg := context.String(logLevelFlag)
	var err error
	loggingLevel, err = logrus.ParseLevel(logLevelArg)
	if err != nil {
		loggingLevel = logrus.DebugLevel
	}

	endpointHost := context.String(endpointHostFlag)
	if endpointHost == "" {
		for {
			endpointHost = LocalIP()
			if endpointHost != "" {
				break
			}
			logrus.Warn("Could not obtain local IP")
			time.Sleep(time.Second * 10)
		}
	}

	var name string
	var tags []string

	i := strings.Index(context.String(serviceFlag), ":")
	if i == -1 {
		name = context.String(serviceFlag)
		tags = []string{}
	} else {
		name = context.String(serviceFlag)[:i]

		tagsString := context.String(serviceFlag)[i+1:]
		tags = strings.Split(tagsString, ",")
	}

	return &Config{
		Register: context.BoolT(registerFlag),
		Proxy:    context.BoolT(proxyFlag),

		Service: Service{
			Name: name,
			Tags: tags,
		},
		Endpoint: Endpoint{
			Host: endpointHost,
			Port: context.Int(endpointPortFlag),
			Type: context.String(endpointTypeFlag),
		},

		Registry: Registry{
			URL:   context.String(registryURLFlag),
			Token: context.String(registryTokenFlag),
			Poll:  context.Duration(registryPollFlag),
		},
		Controller: Controller{
			URL:   context.String(controllerURLFlag),
			Poll:  context.Duration(controllerPollFlag),
			Token: context.String(controllerTokenFlag),
		},

		Supervise: context.Bool(superviseFlag),
		App:       context.Args(),

		Log:            context.BoolT(logFlag),
		LogstashServer: context.String(logstashServerFlag),

		LogLevel: loggingLevel,
	}
}

// Validate the configuration
func (c *Config) Validate() error {

	if !c.Register && !c.Proxy {
		return errors.New("Sidecar serves no purpose. Please enable either proxy or registry or both")
	}

	// Create list of validation checks
	validators := []ValidatorFunc{}

	if c.Supervise {
		validators = append(validators,
			func() error {
				if len(c.App) == 0 {
					return fmt.Errorf("Supervision mode requires application launch arguments")
				}
				return nil
			},
		)
	}

	if c.Log {
		validators = append(validators,
			IsNotEmpty("Logstash Host", c.LogstashServer),
		)
	}

	// Registry URL is needed for both proxying and registering.  Registry token is not required in all auth cases
	validators = append(validators, IsValidURL("Registry URL", c.Registry.URL))

	if c.Register {
		validators = append(validators,
			IsNotEmpty("Service Name", c.Service.Name),
			IsInRange("Service Endpoint Port", c.Endpoint.Port, 1, 65535),
			IsInSet("Service Endpoint Type", c.Endpoint.Type, []string{"http", "https", "tcp", "udp", "user"}),
		)
	}

	if c.Proxy {
		validators = append(validators,
			IsValidURL("Controller URL", c.Controller.URL),
			IsInRangeDuration("Controller polling interval", c.Controller.Poll, 5*time.Second, 1*time.Hour),
		)

	}

	return Validate(validators)
}

// LocalIP retrieves the IP address of the sidecar
func LocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, address := range addrs {
		// check the address type and if it is not a loopback return it
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}

	return ""
}
