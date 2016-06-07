package nginx

import (
	"github.com/amalgam8/controller/checker"
	"github.com/amalgam8/controller/proxyconfig"
	"github.com/amalgam8/controller/resources"
	"io"
	"strings"
	"text/template"
)

// Generator produces NGINX configurations for tenants
type Generator interface {
	Generate(w io.Writer, id string) error
}

type generator struct {
	template     *template.Template
	checker      checker.Checker
	proxyManager proxyconfig.Manager
}

// Config options for the NGINX generator
type Config struct {
	Path         string
	Catalog      checker.Checker
	ProxyManager proxyconfig.Manager
}

// NewGenerator creates a new NGINX generator using the given Golang template file
func NewGenerator(conf Config) (Generator, error) {
	t, err := template.ParseFiles(conf.Path)
	if err != nil {
		return nil, err
	}

	g := &generator{
		template:     t,
		checker:      conf.Catalog,
		proxyManager: conf.ProxyManager,
	}

	return g, nil
}

// Generate a NGINX config for a tenant using its catalog and proxy configuration.
func (g *generator) Generate(w io.Writer, id string) error {
	// Get inputs
	catalog, err := g.serviceCatalog(id)
	if err != nil {
		return err
	}

	conf, err := g.proxyConfig(id)
	if err != nil {
		return err
	}

	// Generate the struct for the template
	templateConf := g.templateConfig(catalog, conf)

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
func (g *generator) templateConfig(catalog resources.ServiceCatalog, conf resources.ProxyConfig) configTemplate {
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

	proxies := make([]serviceTemplate, 0, len(catalog.Services))
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
			versions := []versionedUpstreams{}
			for k, v := range upstreams {
				versions = append(versions, versionedUpstreams{k, v})
			}
			proxies = append(proxies, serviceTemplate{
				ServiceName:      service.Name,
				Versions:         versions,
				VersionDefault:   versionFilters[service.Name].Default,
				VersionSelectors: versionFilters[service.Name].Selectors,
				Rules:            rules[service.Name],
			})
		}
	}

	// Create the struct expected by the template
	templateConf := configTemplate{
		Port:                 conf.Port,
		ReqTrackingHeader:    conf.ReqTrackingHeader,
		LogReqTrackingHeader: strings.Replace(strings.ToLower(conf.ReqTrackingHeader), "-", "_", -1),
		Proxies:              proxies,
	}

	return templateConf
}

// serviceCatalog
func (g *generator) serviceCatalog(id string) (resources.ServiceCatalog, error) {
	return g.checker.Get(id)
}

// proxyConfig
func (g *generator) proxyConfig(id string) (resources.ProxyConfig, error) {
	return g.proxyManager.Get(id)
}

// configTemplate is used by the template file to generate the NGINX config
type configTemplate struct {
	Port                 int
	ReqTrackingHeader    string
	LogReqTrackingHeader string
	Proxies              []serviceTemplate
}

type versionedUpstreams struct {
	UpstreamName string
	Upstreams    []string
}

// serviceTemplate is used by the template file to generate service configurations in the NGINX config
type serviceTemplate struct {
	ServiceName      string
	Versions         []versionedUpstreams
	VersionDefault   string
	VersionSelectors string
	Rules            []resources.Rule
}
