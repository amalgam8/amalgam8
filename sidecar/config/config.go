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
	"io/ioutil"
	"net"
	"reflect"
	"strings"
	"time"

	"net/url"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

const (
	//TerminateProcess signal that supervisor should kill other processes and exit process on failure
	TerminateProcess = "terminate"

	//IgnoreProcess signal that supervisor should ignore this process on failure
	IgnoreProcess = "ignore"
)

// Command to be managed by sidecar app supervisor
type Command struct {
	Cmd    []string `yaml:"cmd"`
	Env    []string `yaml:"env"`
	OnExit string   `yaml:"on_exit"`
}

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

// Dnsconfig - DNS server configuration
type Dnsconfig struct {
	Port   int    `yaml:"port"`
	Domain string `yaml:"domain"`
}

// Health check types.
const (
	HTTPHealthCheck    = "http"
	HTTPSHealthCheck   = "https"
	TCPHealthCheck     = "tcp"
	CommandHealthCheck = "file"
)

// HealthCheck configuration.
type HealthCheck struct {
	Type     string        `yaml:"type"`
	Value    string        `yaml:"value"`
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
	Method   string        `yaml:"method"`
	Code     int           `yaml:"code"`
	Args     []string      `yaml:"args"`
}

// Config stores the various configuration options for the sidecar
type Config struct {
	Register bool `yaml:"register"`
	Proxy    bool `yaml:"proxy"`
	DNS      bool `yaml:"dns"`

	Service  Service  `yaml:"service"`
	Endpoint Endpoint `yaml:"endpoint"`

	Registry   Registry   `yaml:"registry"`
	Controller Controller `yaml:"controller"`
	Dnsconfig  Dnsconfig  `yaml:"dnsconfig"`

	Supervise bool `yaml:"supervise"`

	HealthChecks []HealthCheck `yaml:"healthchecks"`

	LogLevel string `yaml:"log_level"`

	Commands []Command `yaml:"commands"`

	Debug string
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
		logrus.Infof("No hostname is configured. Using local IP instead...")
		config.Endpoint.Host = waitForLocalIP()
		logrus.Infof("Obtained local IP %s", config.Endpoint.Host)
	}

	return &config, nil
}

func (c *Config) loadFromFile(configFile string) error {
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(bytes, c)
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
	loadFromContextIfSet(&c.DNS, dnsFlag)
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
	loadFromContextIfSet(&c.Dnsconfig.Port, dnsConfigPortFlag)
	loadFromContextIfSet(&c.Dnsconfig.Domain, dnsConfigDomainFlag)
	loadFromContextIfSet(&c.LogLevel, logLevelFlag)
	loadFromContextIfSet(&c.Debug, debugFlag)

	if context.IsSet(serviceFlag) {
		name, tags := parseServiceNameAndTags(context.String(serviceFlag))
		c.Service.Name = name
		c.Service.Tags = tags
	}

	// For health check flags, we only support default values.
	if context.IsSet(healthchecksFlag) {
		hcValues := context.StringSlice(healthchecksFlag)
		for _, hcValue := range hcValues {
			// Parse the healthcheck type from URL scheme
			u, err := url.Parse(hcValue)
			if err != nil {
				return fmt.Errorf("Could not parse healthcheck: '%s'", hcValue)
			}

			var hcType string
			switch u.Scheme {
			case "http":
				hcType = HTTPHealthCheck
			case "https":
				hcType = HTTPSHealthCheck
			case "tcp":
				hcType = TCPHealthCheck
			case "file":
				hcType = CommandHealthCheck
			default:
				return fmt.Errorf("Unsupported health check type: %v", u.Scheme)
			}

			var hc HealthCheck
			switch hcType {
			case CommandHealthCheck:
				hc = HealthCheck{
					Type:  hcType,
					Value: u.Path,
				}
			default:
				hc = HealthCheck{
					Type:  hcType,
					Value: hcValue,
				}
			}

			c.HealthChecks = append(c.HealthChecks, hc)
		}
	}

	if context.Args().Present() {
		cmd := Command{
			Cmd:    context.Args(),
			OnExit: TerminateProcess,
		}
		c.Commands = append(c.Commands, cmd)
	}

	return nil
}

// Validate the configuration
func (c *Config) Validate() error {

	if c.Supervise {
		logrus.Warn("WARNING: --supervise flag is deprecated and may not be supported in the future.")
	}

	if !c.Register && !c.Proxy {
		return errors.New("Sidecar serves no purpose. Please enable either proxy or registry or both")
	}

	// Create list of validation checks
	validators := []ValidatorFunc{}

	validators = append(validators,
		func() error {
			for _, cmd := range c.Commands {
				if cmd.OnExit != "" && (cmd.OnExit != TerminateProcess && cmd.OnExit != IgnoreProcess) {
					return fmt.Errorf("Unrecognized OnExit command '%v'. Supported"+
						" process OnExit types are 'ignore' and 'terminate'", cmd.OnExit)
				}
				if len(cmd.Cmd) == 0 {
					return fmt.Errorf("Invalid command provided for process")
				}
			}
			return nil
		},
	)

	// Registry URL is needed for both proxying and registering.  Registry token is not required in all auth cases
	validators = append(validators, IsValidURL("Registry URL", c.Registry.URL))

	if c.Register {
		// TODO: validate health checks

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

	if c.DNS {
		validators = append(validators,
			IsInRange("Dns Port", c.Dnsconfig.Port, 1, 65535),
			IsValidDomain("Dns Domain", c.Dnsconfig.Domain),
		)
	}

	return Validate(validators)
}

// waitForLocalIP waits until a local IP is available
func waitForLocalIP() string {
	ip := ""
	for {
		ip = localIP()
		if ip != "" {
			break
		}
		logrus.Warn("Could not obtain local IP")
		time.Sleep(time.Second * 10)
	}
	return ip
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
