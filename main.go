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

package main

import (
	"io/ioutil"
	"os"
	"strings"
	"time"

	"fmt"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/registry/client"
	"github.com/amalgam8/sidecar/config"
	"github.com/amalgam8/sidecar/register"
	"github.com/amalgam8/sidecar/router/checker"
	"github.com/amalgam8/sidecar/router/clients"
	"github.com/amalgam8/sidecar/router/nginx"
	"github.com/amalgam8/sidecar/supervisor"
	"github.com/codegangsta/cli"
)

func main() {
	// Initial logging until we parse the user provided log_level arg
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(os.Stderr)

	app := cli.NewApp()

	app.Name = "sidecar"
	app.Usage = "Amalgam8 Sidecar"
	app.Version = "0.2.0"
	app.Flags = config.TenantFlags
	app.Action = sidecarCommand

	err := app.Run(os.Args)
	if err != nil {
		logrus.WithError(err).Error("Failure running main")
	}
}

func sidecarCommand(context *cli.Context) {
	conf := config.New(context)
	if err := sidecarMain(*conf); err != nil {
		logrus.WithError(err).Error("Setup failed")
	}
}

func sidecarMain(conf config.Config) error {
	var err error

	logrus.SetLevel(conf.LogLevel)

	if err = conf.Validate(false); err != nil {
		logrus.WithError(err).Error("Validation of config failed")
		return err
	}

	if conf.Log {
		//Replace the LOGSTASH_REPLACEME string in filebeat.yml with
		//the value provided by the user

		//TODO: Make this configurable
		filebeatConf := "/etc/filebeat/filebeat.yml"
		filebeat, err := ioutil.ReadFile(filebeatConf)
		if err != nil {
			logrus.WithError(err).Error("Could not read filebeat conf")
			return err
		}

		fileContents := strings.Replace(string(filebeat), "LOGSTASH_REPLACEME", conf.LogstashServer, -1)

		err = ioutil.WriteFile("/tmp/filebeat.yml", []byte(fileContents), 0)
		if err != nil {
			logrus.WithError(err).Error("Could not write filebeat conf")
			return err
		}

		// TODO: Log failure?
		go supervisor.DoLogManagement("/tmp/filebeat.yml")
	}

	if conf.Proxy {
		if err = startProxy(&conf); err != nil {
			logrus.WithError(err).Error("Could not start proxy")
		}
	}

	if conf.Register {
		if err = conf.Validate(true); err != nil {
			logrus.WithError(err).Error("Validation of config failed")
			return err
		}
		logrus.Info("Registering")

		registryClient, err := client.New(client.Config{
			URL:       conf.Registry.URL,
			AuthToken: conf.Registry.Token,
		})
		if err != nil {
			logrus.WithError(err).Error("Could not create registry client")
			return err
		}

		address := fmt.Sprintf("%v:%v", conf.EndpointHost, conf.EndpointPort)
		serviceInstance := &client.ServiceInstance{
			ServiceName: conf.ServiceName,
			Endpoint: client.ServiceEndpoint{
				Type:  conf.EndpointType,
				Value: address,
			},
			TTL: 60,
		}

		if conf.ServiceVersion != "" {
			data, err := json.Marshal(map[string]string{"version": conf.ServiceVersion})
			if err == nil {
				serviceInstance.Metadata = data
			} else {
				logrus.WithError(err).Warn("Could not marshal service version metadata")
			}
		}

		agent, err := register.NewRegistrationAgent(register.RegistrationConfig{
			Client:          registryClient,
			ServiceInstance: serviceInstance,
		})
		if err != nil {
			logrus.WithError(err).Error("Could not create registry agent")
			return err
		}

		agent.Start()
	}

	if conf.Supervise {
		supervisor.DoAppSupervision(conf.AppArgs)
	} else {
		select {}
	}

	return nil
}

func startProxy(conf *config.Config) error {
	var err error

	configBytes, err := ioutil.ReadFile("/etc/nginx/amalgam8.conf")
	if err != nil {
		logrus.WithError(err).Error("Missing /etc/nginx/amalgam8.conf")
		return err
	}

	configStr := string(configBytes)
	configStr = strings.Replace(configStr, "__SERVICE_NAME__", conf.ServiceName, -1)

	output, err := os.OpenFile("/etc/nginx/amalgam8.conf", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		logrus.WithError(err).Error("Couldn't open /etc/nginx/amalgam8.conf file for editing")
		return err
	}

	// Write the config
	fmt.Fprintf(output, configStr)
	output.Close()

	rc := clients.NewController(conf)
	nc := clients.NewNGINXClient("http://localhost:5813")

	nginx, err := nginx.NewNginx(
		nginx.Conf{
			ServiceName: conf.ServiceName,
			Service:     nginx.NewService(),
			Config:      nginx.NewConfig(),
			NGINXClient: nc,
		},
	)

	if err != nil {
		logrus.WithError(err).Error("Failed to initialize NGINX object")
		return err
	}

	err = checkIn(rc, conf)
	if err != nil {
		logrus.WithError(err).Error("Check in failed")
		return err
	}

	// for Kafka enabled tenants we should do both polling and listening
	if len(conf.Kafka.Brokers) != 0 {
		go func() {
			time.Sleep(time.Second * 10)
			logrus.Info("Attempting to connect to Kafka")
			var consumer checker.Consumer
			for {
				consumer, err = checker.NewConsumer(checker.ConsumerConfig{
					Brokers:     conf.Kafka.Brokers,
					Username:    conf.Kafka.Username,
					Password:    conf.Kafka.Password,
					ClientID:    conf.Kafka.APIKey,
					Topic:       "A8_NewRules",
					SASLEnabled: conf.Kafka.SASL,
				})
				if err != nil {
					logrus.WithError(err).Error("Could not connect to Kafka, trying again . . .")
					time.Sleep(time.Second * 5) // TODO: exponential falloff?
				} else {
					break
				}
			}
			logrus.Info("Successfully connected to Kafka")

			listener := checker.NewListener(conf, consumer, rc, nginx)

			// listen to Kafka indefinitely
			if err := listener.Start(); err != nil {
				logrus.WithError(err).Error("Could not listen to Kafka")
			}
		}()
	}

	poller := checker.NewPoller(conf, rc, nginx)
	go func() {
		if err = poller.Start(); err != nil {
			logrus.WithError(err).Error("Could not poll Controller")
		}
	}()

	return nil
}

func getCredentials(controller clients.Controller) (clients.TenantCredentials, error) {

	for {
		creds, err := controller.GetCredentials()
		if err != nil {
			if isRetryable(err) {
				time.Sleep(time.Second * 5)
				continue
			} else {
				return creds, err
			}
		}

		return creds, err
	}
}

func registerWithProxy(controller clients.Controller, confNotValidErr error) error {
	if confNotValidErr != nil {
		// Config not valid, can't register
		logrus.WithError(confNotValidErr).Error("Validation of config failed")
		return confNotValidErr
	}

	for {
		err := controller.Register()
		if err != nil {
			if isRetryable(err) {
				time.Sleep(time.Second * 5)
				continue
			} else {
				return err
			}
		}

		return err
	}
}

func checkIn(controller clients.Controller, conf *config.Config) error {

	confNotValidErr := conf.Validate(true)

	creds, err := getCredentials(controller)
	if err != nil {
		// if id not found error
		if _, ok := err.(*clients.TenantNotFoundError); ok {
			logrus.Info("ID not found, registering with controller")
			err = registerWithProxy(controller, confNotValidErr)
			if err != nil {
				// tenant already exists, possible race condition in container group
				if _, ok = err.(*clients.ConflictError); ok {
					logrus.Warn("Possible race condition occurred during register")
					return nil
				}
				// unrecoverable error occurred registering with controller
				logrus.WithError(err).Error("Could not register with Controller")
				return err
			}

			// register succeeded
			return nil
		}
		// unrecoverable error occurred getting credentials from controller
		logrus.WithError(err).Error("Could not retrieve credentials")
		return err
	}

	if conf.ForceUpdate {
		// TODO
	}

	// if sidecar already has valid config do not need to set anything
	if confNotValidErr != nil {
		logrus.Info("Updating credentials with those from controller")
		conf.Kafka.APIKey = creds.Kafka.APIKey
		conf.Kafka.Brokers = creds.Kafka.Brokers
		conf.Kafka.Password = creds.Kafka.Password
		conf.Kafka.RestURL = creds.Kafka.RestURL
		conf.Kafka.SASL = creds.Kafka.SASL
		conf.Kafka.Username = creds.Kafka.User

		conf.Registry.Token = creds.Registry.Token
		conf.Registry.URL = creds.Registry.URL
	}
	return nil
}

func isRetryable(err error) bool {

	if _, ok := err.(*clients.ConnectionError); ok {
		return true
	}

	if _, ok := err.(*clients.NetworkError); ok {
		return true
	}

	if _, ok := err.(*clients.ServiceUnavailable); ok {
		return true
	}

	return false
}
