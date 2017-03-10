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
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	amalgam8registry "github.com/amalgam8/amalgam8/pkg/adapters/discovery/amalgam8"
	"github.com/amalgam8/amalgam8/pkg/adapters/discovery/eureka"
	kubediscovery "github.com/amalgam8/amalgam8/pkg/adapters/discovery/kubernetes"
	amalgam8controller "github.com/amalgam8/amalgam8/pkg/adapters/rules/amalgam8"
	kuberules "github.com/amalgam8/amalgam8/pkg/adapters/rules/kubernetes"
	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/pkg/auth"
	kubepkg "github.com/amalgam8/amalgam8/pkg/kubernetes"
	"github.com/amalgam8/amalgam8/pkg/version"
	"github.com/amalgam8/amalgam8/sidecar/config"
	"github.com/amalgam8/amalgam8/sidecar/dns"
	"github.com/amalgam8/amalgam8/sidecar/identity"
	"github.com/amalgam8/amalgam8/sidecar/proxy"
	"github.com/amalgam8/amalgam8/sidecar/proxy/monitor"
	"github.com/amalgam8/amalgam8/sidecar/register"
	"github.com/amalgam8/amalgam8/sidecar/register/healthcheck"
	"github.com/amalgam8/amalgam8/sidecar/supervisor"
	"github.com/urfave/cli"
	"k8s.io/client-go/kubernetes"
)

// Main is the entry point for the sidecar.
func Main() {
	logrus.ErrorKey = "error"
	logrus.SetLevel(logrus.DebugLevel) // Initial logging level until we parse the user provided log level argument
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
	// When PID=1, launch new sidecar process and assume init responsibilities.
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

// Run the sidecar with the given configuration.
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

	var kubeClient kubernetes.Interface
	if conf.DiscoveryAdapter == config.KubernetesAdapter ||
		conf.RulesAdapter == config.KubernetesAdapter {
		kubeClient, err = kubepkg.NewClient(kubepkg.Config{
			URL:   conf.Kubernetes.URL,
			Token: conf.Kubernetes.Token,
		})

		if err != nil {
			logrus.WithError(err).Error("Could not create Kubernetes client")
			return err
		}
	}

	identity, err := identity.New(&conf, kubeClient)
	if err != nil {
		logrus.WithError(err).Error("Could not create identity provider")
		return err
	}

	var discovery api.ServiceDiscovery
	if conf.DNS || conf.Proxy {
		discovery, err = buildServiceDiscovery(&conf, kubeClient)
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
			logrus.WithError(err).Error("Could not start DNS server")
			return err
		}
		go server.ListenAndServe()
	}

	if conf.Proxy {
		err := startProxy(&conf, identity, discovery)
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

		registrationAgent, err := register.NewRegistrationAgent(register.RegistrationConfig{
			Registry: registry,
			Identity: identity,
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

func buildServiceRegistry(conf *config.Config) (api.ServiceRegistry, error) {
	switch strings.ToLower(conf.DiscoveryAdapter) {
	case config.Amalgam8Adapter:
		regConf := amalgam8registry.RegistryConfig{
			URL:       conf.A8Registry.URL,
			AuthToken: conf.A8Registry.Token,
		}
		return amalgam8registry.NewRegistryAdapter(regConf)
	case "":
		return nil, fmt.Errorf("no service discovery type specified")
	default:
		return nil, fmt.Errorf("registration using '%s' is not supported", conf.DiscoveryAdapter)
	}
}

func buildServiceDiscovery(conf *config.Config, kubeClient kubernetes.Interface) (api.ServiceDiscovery, error) {
	switch strings.ToLower(conf.DiscoveryAdapter) {
	case config.Amalgam8Adapter:
		regConf := amalgam8registry.RegistryConfig{
			URL:       conf.A8Registry.URL,
			AuthToken: conf.A8Registry.Token,
		}
		return amalgam8registry.NewCachedDiscoveryAdapter(regConf, conf.A8Registry.Poll)
	case config.KubernetesAdapter:
		kubConf := kubediscovery.Config{
			Namespace: auth.NamespaceFrom(conf.Kubernetes.Namespace),
			Client:    kubeClient,
		}
		return kubediscovery.New(kubConf)
	case config.EurekaAdapter:
		eurConf := eureka.Config{
			URLs: conf.Eureka.URLs,
		}
		return eureka.New(eurConf)
	case "":
		return nil, fmt.Errorf("no service discovery type specified")
	default:
		return nil, fmt.Errorf("discovery using '%s' is not supported", conf.DiscoveryAdapter)
	}
}

func buildServiceRules(conf *config.Config) (api.RulesService, error) {
	switch strings.ToLower(conf.RulesAdapter) {
	case config.Amalgam8Adapter:
		controllerConf := amalgam8controller.ControllerConfig{
			URL:       conf.A8Controller.URL,
			AuthToken: conf.A8Controller.Token,
		}
		return amalgam8controller.NewCachedRulesAdapter(controllerConf, conf.A8Controller.Poll)
	case config.KubernetesAdapter:
		kubConf := kuberules.Config{
			URL:       conf.Kubernetes.URL,
			Token:     conf.Kubernetes.Token,
			Namespace: auth.NamespaceFrom(conf.Kubernetes.Namespace),
		}
		return kuberules.New(kubConf)
	case "":
		return nil, fmt.Errorf("no service rules type specified")
	default:
		return nil, fmt.Errorf("rules using '%s' is not supported", conf.RulesAdapter)
	}
}

func buildProxyAdapter(conf *config.Config, identity identity.Provider, discovery monitor.DiscoveryMonitor,
	rules monitor.RulesMonitor, discoveryClient api.ServiceDiscovery, rulesClient api.RulesService) (proxy.Adapter, error) {

	switch conf.ProxyAdapter {
	case config.EnvoyAdapter:
		return proxy.NewEnvoyAdapter(conf, discovery, identity, rules, discoveryClient, rulesClient)
	default:
		return nil, fmt.Errorf("Unsupported proxy adapter: %v", conf.ProxyAdapter)

	}
}

func startProxy(conf *config.Config, identity identity.Provider, discovery api.ServiceDiscovery) error {
	rules, err := buildServiceRules(conf)
	if err != nil {
		logrus.WithError(err).Error("Could not create service rules client")
		return err
	}

	rulesMonitor := monitor.NewRulesMonitor(monitor.RulesConfig{
		Rules:        rules,
		PollInterval: conf.A8Controller.Poll,
	})

	discoveryMonitor := monitor.NewDiscoveryMonitor(monitor.DiscoveryConfig{
		Discovery: discovery,
	})

	proxyAdapter, err := buildProxyAdapter(conf, identity, discoveryMonitor, rulesMonitor, discovery, rules)
	if err != nil {
		logrus.WithError(err).Error("Could not build proxy adapter")
		return err
	}

	if err := proxyAdapter.Start(); err != nil {
		logrus.WithError(err).Error("Could not start proxy adapter")
		return err
	}

	go func() {
		if err := discoveryMonitor.Start(); err != nil {
			logrus.WithError(err).Error("Discovery monitor failed")
		}
	}()

	go func() {
		if err := rulesMonitor.Start(); err != nil {
			logrus.WithError(err).Error("Rules monitor failed")
		}
	}()

	return nil
}
