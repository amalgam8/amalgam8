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
	logrus.ErrorKey = "error"
	logrus.SetLevel(logrus.DebugLevel) // Initial logging until we parse the user provided log level argument
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

	if err = conf.Validate(); err != nil {
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

	rc := clients.NewController(conf)
	nc := clients.NewNGINXClient("http://localhost:5813")

	n, err := nginx.NewNginx(
		nginx.Conf{
			Service:     nginx.NewService(),
			NGINXClient: nc,
		},
	)
	if err != nil {
		logrus.WithError(err).Error("Failed to initialize NGINX object")
		return err
	}

	registryClient, err := client.New(client.Config{
		URL:       conf.Registry.URL,
		AuthToken: conf.Registry.Token,
	})
	if err != nil {
		logrus.WithError(err).Error("Could not create registry client")
		return err
	}

	listener := checker.NewListener(n)

	poller := checker.NewPoller(conf, rc, listener)
	go func() {
		if err = poller.Start(); err != nil {
			logrus.WithError(err).Error("Controller poll failed")
		}
	}()

	checker := checker.New(checker.Config{
		Listener:       listener,
		RegistryClient: registryClient,
		Conf:           conf,
	})
	go func() {
		if err = checker.Start(); err != nil {
			logrus.WithError(err).Error("Registry poll failed")
		}
	}()

	return nil
}

// TODO move this to controller client?
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
