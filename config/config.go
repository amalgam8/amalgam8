package config

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/sidecar/util"
	"github.com/codegangsta/cli"
	"os"
	"time"
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
	ServiceName  string
	EndpointHost string
	EndpointPort int
	Register     bool
	Proxy        bool
	Log          bool
	Supervise    bool
	Tenant       Tenant
	Controller   Controller
	Registry     Registry
	Kafka        Kafka
	Nginx        Nginx
	LogLevel     logrus.Level
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
		ServiceName:  context.String(serviceName),
		EndpointHost: context.String(endpointHost),
		EndpointPort: context.Int(endpointPort),
		Register:     context.BoolT(register),
		Proxy:        context.BoolT(proxy),
		Log:          context.BoolT(log),
		Supervise:    context.Bool(supervise),
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
	}
}

// Validate the configuration
func (c *Config) Validate() error {

	// Create list of validation checks
	validators := []util.ValidatorFunc{
		util.IsNotEmpty("Service Name", c.ServiceName),
		util.IsNotEmpty("Registry token", c.Registry.Token),
		util.IsValidURL("Regsitry URL", c.Registry.URL),
		func() error {
			if c.Tenant.TTL.Seconds() < c.Tenant.Heartbeat.Seconds() {
				return fmt.Errorf("Tenant TTL (%v) is less than heartbeat interval (%v)", c.Tenant.TTL, c.Tenant.Heartbeat)
			}
			return nil
		},
		util.IsInRange("Tenant port", c.Nginx.Port, 1, 65535),
	}

	if c.Register {
		validators = append(validators,
			util.IsNotEmpty("Service Endpoint Host", c.EndpointHost),
			util.IsInRange("Service Endpoint Port", c.EndpointPort, 1, 65535),
			util.IsInRangeDuration("Tenant TTL", c.Tenant.TTL, 5*time.Second, 1*time.Hour),
			util.IsInRangeDuration("Tenant heartbeat interval", c.Tenant.TTL, 5*time.Second, 1*time.Hour),
		)
	}

	if c.Proxy {
		validators = append(validators,
			util.IsNotEmpty("Tenant ID", c.Tenant.ID),
			util.IsNotEmpty("Tenant token", c.Tenant.Token),
			util.IsInRange("Tenant port", c.Tenant.Port, 1, 65535),
			util.IsValidURL("Controller URL", c.Controller.URL),
			util.IsInRangeDuration("Controller polling interval", c.Controller.Poll, 5*time.Second, 1*time.Hour),
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
					if err := util.IsNotEmpty("Kafka broker", broker)(); err != nil {
						return err
					}
				}
				return nil
			},
		)
		if c.Kafka.SASL {
			validators = append(validators,
				util.IsNotEmpty("Kafka username", c.Kafka.Username),
				util.IsNotEmpty("Kafka password", c.Kafka.Password),
				util.IsNotEmpty("Kafka token", c.Kafka.APIKey),
				util.IsValidURL("Kafka Rest URL", c.Kafka.RestURL),
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

	return util.Validate(validators)
}
