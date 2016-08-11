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

	"github.com/Sirupsen/logrus"
	controllerclient "github.com/amalgam8/controller/client"
	registryclient "github.com/amalgam8/registry/client"
	"github.com/amalgam8/sidecar/config"
	"github.com/amalgam8/sidecar/proxy"
	"github.com/amalgam8/sidecar/proxy/monitor"
	"github.com/amalgam8/sidecar/proxy/nginx"
	"github.com/amalgam8/sidecar/register"
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

		registryClient, err := registryclient.New(registryclient.Config{
			URL:       conf.Registry.URL,
			AuthToken: conf.Registry.Token,
		})
		if err != nil {
			logrus.WithError(err).Error("Could not create registry client")
			return err
		}

		address := fmt.Sprintf("%v:%v", conf.EndpointHost, conf.EndpointPort)
		serviceInstance := &registryclient.ServiceInstance{
			ServiceName: conf.ServiceName,
			Tags:        conf.ServiceTags,
			Endpoint: registryclient.ServiceEndpoint{
				Type:  conf.EndpointType,
				Value: address,
			},
			TTL: 60,
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

	nginxClient := nginx.NewClient("http://localhost:5813")
	nginxManager := nginx.NewManager(
		nginx.Config{
			Service: nginx.NewService(),
			Client:  nginxClient,
		},
	)
	nginxProxy := proxy.NewNGINXProxy(nginxManager)

	controllerClient, err := controllerclient.New(controllerclient.Config{
		URL:       conf.Controller.URL,
		AuthToken: conf.Controller.Token,
	})
	if err != nil {
		logrus.WithError(err).Error("Could not create controller client")
		return err
	}

	registryClient, err := registryclient.New(registryclient.Config{
		URL:       conf.Registry.URL,
		AuthToken: conf.Registry.Token,
	})
	if err != nil {
		logrus.WithError(err).Error("Could not create registry client")
		return err
	}

	controllerMonitor := monitor.NewController(monitor.ControllerConfig{
		Client: controllerClient,
		Listeners: []monitor.ControllerListener{
			nginxProxy,
		},
		PollInterval: conf.Controller.Poll,
	})
	go func() {
		if err = controllerMonitor.Start(); err != nil {
			logrus.WithError(err).Error("Controller monitor failed")
		}
	}()

	registryMonitor := monitor.NewRegistry(monitor.RegistryConfig{
		PollInterval: conf.Registry.Poll,
		Listeners: []monitor.RegistryListener{
			nginxProxy,
		},
		RegistryClient: registryClient,
	})
	go func() {
		if err = registryMonitor.Start(); err != nil {
			logrus.WithError(err).Error("Registry monitor failed")
		}
	}()

	return nil
}
