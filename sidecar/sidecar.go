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

	"encoding/json"
	"net/http"
	"time"

	controllerclient "github.com/amalgam8/amalgam8/controller/client"
	"github.com/amalgam8/amalgam8/controller/rules"
	"github.com/amalgam8/amalgam8/pkg/version"
	registryclient "github.com/amalgam8/amalgam8/registry/client"
	"github.com/amalgam8/amalgam8/sidecar/api"
	"github.com/amalgam8/amalgam8/sidecar/config"
	"github.com/amalgam8/amalgam8/sidecar/proxy"
	"github.com/amalgam8/amalgam8/sidecar/proxy/monitor"
	"github.com/amalgam8/amalgam8/sidecar/proxy/nginx"
	"github.com/amalgam8/amalgam8/sidecar/register"
	"github.com/amalgam8/amalgam8/sidecar/register/healthcheck"
	"github.com/amalgam8/amalgam8/sidecar/supervisor"
	"github.com/ant0ine/go-json-rest/rest"
)

// Main is the entrypoint for the sidecar when running as an executable
func Main() {
	logrus.ErrorKey = "error"
	logrus.SetLevel(logrus.DebugLevel) // Initial logging until we parse the user provided log level argument
	logrus.SetOutput(os.Stderr)

	app := cli.NewApp()

	app.Name = "sidecar"
	app.Usage = "Amalgam8 Sidecar"
	app.Version = version.Build.Version
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
	var registrationAgent *register.RegistrationAgent

	if conf.Debug != "" {
		cliCommand(conf.Debug)
		return nil
	}

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

	appSupervisor := supervisor.NewAppSupervisor(&conf)

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

		registrationAgent, err = register.NewRegistrationAgent(register.RegistrationConfig{
			Client:          registryClient,
			ServiceInstance: serviceInstance,
		})
		if err != nil {
			logrus.WithError(err).Error("Could not create registry agent")
			return err
		}

		hcAgents, err := healthcheck.BuildAgents(conf.HealthChecks)
		if err != nil {
			logrus.WithError(err).Error("Could not build health checks")
			return err
		}

		// Control the registration agent via the health checker if any health checks were provided. If no
		// health checks are provided, just start the registration agent.
		if len(hcAgents) > 0 {
			checker := register.NewHealthChecker(registrationAgent, hcAgents)

			// Delay slightly to give time for the application to start
			time.AfterFunc(1*time.Second, checker.Start) // TODO: make this delay configurable or implement a better solution.
		} else {
			registrationAgent.Start()
		}
	}

	appSupervisor.DoAppSupervision(registrationAgent)

	return nil
}

func startProxy(conf *config.Config) error {
	var err error

	nginxClient := nginx.NewClient("http://localhost:5813")
	nginxManager := nginx.NewManager(
		nginx.Config{
			Service: nginx.NewService(fmt.Sprintf("%v:%v", conf.Service.Name, strings.Join(conf.Service.Tags, ","))),
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

	debugger := api.NewDebugAPI(nginxProxy)

	a := rest.NewApi()
	a.Use(
		&rest.TimerMiddleware{},
		&rest.RecorderMiddleware{},
		&rest.RecoverMiddleware{
			EnableResponseStackTrace: false,
		},
		&rest.ContentTypeCheckerMiddleware{},
		//&middleware.RequestIDMiddleware{},
		//&middleware.LoggingMiddleware{},
		//middleware.NewRequireHTTPS(middleware.CheckRequest{
		//	IsSecure: middleware.IsUsingSecureConnection,
		//	Disabled: !conf.RequireHTTPS,
		//}),
	)

	routes := debugger.Routes()
	router, err := rest.MakeRouter(
		routes...,
	)
	if err != nil {
		logrus.WithError(err).Error("Could not start API server")
		return err
	}
	a.SetApp(router)

	go func() {
		http.ListenAndServe(fmt.Sprintf(":%v", 6116), a.MakeHandler())
	}()

	return nil
}

// Instance TODO
type Instance struct {
	Tags string `json:"tags"`
	Type string `json:"type"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

// NGINXState nginx cached lua state
type NGINXState struct {
	Routes    map[string][]rules.Rule `json:"routes"`
	Instances map[string][]Instance   `json:"instances"`
	Actions   map[string][]rules.Rule `json:"actions"`
}

func cliCommand(command string) {

	switch command {
	case "show-state":
		httpClient := http.Client{
			Timeout: time.Second * 10,
		}
		req, err := http.NewRequest("GET", "http://localhost:6116/state", nil)
		if err != nil {
			fmt.Println("Error occurred building the request:", err.Error())
			return
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			fmt.Println("Error occurred sending the request:", err.Error())
			return
		}

		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error occurred reading response body:", err.Error())
			return
		}

		sidecarstate := struct {
			Instances []registryclient.ServiceInstance `json:"instances"`
			Rules     []rules.Rule                     `json:"rules"`
		}{}

		err = json.Unmarshal(respBytes, &sidecarstate)
		if err != nil {
			fmt.Println("Error occurred loading JSON response:", err.Error())
			return
		}

		req, err = http.NewRequest("GET", "http://localhost:5813/a8-admin", nil)
		if err != nil {
			fmt.Println("Error occurred building the request:", err.Error())
			return
		}

		resp, err = httpClient.Do(req)
		if err != nil {
			fmt.Println("Error occurred sending the request:", err.Error())
			return
		}

		respBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error occurred reading response body:", err.Error())
			return
		}

		nginxstate := NGINXState{}

		err = json.Unmarshal(respBytes, &nginxstate)
		if err != nil {
			fmt.Println("Error occurred loading JSON response:", err.Error())
			return
		}

		sidecarBytes, _ := json.MarshalIndent(&sidecarstate, "", "   ")
		nginxBytes, _ := json.MarshalIndent(&nginxstate, "", "   ")
		fmt.Println("\n**************\nSidecar cached state:\n**************")
		fmt.Println(string(sidecarBytes))
		fmt.Println("\n**************\nNginx cached state:\n**************")
		fmt.Println(string(nginxBytes))

	default:
		fmt.Println("Unrecognized command: ", command)
	}
}
