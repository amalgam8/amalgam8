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

package sidecar

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	controllerclient "github.com/amalgam8/amalgam8/controller/client"
	registryclient "github.com/amalgam8/amalgam8/registry/client"
	"github.com/amalgam8/amalgam8/sidecar/config"
	"github.com/amalgam8/amalgam8/sidecar/proxy"
	"github.com/amalgam8/amalgam8/sidecar/proxy/monitor"
	"github.com/amalgam8/amalgam8/sidecar/proxy/nginx"
	"github.com/amalgam8/amalgam8/sidecar/register"
	"github.com/amalgam8/amalgam8/sidecar/supervisor"
)

// Main is the entrypoint for the sidecar when running as an executable
func Main() {
	logrus.ErrorKey = "error"
	logrus.SetLevel(logrus.DebugLevel) // Initial logging until we parse the user provided log level argument
	logrus.SetOutput(os.Stderr)

	app := cli.NewApp()

	app.Name = "sidecar"
	app.Usage = "Amalgam8 Sidecar"
	app.Version = "0.2.0"
	app.Flags = config.Flags
	app.Action = sidecarCommand

	err := app.Run(os.Args)
	if err != nil {
		logrus.WithError(err).Error("Failure launching sidecar")
	}
}

func sidecarCommand(context *cli.Context) error {
	conf, err := config.New(context)
	if err != nil {
		return err
	}
	return Run(*conf)
}

// Run the sidecar with the given configuration
func Run(conf config.Config) error {
	var err error

	if err = conf.Validate(); err != nil {
		logrus.WithError(err).Error("Validation of config failed")
		return err
	}

	logrusLevel, err := logrus.ParseLevel(conf.LogLevel)
	if err != nil {
		logrus.WithError(err).Errorf("Failure parsing requested log level (%v)", conf.LogLevel)
		logrusLevel = logrus.DebugLevel
	}
	logrus.SetLevel(logrusLevel)

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

		address := fmt.Sprintf("%v:%v", conf.Endpoint.Host, conf.Endpoint.Port)
		serviceInstance := &registryclient.ServiceInstance{
			ServiceName: conf.Service.Name,
			Tags:        conf.Service.Tags,
			Endpoint: registryclient.ServiceEndpoint{
				Type:  conf.Endpoint.Type,
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
		supervisor.DoAppSupervision(conf.App)
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
