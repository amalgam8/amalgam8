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

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/sidecar/config"
	"github.com/amalgam8/amalgam8/sidecar/proxy/monitor"
	"github.com/amalgam8/amalgam8/sidecar/proxy/nginx"
)

// NGINXAdapter manages an NGINX based proxy.
type NGINXAdapter struct {
	client           nginx.Client
	manager          nginx.Manager
	service          nginx.Service
	discoveryMonitor monitor.DiscoveryMonitor
	rulesMonitor     monitor.RulesMonitor

	instances []api.ServiceInstance
	rules     []api.Rule
	mutex     sync.Mutex
}

// NewNGINXAdapter creates a new adapter instance.
func NewNGINXAdapter(conf *config.Config, discoveryMonitor monitor.DiscoveryMonitor, rulesMonitor monitor.RulesMonitor) (*NGINXAdapter, error) {
	client := nginx.NewClient("http://localhost:5813") // FIXME: hardcoded
	service := nginx.NewService(conf.Service.Name, conf.Service.Tags)
	manager := nginx.NewManager(nginx.ManagerConfig{
		Client: client,
	})
	if conf.ProxyConfig.TLS {
		if err := nginx.GenerateConfig(conf.ProxyConfig); err != nil {
			logrus.WithError(err).Error("Could not generate NGINX SSL config")
			return nil, err
		}
	}
	return &NGINXAdapter{
		client:           client,
		manager:          manager,
		service:          service,
		discoveryMonitor: discoveryMonitor,
		rulesMonitor:     rulesMonitor,

		instances: []api.ServiceInstance{},
		rules:     []api.Rule{},
	}, nil
}

// Start NGINX proxy.
func (a *NGINXAdapter) Start() error {
	if err := a.service.Start(); err != nil {
		logrus.WithError(err).Error("NGINX service failed to start")
		return err
	}

	a.discoveryMonitor.AddListener(a)
	a.rulesMonitor.AddListener(a)

	return nil
}

// Stop NGINX proxy.
func (a *NGINXAdapter) Stop() error {
	a.discoveryMonitor.RemoveListener(a)
	a.rulesMonitor.RemoveListener(a)

	return a.service.Stop()
}

// CatalogChange updates on a change in the catalog.
func (a *NGINXAdapter) CatalogChange(instances []api.ServiceInstance) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.instances = instances
	return a.manager.Update(a.instances, a.rules)
}

// RuleChange updates NGINX on a change in the proxy configuration.
func (a *NGINXAdapter) RuleChange(rules []api.Rule) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.rules = rules
	return a.manager.Update(a.instances, a.rules)
}
