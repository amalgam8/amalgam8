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

package router

import (
	"strconv"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/resources"
	"github.com/amalgam8/sidecar/router/clients"
	"github.com/amalgam8/sidecar/router/monitor"
	"github.com/amalgam8/sidecar/router/nginx"
)

type NGINXRouter interface {
	monitor.ControllerListener
	monitor.RegistryListener
}

type nginxRouter struct {
	catalog     resources.ServiceCatalog
	proxyConfig resources.ProxyConfig
	nginx       nginx.Nginx
	mutex       sync.Mutex
}

func NewNGINXRouter(nginxClient nginx.Nginx) NGINXRouter {
	return &nginxRouter{
		proxyConfig: resources.ProxyConfig{
			LoadBalance: "round_robin",
			Filters: resources.Filters{
				Versions: []resources.Version{},
				Rules:    []resources.Rule{},
			},
		},
		nginx: nginxClient,
	}
}

func (l *nginxRouter) CatalogChange(catalog resources.ServiceCatalog) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.catalog = catalog
	return l.updateNGINX()
}

func (l *nginxRouter) RuleChange(proxyConfig resources.ProxyConfig) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.proxyConfig = proxyConfig
	return l.updateNGINX()
}

func (l *nginxRouter) updateNGINX() error {
	nginxJSON := l.buildConfig()
	return l.nginx.Update(nginxJSON)
}

func (l *nginxRouter) buildConfig() clients.NGINXJson {
	conf := clients.NGINXJson{
		Upstreams: make(map[string]clients.NGINXUpstream, 0),
		Services:  make(map[string]clients.NGINXService, 0),
	}
	faults := []clients.NGINXFault{}
	for _, rule := range l.proxyConfig.Filters.Rules {
		fault := clients.NGINXFault{
			Delay:            rule.Delay,
			DelayProbability: rule.DelayProbability,
			AbortProbability: rule.AbortProbability,
			AbortCode:        rule.ReturnCode,
			Source:           rule.Source,
			Destination:      rule.Destination,
			Header:           rule.Header,
			Pattern:          rule.Pattern,
		}
		faults = append(faults, fault)
	}
	conf.Faults = faults

	types := map[string]string{}
	for _, service := range l.catalog.Services {
		upstreams := map[string][]clients.NGINXEndpoint{}
		for _, endpoint := range service.Endpoints {
			version := endpoint.Metadata.Version
			upstreamName := service.Name
			if version != "" {
				upstreamName += ":" + version
			} else {
				upstreamName += ":" + "UNVERSIONED"
			}

			types[service.Name] = endpoint.Type

			vals := strings.Split(endpoint.Value, ":")
			if len(vals) != 2 {
				logrus.WithFields(logrus.Fields{
					"endpoint": endpoint,
					"values":   vals,
				}).Error("could not parse host and port from service endpoint")
			}
			host := vals[0]
			port, err := strconv.Atoi(vals[1])
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"err":  err,
					"port": vals[1],
				}).Error("port not a valid int")
			}

			versionUpstreams := upstreams[upstreamName]
			nginxEndpoint := clients.NGINXEndpoint{
				Host: host,
				Port: port,
			}
			if versionUpstreams == nil {
				versionUpstreams = []clients.NGINXEndpoint{nginxEndpoint}
			} else {
				versionUpstreams = append(versionUpstreams, nginxEndpoint)
			}
			upstreams[upstreamName] = versionUpstreams
		}

		for k, v := range upstreams {
			conf.Upstreams[k] = clients.NGINXUpstream{
				Upstreams: v,
			}
		}
	}

	versions := map[string]resources.Version{}
	for _, version := range l.proxyConfig.Filters.Versions {
		versions[version.Service] = version
	}

	for k, v := range types {
		if version, ok := versions[k]; ok {
			conf.Services[k] = clients.NGINXService{
				Default:   version.Default,
				Selectors: version.Selectors,
				Type:      v,
			}
		} else {
			conf.Services[k] = clients.NGINXService{
				Type: v,
			}
		}
	}

	return conf
}
