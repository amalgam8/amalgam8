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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	controllerclient "github.com/amalgam8/amalgam8/controller/client"
	"github.com/amalgam8/amalgam8/controller/rules"
	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/pkg/version"
	"github.com/amalgam8/amalgam8/registry/adapters/eureka"
	"github.com/amalgam8/amalgam8/registry/adapters/kubernetes"
	registryapi "github.com/amalgam8/amalgam8/pkg/api"
	registryclient "github.com/amalgam8/amalgam8/registry/client"
	"github.com/amalgam8/amalgam8/sidecar/api"
	"github.com/amalgam8/amalgam8/sidecar/config"
	"github.com/amalgam8/amalgam8/sidecar/dns"
	"github.com/amalgam8/amalgam8/sidecar/proxy"
	"github.com/amalgam8/amalgam8/sidecar/proxy/monitor"
	"github.com/amalgam8/amalgam8/sidecar/proxy/nginx"
	"github.com/amalgam8/amalgam8/sidecar/register"
	"github.com/amalgam8/amalgam8/sidecar/register/healthcheck"
	"github.com/amalgam8/amalgam8/sidecar/supervisor"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/urfave/cli"
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
	// when PID=1, launch new sidecar process and assume init responsibilities
	if os.Getpid() == 1 {
		supervisor.Init()

		// Init should never return
		return nil
	}

	conf, err := config.New(context)
	if err != nil {
		return err
	}
	return Run(*conf)

}

// Run the sidecar with the given configuration
func Run(conf config.Config) error {
	var err error

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

	var discovery registryapi.ServiceDiscovery
	if conf.DNS || conf.Proxy {
		discovery, err = buildServiceDiscovery(&conf)
		if err != nil {
			logrus.WithError(err).Error("Could not create service discovery backend adapter")
			return err
		}
	}

	if conf.DNS {
		dnsConfig := dns.Config{
			Discovery: discovery,
			Port:      uint16(conf.Dnsconfig.Port),
			Domain:    conf.Dnsconfig.Domain,
		}
		server, err := dns.NewServer(dnsConfig)
		if err != nil {
			logrus.WithError(err).Error("Could not start dns server")
			return err
		}
		go server.ListenAndServe()
	}

	if conf.Proxy {
		err := startProxy(&conf, discovery)
		if err != nil {
			logrus.WithError(err).Error("Could not start proxy")
			return err
		}
	}

	var lifecycle register.Lifecycle
	if conf.Register {
		registry, err := buildServiceRegistry(&conf)
		if err != nil {
			return err
		}

		address := fmt.Sprintf("%v:%v", conf.Endpoint.Host, conf.Endpoint.Port)
		serviceInstance := &registryapi.ServiceInstance{
			ServiceName: conf.Service.Name,
			Tags:        conf.Service.Tags,
			Endpoint: registryapi.ServiceEndpoint{
				Type:  conf.Endpoint.Type,
				Value: address,
			},
			TTL: 60,
		}

		registrationAgent, err := register.NewRegistrationAgent(register.RegistrationConfig{
			Registry:        registry,
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
			lifecycle = register.NewHealthChecker(registrationAgent, hcAgents)
		} else {
			lifecycle = registrationAgent
		}

		// Delay slightly to give time for the application to start
		// TODO: make this delay configurable or implement a better solution.
		time.AfterFunc(1*time.Second, lifecycle.Start)
	}

	appSupervisor := supervisor.NewAppSupervisor(&conf, lifecycle)
	appSupervisor.DoAppSupervision()

	return nil
}

func buildServiceRegistry(conf *config.Config) (registryapi.ServiceRegistry, error) {
	switch strings.ToLower(conf.Registry.Backend) {
	case config.Amalgam8Backend:
		regConf := registryclient.Config{
			URL:       conf.Registry.Amalgam8.URL,
			AuthToken: conf.Registry.Amalgam8.Token,
		}
		return registryclient.New(regConf)
	case "":
		return nil, fmt.Errorf("no service registry backend specified")
	default:
		return nil, fmt.Errorf("registration using '%s' is not supported", conf.Registry.Backend)
	}
}

func buildServiceDiscovery(conf *config.Config) (registryapi.ServiceDiscovery, error) {
	switch strings.ToLower(conf.Registry.Backend) {
	case config.Amalgam8Backend:
		regConf := registryclient.CacheConfig{
			Config: registryclient.Config{
				URL:       conf.Registry.Amalgam8.URL,
				AuthToken: conf.Registry.Amalgam8.Token,
			},
			PollInterval: conf.Registry.Poll,
		}
		return registryclient.NewCache(regConf)
	case config.KubernetesBackend:
		kubConf := kubernetes.Config{
			URL:       conf.Registry.Kubernetes.URL,
			Token:     conf.Registry.Kubernetes.Token,
			Namespace: auth.NamespaceFrom(conf.Registry.Kubernetes.Namespace),
		}
		return kubernetes.New(kubConf)
	case config.EurekaBackend:
		eurConf := eureka.Config{
			URLs: conf.Registry.Eureka.URLs,
		}
		return eureka.New(eurConf)
	case "":
		return nil, fmt.Errorf("no service discovery backend specified")
	default:
		return nil, fmt.Errorf("discovery using '%s' is not supported", conf.Registry.Backend)
	}
}

func startProxy(conf *config.Config, discovery registryapi.ServiceDiscovery) error {
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

	controllerMonitor := monitor.NewControllerMonitor(monitor.ControllerConfig{
		Client: controllerClient,
		Listeners: []monitor.ControllerListener{
			nginxProxy,
		},
		PollInterval: conf.Controller.Poll,
	})

	registryMonitor := monitor.NewRegistryMonitor(monitor.RegistryConfig{
		Discovery: discovery,
		Listeners: []monitor.RegistryListener{
			nginxProxy,
		},
	})

	go func() {
		if err = controllerMonitor.Start(); err != nil {
			logrus.WithError(err).Error("Controller monitor failed")
		}
	}()
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
		&rest.RecoverMiddleware{EnableResponseStackTrace: false},
		&rest.ContentTypeCheckerMiddleware{},
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
	IP   string `json:"ip"`
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
			Instances []registryapi.ServiceInstance `json:"instances"`
			Rules     []rules.Rule                  `json:"rules"`
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
