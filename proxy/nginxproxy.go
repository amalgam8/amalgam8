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
	"github.com/amalgam8/controller/rules"
	"github.com/amalgam8/registry/client"
	"github.com/amalgam8/sidecar/proxy/monitor"
	"github.com/amalgam8/sidecar/proxy/nginx"
)

// NGINXProxy updates NGINX to reflect changes in the A8 controller and A8 registry
type NGINXProxy interface {
	monitor.ControllerListener
	monitor.RegistryListener
}

type nginxProxy struct {
	instances []client.ServiceInstance
	rules     []rules.Rule
	nginx     nginx.Manager
	mutex     sync.Mutex
}

// NewNGINXProxy instantiates a new instance
func NewNGINXProxy(nginxClient nginx.Manager) NGINXProxy {
	return &nginxProxy{
		rules:     []rules.Rule{},
		instances: []client.ServiceInstance{},
		nginx:     nginxClient,
	}
}

// CatalogChange updates NGINX on a change in the catalog
func (n *nginxProxy) CatalogChange(instances []client.ServiceInstance) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.instances = instances
	return n.updateNGINX()
}

// RuleChange updates NGINX on a change in the proxy configuration
func (n *nginxProxy) RuleChange(rules []rules.Rule) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.rules = rules
	return n.updateNGINX()
}

func (n *nginxProxy) updateNGINX() error {
	logrus.Debug("Updating NGINX")
	return n.nginx.Update(n.instances, n.rules)
}
