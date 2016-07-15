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
	"strings"

	"time"

	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/controller/resources"
)

// Generator produces NGINX configurations for tenants
type Generator interface {
	Generate(id string, lastUpdate *time.Time) (*resources.NGINXJson, error)
	TemplateConfig(catalog resources.ServiceCatalog, conf resources.ProxyConfig) resources.NGINXJson
}

type generator struct {
	db database.Tenant
}

// Config options for the NGINX generator
type Config struct {
	Database database.Tenant
}

// NewGenerator creates a new NGINX generator using the given Golang template file
func NewGenerator(conf Config) (Generator, error) {

	g := &generator{
		db: conf.Database,
	}

	return g, nil
}

// Generate a NGINX config for a tenant using its catalog and proxy configuration.
func (g *generator) Generate(id string, lastUpdate *time.Time) (*resources.NGINXJson, error) {

	// Get inputs
	entry, err := g.db.Read(id)
	if err != nil {
		return nil, err
	}

	if lastUpdate != nil && entry.ServiceCatalog.LastUpdate.After(*lastUpdate) {
		return nil, nil
	}

	retval := g.TemplateConfig(entry.ServiceCatalog, entry.ProxyConfig)

	return &retval, nil
}

/*
It is possible for rules and Registry to become out of sync.

For instance, given an initial setup...

Rules:
Rules for Service A
Rules for Service B

SD:
Service A
Service B

NGINX output:
Service A with rules
Service B with rules

SD could miss a heartbeat to Service A and no longer register it, leading to...

Rules:
Rules for Service A
Rules for Service B

SD:
Service B

NGINX output:
Service B with rules

NGINX output is the intersection of rules and the Registry catalog.
Rules are independent of Services except (maybe) when they are initially created.
*/

// templateConfig generates the structure expected by the template file which is used to generate NGINX. It also filters
// out non-HTTP endpoints.
func (g *generator) TemplateConfig(catalog resources.ServiceCatalog, conf resources.ProxyConfig) resources.NGINXJson {

	retval := resources.NGINXJson{
		Upstreams: make(map[string]resources.NGINXUpstream, 0),
		Services:  make(map[string]resources.NGINXService, 0),
	}
	faults := []resources.NGINXFault{}
	for _, rule := range conf.Filters.Rules {
		fault := resources.NGINXFault{
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
	retval.Faults = faults

	types := map[string]string{}
	for _, service := range catalog.Services {
		upstreams := map[string][]resources.NGINXEndpoint{}
		for _, endpoint := range service.Endpoints {
			version := endpoint.Metadata.Version
			upstreamName := service.Name
			if version != "" {
				upstreamName += "_" + version
			} else {
				upstreamName += "_" + "UNVERSIONED"
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
			nginxEndpoint := resources.NGINXEndpoint{
				Host: host,
				Port: port,
			}
			if versionUpstreams == nil {
				versionUpstreams = []resources.NGINXEndpoint{nginxEndpoint}
			} else {
				versionUpstreams = append(versionUpstreams, nginxEndpoint)
			}
			upstreams[upstreamName] = versionUpstreams
		}

		for k, v := range upstreams {
			retval.Upstreams[k] = resources.NGINXUpstream{
				Upstreams: v,
			}
		}
	}

	versions := map[string]resources.Version{}
	for _, version := range conf.Filters.Versions {
		versions[version.Service] = version
	}

	for k, v := range types {
		if version, ok := versions[k]; ok {
			retval.Services[k] = resources.NGINXService{
				Default:   version.Default,
				Selectors: version.Selectors,
				Type:      v,
			}
		} else {
			retval.Services[k] = resources.NGINXService{
				Type: v,
			}
		}
	}

	return retval
}
