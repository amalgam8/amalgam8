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

	"io/ioutil"

	"reflect"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"gopkg.in/yaml.v2"
)

// Service configuration
type Service struct {
	Name string   `yaml:"name"`
	Tags []string `yaml:"tags"`
}

// Endpoint configuration
type Endpoint struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Type string `yaml:"type"`
}

// Registry configuration
type Registry struct {
	URL   string        `yaml:"url"`
	Token string        `yaml:"token"`
	Poll  time.Duration `yaml:"poll"`
}

// Controller configuration
type Controller struct {
	URL   string        `yaml:"url"`
	Token string        `yaml:"token"`
	Poll  time.Duration `yaml:"poll"`
}

// Config stores the various configuration options for the sidecar
type Config struct {
	Register bool `yaml:"register"`
	Proxy    bool `yaml:"proxy"`

	Service  Service  `yaml:"service"`
	Endpoint Endpoint `yaml:"endpoint"`

	Registry   Registry   `yaml:"registry"`
	Controller Controller `yaml:"controller"`

	Supervise bool     `yaml:"supervise"`
	App       []string `yaml:"app"`

	Log            bool   `yaml:"log"`
	LogstashServer string `yaml:"logstash_server"`

	LogLevel string `yaml:"log_level"`
}

// New creates a new Config object from the given commandline flags, environment variables, and configuration file context.
func New(context *cli.Context) (*Config, error) {

	// Initialize configuration with default values
	config := *&DefaultConfig

	// Load configuration from file, if specified
	if context.IsSet(configFlag) {
		err := config.loadFromFile(context.String(configFlag))
		if err != nil {
			return nil, err
		}
	}

	// Load configuration from context (commandline flags and environment variables)
	err := config.loadFromContext(context)
	if err != nil {
		return nil, err
	}

	if config.Endpoint.Host == "" {
		config.Endpoint.Host = waitForLocalIP()
	}

	return &config, nil
}

func (c *Config) loadFromFile(configFile string) error {
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(bytes, c)
	if err != nil {
		return nil
	}

	return nil
}

func (c *Config) loadFromContext(context *cli.Context) error {
	loadFromContextIfSet := func(ptr interface{}, flagName string) {
		if !context.IsSet(flagName) {
			return
		}

		configValue := reflect.ValueOf(ptr).Elem()
		var flagValue interface{}
		switch configValue.Kind() {
		case reflect.Bool:
			flagValue = context.Bool(flagName)
		case reflect.String:
			flagValue = context.String(flagName)
		case reflect.Int:
			flagValue = context.Int(flagName)
		case reflect.Int64:
			flagValue = context.Duration(flagName)
		case reflect.Float64:
			flagValue = context.Float64(flagName)
		case reflect.Slice:
			switch configValue.Type().Elem().Kind() {
			case reflect.String:
				flagValue = context.StringSlice(flagName)
			case reflect.Int:
				flagValue = context.IntSlice(flagName)
			default:
				logrus.Errorf("unsupported configuration type '%v' for '%v'", configValue.Kind(), flagName)
			}
		default:
			logrus.Errorf("unsupported configuration type '%v' for '%v'", configValue.Kind(), flagName)
		}

		configValue.Set(reflect.ValueOf(flagValue))
	}

	loadFromContextIfSet(&c.Register, registerFlag)
	loadFromContextIfSet(&c.Proxy, proxyFlag)
	loadFromContextIfSet(&c.Endpoint.Host, endpointHostFlag)
	loadFromContextIfSet(&c.Endpoint.Port, endpointPortFlag)
	loadFromContextIfSet(&c.Endpoint.Type, endpointTypeFlag)
	loadFromContextIfSet(&c.Registry.URL, registryURLFlag)
	loadFromContextIfSet(&c.Registry.Token, registryTokenFlag)
	loadFromContextIfSet(&c.Registry.Poll, registryPollFlag)
	loadFromContextIfSet(&c.Controller.URL, controllerURLFlag)
	loadFromContextIfSet(&c.Controller.Token, controllerTokenFlag)
	loadFromContextIfSet(&c.Controller.Poll, controllerPollFlag)
	loadFromContextIfSet(&c.Supervise, superviseFlag)
	loadFromContextIfSet(&c.Log, logFlag)
	loadFromContextIfSet(&c.LogstashServer, logstashServerFlag)
	loadFromContextIfSet(&c.LogLevel, logLevelFlag)

	if context.IsSet(serviceFlag) {
		name, tags := parseServiceNameAndTags(context.String(serviceFlag))
		c.Service.Name = name
		c.Service.Tags = tags
	}

	if context.Args().Present() {
		c.App = context.Args()
	}

	return nil
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

// waitForLocalIP waits until a local IP is available
func waitForLocalIP() string {
	for {
		ip := localIP()
		if ip != "" {
			break
		}
		logrus.Warn("Could not obtain local IP")
		time.Sleep(time.Second * 10)
	}
	return ""
}

// localIP retrieves the IP address of the system
func localIP() string {
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

func parseServiceNameAndTags(service string) (name string, tags []string) {
	i := strings.Index(service, ":")
	if i == -1 {
		name = service
		tags = []string{}
	} else {
		name = service[:i]
		tags = strings.Split(service[i+1:], ",")
	}
	return
}
