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

package discovery

import (
	"net/http"
	"sort"

	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/sidecar/proxy/envoy"
	"github.com/amalgam8/amalgam8/sidecar/util"
	"github.com/ant0ine/go-json-rest/rest"
)

const (
	apiVer                       = "/v1"
	routeParamServiceClusterName = "csname"
	routeParamServiceNodeName    = "snname"

	routeParamServiceName = "sname"
	registrationPath      = apiVer + "/registration"
	registrationTemplate  = registrationPath + "/#" + routeParamServiceName

	clusterPath     = apiVer + "/clusters"
	clusterTemplate = clusterPath + "/#" + routeParamServiceClusterName + "/#" + routeParamServiceNodeName

	routeParamRouteConfigName = "route_conf_name"
	routesPath                = apiVer + "/routes"
	routesTemplate            = routesPath + "/#" + routeParamRouteConfigName + "/#" + routeParamServiceClusterName + "/#" + routeParamServiceNodeName
	///v1/routes/(string: route_config_name)/(string: service_cluster)/(string: service_node)
)

// VirtualHosts is the array of route info returned by GET routes
type VirtualHosts struct {
	VirtualHosts []envoy.VirtualHost `json:"virtual_hosts"`
}

// Clusters is the array of clusters returned by GET clusters
type Clusters struct {
	Clusters []envoy.Cluster `json:"clusters"`
}

// Hosts is the array of hosts returned by the GET registration
type Hosts struct {
	Hosts []Host `json:"hosts"`
}

// Host is the endpoint and tag data for a service instance
type Host struct {
	IPAddr string            `json:"ip_address"`
	Port   uint16            `json:"port"`
	Tags   map[string]string `json:"tags"`
}

// ByIPPort ip and port based sorting for hosts.
type ByIPPort []Host

// Len length.
func (s ByIPPort) Len() int {
	return len(s)
}

// Swap elements.
func (s ByIPPort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less compare elements.
func (s ByIPPort) Less(i, j int) bool {
	if s[i].IPAddr == s[j].IPAddr {
		return s[i].Port < s[j].Port
	}
	return s[i].IPAddr < s[j].IPAddr
}

// Discovery handles discovery API calls
type Discovery struct {
	discovery api.ServiceDiscovery
	rules     api.RulesService
	tlsConfig *envoy.SSLContext
}

// NewDiscovery creates struct
func NewDiscovery(discovery api.ServiceDiscovery, rules api.RulesService, tlsConfig *envoy.SSLContext) *Discovery {
	return &Discovery{
		discovery: discovery,
		rules:     rules,
		tlsConfig: tlsConfig,
	}
}

// Routes for discovery API
func (d *Discovery) Routes(middlewares ...rest.Middleware) []*rest.Route {
	routes := []*rest.Route{
		rest.Get(registrationTemplate, d.getRegistration),
		rest.Get(clusterTemplate, d.getClusters),
		rest.Get(routesTemplate, d.getRoutes),
	}

	for _, route := range routes {
		route.Func = rest.WrapMiddlewares(middlewares, route.Func)
	}
	return routes
}

// getRegistration
func (d *Discovery) getRegistration(w rest.ResponseWriter, req *rest.Request) {
	sname := req.PathParam(routeParamServiceName)

	hosts, err := d.getHosts(sname)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	resp := Hosts{
		Hosts: hosts,
	}

	w.WriteHeader(http.StatusOK)
	w.WriteJson(&resp)
}

func (d *Discovery) getHosts(name string) ([]Host, error) {
	service, tags := envoy.ParseServiceKey(name)

	instances, err := d.discovery.ListServiceInstances(service)
	if err != nil {
		logrus.WithError(err).Warnf("Failed to get the list of service instances")
		return []Host{}, err
	}

	filteredInstances := filterInstances(instances, tags)

	return translate(filteredInstances), nil
}

// get service name, tags
// get instances by service name
// filter instances (only instances that have ALL the tags)
// convert instances to SD hosts
func translate(instances []*api.ServiceInstance) []Host {
	hosts := []Host{}
	tags := make(map[string]string)

	for _, instance := range instances {
		var host Host
		ip, port, err := util.SplitHostPort(instance.Endpoint)
		if err != nil {
			logrus.WithError(err).Warnf("unable to resolve ip address for instance '%s'", instance.ID)
			continue
		}
		host = Host{IPAddr: ip.String(), Port: port, Tags: tags}
		hosts = append(hosts, host)
	}

	// Ensure that order is preserved between calls.
	sort.Sort(ByIPPort(hosts))

	return hosts
}

func filterInstances(instances []*api.ServiceInstance, tags []string) []*api.ServiceInstance {
	tagMap := make(map[string]struct{})
	for i := range tags {
		tagMap[tags[i]] = struct{}{}
	}

	filtered := make([]*api.ServiceInstance, 0, len(instances))
	for _, instance := range instances {
		count := 0
		for i := range instance.Tags {
			_, exists := tagMap[instance.Tags[i]]
			if exists {
				count++
			}
		}

		if count == len(tags) {
			filtered = append(filtered, instance)
		}
	}

	return filtered
}

// getRegistration
func (d *Discovery) getClusters(w rest.ResponseWriter, req *rest.Request) {
	//clusterName := req.PathParam(routeParamServiceClusterName)
	//
	//nodeName := req.PathParam(routeParamServiceNodeName)

	instances, err := d.discovery.ListInstances()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	ruleSet, err := d.rules.ListRules(&api.RuleFilter{})
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	resp := Clusters{
		Clusters: buildClusters(instances, ruleSet.Rules, d.tlsConfig),
	}

	w.WriteHeader(http.StatusOK)
	w.WriteJson(&resp)
}

func buildClusters(instances []*api.ServiceInstance, rules []api.Rule, tlsConfig *envoy.SSLContext) []envoy.Cluster {

	clusterNames := make(map[string]struct{})
	for _, instance := range instances {
		clusterName := envoy.BuildServiceKey(instance.ServiceName, instance.Tags)
		clusterNames[clusterName] = struct{}{}
		// Need a default cluster for every service
		clusterNames[instance.ServiceName] = struct{}{}
	}

	// TODO ?
	envoy.SanitizeRules(rules)
	rules = AddDefaultRouteRules(rules, instances)

	backends := make(map[string]*api.Backend)
	for _, rule := range rules {
		if rule.Route != nil {
			for _, backend := range rule.Route.Backends {
				key := envoy.BuildServiceKey(backend.Name, backend.Tags)
				// TODO if two backends map to the same key, it will overwrite
				//  and will lose resilience field options in this case
				backends[key] = &backend
			}
		}
	}

	clusters := make([]envoy.Cluster, 0, len(clusterNames))
	for name := range clusterNames {
		cluster := envoy.Cluster{
			Name:             name,
			ServiceName:      name,
			Type:             "sds",
			LbType:           "round_robin", //TODO this needs to be configurable
			ConnectTimeoutMs: 1000,
			CircuitBreakers:  &envoy.CircuitBreakers{},
			OutlierDetection: &envoy.OutlierDetection{
				MaxEjectionPercent: 100,
			},
			SSLContext: tlsConfig,
		}

		if backend, ok := backends[name]; ok {
			if backend.LbType != "" {
				cluster.LbType = backend.LbType
			}

			if backend.Resilience != nil {
				// Cluster level settings
				if backend.Resilience.MaxRequestsPerConnection > 0 {
					cluster.MaxRequestsPerConnection = backend.Resilience.MaxRequestsPerConnection
				}

				// Envoy Circuit breaker config options
				if backend.Resilience.MaxConnections > 0 {
					cluster.CircuitBreakers.MaxConnections = backend.Resilience.MaxConnections
				}
				if backend.Resilience.MaxRequests > 0 {
					cluster.CircuitBreakers.MaxRequests = backend.Resilience.MaxRequests
				}
				if backend.Resilience.MaxPendingRequest > 0 {
					cluster.CircuitBreakers.MaxPendingRequest = backend.Resilience.MaxPendingRequest
				}

				// Envoy outlier detection settings that complete circuit breaker
				if backend.Resilience.SleepWindow > 0 {
					cluster.OutlierDetection.BaseEjectionTimeMS = int(backend.Resilience.SleepWindow * 1000)
				}
				if backend.Resilience.ConsecutiveErrors > 0 {
					cluster.OutlierDetection.ConsecutiveError = backend.Resilience.ConsecutiveErrors
				}
				if backend.Resilience.DetectionInterval > 0 {
					cluster.OutlierDetection.IntervalMS = int(backend.Resilience.DetectionInterval * 1000)
				}
			}
		}

		clusters = append(clusters, cluster)

	}

	sort.Sort(envoy.ClustersByName(clusters))

	return clusters
}

func (d *Discovery) getRoutes(w rest.ResponseWriter, req *rest.Request) {

	instances, err := d.discovery.ListInstances()
	if err != nil {
		logrus.WithError(err).Error("Unable to retrieve instances")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	ruleSet, err := d.rules.ListRules(&api.RuleFilter{})
	if err != nil {
		logrus.WithError(err).Error("Unable to retrieve rules")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	envoy.SanitizeRules(ruleSet.Rules)
	rules := AddDefaultRouteRules(ruleSet.Rules, instances)

	routes := buildRoutes(rules)

	respJSON := VirtualHosts{
		VirtualHosts: []envoy.VirtualHost{
			{
				Name:    "backend",
				Domains: []string{"*"},
				Routes:  routes,
			},
		},
	}

	w.WriteHeader(http.StatusOK)
	w.WriteJson(&respJSON)
}

func buildRoutes(ruleList []api.Rule) []envoy.Route {
	routes := []envoy.Route{}
	for _, rule := range ruleList {
		if rule.Route != nil {
			var headers []envoy.Header
			if rule.Match != nil {
				headers = make([]envoy.Header, 0, len(rule.Match.Headers))
				for k, v := range rule.Match.Headers {
					headers = append(
						headers,
						envoy.Header{
							Name:  k,
							Value: v,
							Regex: true,
						},
					)
				}
			}

			for _, backend := range rule.Route.Backends {
				var path, prefix, prefixRewrite string
				if backend.URI != nil {
					path = backend.URI.Path
					prefix = backend.URI.Prefix
					prefixRewrite = backend.URI.PrefixRewrite
				} else {
					prefix = fmt.Sprintf("/%v/", backend.Name)
					prefixRewrite = "/"
				}

				clusterName := envoy.BuildServiceKey(backend.Name, backend.Tags)

				runtime := &envoy.Runtime{
					Key:     envoy.BuildWeightKey(backend.Name, backend.Tags),
					Default: int(backend.Weight * 100),
				}

				route := envoy.Route{
					Runtime:       runtime,
					Path:          path,
					Prefix:        prefix,
					PrefixRewrite: prefixRewrite,
					Cluster:       clusterName,
					Headers:       headers,
					RetryPolicy: envoy.RetryPolicy{
						Policy: "5xx,connect-failure,refused-stream",
					},
				}

				if rule.Route.HTTPReqTimeout > 0 {
					// convert from float sec to in ms
					route.TimeoutMS = int(rule.Route.HTTPReqTimeout * 1000)
				}
				if rule.Route.HTTPReqRetries > 0 {
					route.RetryPolicy.NumRetries = rule.Route.HTTPReqRetries
				}

				routes = append(routes, route)
			}
		}
	}

	return routes
}

// AddDefaultRouteRules TODO same as the one in envoy manager but takes []*api.ServiceInstance instead of []api.ServiceInstance
func AddDefaultRouteRules(ruleList []api.Rule, instances []*api.ServiceInstance) []api.Rule {
	serviceMap := make(map[string]struct{})
	for _, instance := range instances {
		serviceMap[instance.ServiceName] = struct{}{}
	}

	for _, rule := range ruleList {
		if rule.Route != nil {
			for _, backend := range rule.Route.Backends {
				delete(serviceMap, backend.Name)
			}
		}
	}

	// Provide defaults for all services without any routing rules.
	defaults := make([]api.Rule, 0, len(serviceMap))
	for service := range serviceMap {
		defaults = append(defaults, api.Rule{
			Route: &api.Route{
				Backends: []api.Backend{
					{
						Name:   service,
						Weight: 1.0,
					},
				},
			},
		})
	}

	return append(ruleList, defaults...)
}
