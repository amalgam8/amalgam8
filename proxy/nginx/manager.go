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

package nginx

import (
	"strconv"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/resources"
	"github.com/amalgam8/sidecar/proxy/clients"
)

// Manager of updates to NGINX
type Manager interface {
	// Update NGINX with the provided configuration
	Update(catalog resources.ServiceCatalog, proxyConfig resources.ProxyConfig) error
}

type manager struct {
	service     Service
	serviceName string
	mutex       sync.Mutex
	client      clients.NGINX
}

// Config options
type Config struct {
	Service Service
	Client  clients.NGINX
}

// NewManager creates new a instance
func NewManager(conf Config) Manager {
	return &manager{
		service: conf.Service,
		client:  conf.Client,
	}
}

// Update NGINX
func (n *manager) Update(catalog resources.ServiceCatalog, proxyConfig resources.ProxyConfig) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	var err error

	// Ensure NGINX is running
	running, err := n.service.Running()
	if err != nil {
		logrus.WithError(err).Error("Could not get status of NGINX service")
		return err
	}

	if !running {
		// NGINX is not running; attempt to start NGINX
		logrus.Info("Starting NGINX")
		if err := n.service.Start(); err != nil {
			logrus.WithError(err).Error("Failed to start NGINX service")
			return err
		}
	}

	nginxConf := n.buildConfig(catalog, proxyConfig)

	if err = n.client.UpdateHTTPUpstreams(nginxConf); err != nil {
		logrus.WithError(err).Error("Failed to update HTTP upstreams with NGINX")
		return err
	}

	return nil
}

// buildConfig constructs the request used by the NGINX client
func (n *manager) buildConfig(catalog resources.ServiceCatalog, proxyConfig resources.ProxyConfig) clients.NGINXJson {
	conf := clients.NGINXJson{
		Upstreams: make(map[string]clients.NGINXUpstream, 0),
		Services:  make(map[string]clients.NGINXService, 0),
	}
	faults := []clients.NGINXFault{}
	for _, rule := range proxyConfig.Filters.Rules {
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
	for _, service := range catalog.Services {
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
	for _, version := range proxyConfig.Filters.Versions {
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
