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

package proxy

import (
	"sync"

	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/sidecar/config"
	"github.com/amalgam8/amalgam8/sidecar/discovery"
	"github.com/amalgam8/amalgam8/sidecar/identity"
	"github.com/amalgam8/amalgam8/sidecar/proxy/envoy"
	"github.com/amalgam8/amalgam8/sidecar/proxy/monitor"
)

// EnvoyAdapter manages an Envoy based proxy.
type EnvoyAdapter struct {
	discoveryMonitor monitor.DiscoveryMonitor
	rulesMonitor     monitor.RulesMonitor

	manager envoy.Manager

	instances []api.ServiceInstance
	rules     []api.Rule
	mutex     sync.Mutex
}

// NewEnvoyAdapter creates a new adapter instance.
func NewEnvoyAdapter(conf *config.Config, discoveryMonitor monitor.DiscoveryMonitor,
	identity identity.Provider, rulesMonitor monitor.RulesMonitor,
	discoveryClient api.ServiceDiscovery, rulesClient api.RulesService) (*EnvoyAdapter, error) {

	if conf.ProxyConfig.HTTPListenerPort == 0 {
		conf.ProxyConfig.HTTPListenerPort = envoy.DefaultHTTPListenerPort
	}

	if conf.ProxyConfig.DiscoveryPort == 0 {
		conf.ProxyConfig.DiscoveryPort = envoy.DefaultDiscoveryPort
	}

	if conf.ProxyConfig.AdminPort == 0 {
		conf.ProxyConfig.AdminPort = envoy.DefaultAdminPort
	}

	if conf.ProxyConfig.WorkingDir == "" {
		conf.ProxyConfig.WorkingDir = envoy.DefaultWorkingDir
	}

	if conf.ProxyConfig.LoggingDir == "" {
		conf.ProxyConfig.LoggingDir = envoy.DefaultLoggingDir
	}

	if conf.ProxyConfig.ProxyBinary == "" {
		conf.ProxyConfig.ProxyBinary = envoy.DefaultEnvoyBinary
	}

	serverConfig := &discovery.Config{
		HTTPAddressSpec: fmt.Sprintf(":%d", conf.ProxyConfig.DiscoveryPort),
		Discovery:       discoveryClient,
		Rules:           rulesClient,
	}
	server, err := discovery.NewDiscoveryServer(serverConfig)
	if err != nil {
		logrus.WithError(err).Error("Discovery server failed to start")
		return nil, err
	}
	err = server.Start()
	if err != nil {
		logrus.WithError(err).Error("Discovery server failed to start")
		return nil, err
	}

	manager := envoy.NewManager(identity, conf)

	return &EnvoyAdapter{
		manager:          manager,
		discoveryMonitor: discoveryMonitor,
		rulesMonitor:     rulesMonitor,

		instances: []api.ServiceInstance{},
		rules:     []api.Rule{},
	}, nil
}

// Start Envoy proxy.
func (a *EnvoyAdapter) Start() error {
	var err error
	func(err error) {
		a.mutex.Lock()
		defer a.mutex.Unlock()
		err = a.manager.Update(a.instances, a.rules)
	}(err)

	if err != nil {
		logrus.WithError(err).Error("Envoy service failed to start")
		return err
	}

	a.discoveryMonitor.AddListener(a)
	a.rulesMonitor.AddListener(a)

	return nil
}

// Stop Envoy proxy.
func (a *EnvoyAdapter) Stop() error {
	a.discoveryMonitor.RemoveListener(a)
	a.rulesMonitor.RemoveListener(a)

	// TODO stop envoy service
	return nil
}

// CatalogChange updates on a change in the catalog.
func (a *EnvoyAdapter) CatalogChange(instances []api.ServiceInstance) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.instances = instances
	// Envoy uses SDS api for dynamically updating services - do NOT need to notify envoy of any changes
	return nil
}

// RuleChange updates Envoy on a change in the proxy configuration.
func (a *EnvoyAdapter) RuleChange(rules []api.Rule) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.rules = rules
	return a.manager.Update(a.instances, a.rules)
}
