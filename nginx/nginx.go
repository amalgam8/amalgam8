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
	"io"
	"strings"
	"text/template"

	"time"

	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/controller/resources"
)

// Generator produces NGINX configurations for tenants
type Generator interface {
	Generate(w io.Writer, id string, lastUpdate *time.Time) error
	TemplateConfig(catalog resources.ServiceCatalog, conf resources.ProxyConfig) resources.ConfigTemplate
}

type generator struct {
	template *template.Template
	db       database.Tenant
}

// Config options for the NGINX generator
type Config struct {
	Path     string
	Database database.Tenant
}

// NewGenerator creates a new NGINX generator using the given Golang template file
func NewGenerator(conf Config) (Generator, error) {
	t, err := template.ParseFiles(conf.Path)
	if err != nil {
		return nil, err
	}

	g := &generator{
		template: t,
		db:       conf.Database,
	}

	return g, nil
}

// Generate a NGINX config for a tenant using its catalog and proxy configuration.
func (g *generator) Generate(w io.Writer, id string, lastUpdate *time.Time) error {
	// Get inputs
	entry, err := g.db.Read(id)
	if err != nil {
		return err
	}

	if lastUpdate != nil && entry.ServiceCatalog.LastUpdate.After(*lastUpdate) {
		return nil
	}

	// Generate the struct for the template
	templateConf := g.TemplateConfig(entry.ServiceCatalog, entry.ProxyConfig)

	// Generate the NGINX configuration
	if err := g.template.Execute(w, &templateConf); err != nil {
		return err
	}

	return nil
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
func (g *generator) TemplateConfig(catalog resources.ServiceCatalog, conf resources.ProxyConfig) resources.ConfigTemplate {
	rules := map[string][]resources.Rule{}
	for _, rule := range conf.Filters.Rules {
		rules[rule.Destination] = append(rules[rule.Destination], rule)
	}

	unversionedVersionFilter := resources.Version{
		Default:   "UNVERSIONED",
		Selectors: "nil",
	}
	versionFilters := map[string]resources.Version{}
	for _, version := range conf.Filters.Versions {
		if version.Default == "" {
			version.Default = unversionedVersionFilter.Default
		}
		if version.Selectors == "" {
			version.Selectors = unversionedVersionFilter.Selectors
		}
		versionFilters[version.Service] = version
	}

	proxies := make([]resources.ServiceTemplate, 0, len(catalog.Services))
	for _, service := range catalog.Services {

		if _, ok := versionFilters[service.Name]; !ok {
			versionFilters[service.Name] = unversionedVersionFilter
		}

		upstreams := map[string][]string{}
		for _, endpoint := range service.Endpoints {
			if endpoint.Type == "http" { // We only support HTTP, not HTTPS or other protocols

				version := endpoint.Metadata.Version
				upstreamName := service.Name
				if version != "" {
					upstreamName += "_" + version
				} else {
					upstreamName += "_" + unversionedVersionFilter.Default
				}

				versionUpstreams := upstreams[upstreamName]
				if versionUpstreams == nil {
					versionUpstreams = []string{endpoint.Value}
				} else {
					versionUpstreams = append(versionUpstreams, endpoint.Value)
				}
				upstreams[upstreamName] = versionUpstreams
			}
		}

		// Only generate a proxy configuration if we have endpoints
		if len(upstreams) > 0 {
			versions := []resources.VersionedUpstreams{}
			for k, v := range upstreams {
				versions = append(versions, resources.VersionedUpstreams{
					UpstreamName: k,
					Upstreams:    v,
				})
			}
			proxies = append(proxies, resources.ServiceTemplate{
				ServiceName:      service.Name,
				Versions:         versions,
				VersionDefault:   versionFilters[service.Name].Default,
				VersionSelectors: versionFilters[service.Name].Selectors,
				Rules:            rules[service.Name],
			})
		}
	}

	// Create the struct expected by the template
	templateConf := resources.ConfigTemplate{
		Port:                 conf.Port,
		ReqTrackingHeader:    conf.ReqTrackingHeader,
		LogReqTrackingHeader: strings.Replace(strings.ToLower(conf.ReqTrackingHeader), "-", "_", -1),
		Proxies:              proxies,
	}

	return templateConf
}
