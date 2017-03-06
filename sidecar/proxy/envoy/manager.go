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

package envoy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"sort"

	"strings"

	"io/ioutil"

	"strconv"

	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/sidecar/config"
	"github.com/amalgam8/amalgam8/sidecar/identity"
	"github.com/amalgam8/amalgam8/sidecar/util"
)

// Envoy config related files
const (
	envoyConfigFile     = "envoy.json"
	runtimePath         = "runtime/routing"
	runtimeVersionsPath = "routing_versions"
	adminLog            = "envoy_admin.log"
	accessLog           = "a8_access.log"

	configDirPerm  = 0775
	configFilePerm = 0664

	DefaultDiscoveryPort    = 6500
	DefaultAdminPort        = 8001
	DefaultHTTPListenerPort = 6379
	DefaultWorkingDir       = "/etc/envoy/"
	DefaultLoggingDir       = "/var/log/"
	DefaultEnvoyBinary      = "envoy"
)

const envoyLogFormat = `` +
	`{` +
	`"status":"%%RESPONSE_CODE%%", ` +
	`"start_time":"%%START_TIME%%", ` +
	`"request_time":%%DURATION%%, ` +
	`"upstream_response_time":"%%RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)%%", ` +
	`"src":"%v", ` +
	`"dst":"%%UPSTREAM_CLUSTER%%", ` +
	`"%v":"%v", ` +
	`"message":"%%REQ(:METHOD)%% %%REQ(X-ENVOY-ORIGINAL-PATH?:PATH)%% %%RESPONSE_CODE%%", ` +
	`"module":"ENVOY", ` +
	`"method":"%%REQ(:METHOD)%%", ` +
	`"path":"%%REQ(X-ENVOY-ORIGINAL-PATH?:PATH)%%", ` +
	`"protocol":"%%PROTOCOL%%", ` +
	`"response_flags":"%%RESPONSE_FLAGS%%", ` +
	`"x_forwarded":"%%REQ(X-FORWARDED-FOR)%%", ` +
	`"user_agent":"%%REQ(USER-AGENT)%%", ` +
	`"request_id":"%%REQ(X-REQUEST-ID)%%", ` +
	`"auth":"%%REQ(:AUTHORITY)%%", ` +
	`"upstream_host":"%%UPSTREAM_HOST%%"` +
	"}\n"

// Manager for updating envoy proxy configuration.
type Manager interface {
	Update(instances []api.ServiceInstance, rules []api.Rule) error
}

// NewManager creates new instance
func NewManager(identity identity.Provider, conf *config.Config) (Manager, error) {
	m := &manager{
		identity:     identity,
		sdsPort:      conf.ProxyConfig.DiscoveryPort,
		adminPort:    conf.ProxyConfig.AdminPort,
		listenerPort: conf.ProxyConfig.HTTPListenerPort,
		workingDir:   conf.ProxyConfig.WorkingDir,
		loggingDir:   conf.ProxyConfig.LoggingDir,
		service: NewService(ServiceConfig{
			DrainTimeSeconds:          3,
			ParentShutdownTimeSeconds: 5,
			EnvoyConfig:               conf.ProxyConfig.WorkingDir + envoyConfigFile,
			EnvoyBinary:               conf.ProxyConfig.ProxyBinary,
		}),
		GrpcHttp1Bridge: conf.ProxyConfig.GrpcHttp1Bridge,
	}

	if err := buildFS(m.workingDir); err != nil {
		return nil, err
	}

	return m, nil
}

type manager struct {
	identity        identity.Provider
	service         Service
	sdsPort         int
	adminPort       int
	listenerPort    int //Single listener port. TODO: Change to array, with port type Http|TCP
	workingDir      string
	loggingDir      string
	GrpcHttp1Bridge bool
}

func (m *manager) Update(instances []api.ServiceInstance, rules []api.Rule) error {

	// TODO if only fault or route values have changed, can update the filesystem and do not need to reload
	//if err := updateFS(m.workingDir, instances, rules); err != nil {
	//	return Config{}, err
	//}

	conf, err := m.generateConfig(rules, instances)
	if err != nil {
		return err
	}

	if err := writeConfigFile(conf, m.workingDir+envoyConfigFile); err != nil {
		return err
	}

	if err := m.service.Reload(); err != nil {
		return err
	}

	return nil
}

func writeConfigFile(conf Config, confPath string) error {
	file, err := os.Create(confPath)
	if err != nil {
		return err
	}

	if err := writeConfig(file, conf); err != nil {
		file.Close()
		return err
	}

	return file.Close()
}

func writeConfig(w io.Writer, conf Config) error {
	out, err := json.MarshalIndent(&conf, "", "  ")
	if err != nil {
		return err
	}

	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return err
}

func (m *manager) generateConfig(rules []api.Rule, instances []api.ServiceInstance) (Config, error) {
	inst, err := m.identity.GetIdentity()
	if err != nil {
		return Config{}, err
	}

	filters := buildFaults(rules, inst.ServiceName, inst.Tags)

	if m.GrpcHttp1Bridge {
		filters = append(filters, buildGrpcHttp1BridgeFilter())
	}

	traceKey := "gremlin_recipe_id"
	traceVal := "-"
	for _, rule := range rules {
		for _, action := range rule.Actions {
			if action.Action == "trace" {
				if rule.Match != nil && rule.Match.Source != nil && rule.Match.Source.Name == inst.ServiceName {
					traceKey = action.LogKey
					traceVal = action.LogValue
				}
			}
		}
	}

	format := fmt.Sprintf(envoyLogFormat, buildSourceName(inst.ServiceName, inst.Tags), traceKey, traceVal)

	return Config{
		RootRuntime: RootRuntime{
			SymlinkRoot:  m.workingDir + runtimePath,
			Subdirectory: "traffic_shift",
		},
		Listeners: []Listener{
			{
				Port: m.listenerPort,
				Filters: []NetworkFilter{
					{
						Type: "read",
						Name: "http_connection_manager",
						Config: NetworkFilterConfig{
							CodecType:         "auto",
							StatPrefix:        "ingress_http",
							UserAgent:         true,
							GenerateRequestID: true,
							RDS: &RDS{
								Cluster:         "rds",
								RouteConfigName: "amalgam8",
								RefreshDelayMS:  1000,
							},
							Filters: filters,
							AccessLog: []AccessLog{
								{
									Path:   m.loggingDir + accessLog,
									Format: format,
								},
							},
						},
					},
				},
			},
		},
		Admin: Admin{
			AccessLogPath: m.loggingDir + adminLog,
			Port:          m.adminPort,
		},
		ClusterManager: ClusterManager{
			Clusters: []Cluster{
				{
					Name:             "rds",
					Type:             "strict_dns",
					ConnectTimeoutMs: 1000,
					LbType:           "round_robin",
					Hosts: []Host{
						{
							URL: fmt.Sprintf("tcp://127.0.0.1:%v", m.sdsPort),
						},
					},
					MaxRequestsPerConnection: 1,
				},
			},
			SDS: SDS{
				Cluster: Cluster{
					Name:             "sds",
					Type:             "strict_dns",
					ConnectTimeoutMs: 1000,
					LbType:           "round_robin",
					Hosts: []Host{
						{
							URL: fmt.Sprintf("tcp://127.0.0.1:%v", m.sdsPort),
						},
					},
					MaxRequestsPerConnection: 1,
				},
				RefreshDelayMs: 1000,
			},
			CDS: CDS{
				Cluster: Cluster{
					Name:             "cds",
					Type:             "strict_dns",
					ConnectTimeoutMs: 1000,
					LbType:           "round_robin",
					Hosts: []Host{
						{
							URL: fmt.Sprintf("tcp://127.0.0.1:%v", m.sdsPort),
						},
					},
					MaxRequestsPerConnection: 1,
				},
				RefreshDelayMs: 1000,
			},
		},
	}, nil
}

const (
	serviceDelimiter = ':'
	tagDelimiter     = ','
)

// BuildServiceKey builds a service key given a service name and tags in the
// form "serviceName:tag1=value1,tag2=value2,tag3=value3" where ':' is the
// service delimiter and ',' is the tag delimiter. We assume that the service
// name and the tags do not contain either delimiter.
func BuildServiceKey(service string, tags []string) string {
	sort.Strings(tags)

	buf := bytes.NewBufferString(service)

	lt := len(tags)
	if lt > 0 {
		buf.WriteByte(serviceDelimiter)
		for i := 0; i < lt-1; i++ {
			buf.WriteString(tags[i])
			buf.WriteByte(tagDelimiter)
		}
		buf.WriteString(tags[lt-1])
	}

	return buf.String()
}

// ParseServiceKey parses service key into service name and tags. We do not
// check for the correctness of the service key.
func ParseServiceKey(s string) (string, []string) {
	res := strings.FieldsFunc(s, func(r rune) bool {
		if r == tagDelimiter || r == serviceDelimiter {
			return true
		}
		return false
	})

	if len(res) <= 0 {
		return "", []string{}
	}

	return res[0], res[1:]
}

// BuildClusters builds clusters from instances applying rule backend info where necessary
func BuildClusters(instances []*api.ServiceInstance, rules []api.Rule, tlsConfig *SSLContext) []Cluster {
	SanitizeRules(rules)
	rules = AddDefaultRouteRules(rules, instances)

	staticClusters := make(map[string]struct{})
	instancesByClusterMap := make(map[string][]*api.ServiceInstance)
	for _, inst := range instances {
		clusterName := BuildServiceKey(inst.ServiceName, inst.Tags)
		if _, _, err := util.SplitHostPort(inst.Endpoint); err != nil {
			staticClusters[clusterName] = struct{}{}
			staticClusters[inst.ServiceName] = struct{}{}
		}
		instancesByClusterMap[clusterName] = append(instancesByClusterMap[clusterName], inst)
		instancesByClusterMap[inst.ServiceName] = append(instancesByClusterMap[inst.ServiceName], inst)
	}

	clusterMap := make(map[string]*api.Backend)
	for _, rule := range rules {
		if rule.Route != nil {
			for _, backend := range rule.Route.Backends {
				key := BuildServiceKey(backend.Name, backend.Tags)
				// TODO if two backends map to the same key, it will overwrite
				//  and will lose resilience field options in this case
				clusterMap[key] = &backend
			}
		}
	}

	clusters := make([]Cluster, 0, len(clusterMap))

	for clusterName, backend := range clusterMap {
		cluster := buildCluster(clusterName, backend, tlsConfig)

		// if cluster contains strict_dns hosts, append them
		if _, exists := staticClusters[clusterName]; exists {
			cluster.Type = "strict_dns"
			cluster.ServiceName = ""
			hosts := make([]Host, 0, len(instancesByClusterMap[clusterName]))
			for _, inst := range instancesByClusterMap[clusterName] {
				portSet := false
				url := fmt.Sprintf("tcp://%v", inst.Endpoint.Value)
				// Envoy complains if no port provided, append :80 by default
				for _, p := range strings.Split(inst.Endpoint.Value, ":") {
					if _, err := strconv.Atoi(p); err == nil {
						portSet = true
						break
					}
				}
				if !portSet {
					url = fmt.Sprintf("tcp://%v:80", inst.Endpoint.Value)
				}
				host := Host{
					URL: url,
				}
				hosts = append(hosts, host)
			}
			cluster.Hosts = hosts

		}
		clusters = append(clusters, cluster)
	}

	sort.Sort(ClustersByName(clusters))

	return clusters
}

// BuildWeightKey builds filesystem key for Route Runtime weight keys
func BuildWeightKey(service string, tags []string) string {
	return fmt.Sprintf("%v.%v", service, BuildServiceKey("_", tags))
}

func buildCluster(clusterName string, backend *api.Backend, tlsConfig *SSLContext) Cluster {
	cluster := Cluster{
		Name:             clusterName,
		ServiceName:      clusterName,
		Type:             "sds",
		LbType:           "round_robin",
		ConnectTimeoutMs: 1000,
		CircuitBreakers:  &CircuitBreakers{},
		OutlierDetection: &OutlierDetection{
			MaxEjectionPercent: 100,
		},
		SSLContext: tlsConfig,
	}

	if backend != nil && backend.LbType != "" {
		// Set default value of LbType to be "round_robin"
		cluster.LbType = backend.LbType
	}

	if backend.Resilience != nil {
		// Cluster level settings
		if backend.Resilience.MaxRequestsPerConnection > 0 {
			cluster.MaxRequestsPerConnection = backend.Resilience.MaxRequestsPerConnection
		}

		// Envoy Circuit breaker config options
		if backend.Resilience.MaxConnections > 0 {
			cluster.CircuitBreakers.Default.MaxConnections = backend.Resilience.MaxConnections
		}
		if backend.Resilience.MaxRequests > 0 {
			cluster.CircuitBreakers.Default.MaxRequests = backend.Resilience.MaxRequests
		}
		if backend.Resilience.MaxPendingRequest > 0 {
			cluster.CircuitBreakers.Default.MaxPendingRequest = backend.Resilience.MaxPendingRequest
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

	return cluster
}

// BuildRoutes builds routes based on rules.  Assumes at least one route for each service
func BuildRoutes(ruleList []api.Rule, instances []*api.ServiceInstance) []Route {
	SanitizeRules(ruleList)
	rules := AddDefaultRouteRules(ruleList, instances)

	staticClusters := make(map[string]struct{})
	for _, inst := range instances {
		clusterName := BuildServiceKey(inst.ServiceName, inst.Tags)
		if _, _, err := util.SplitHostPort(inst.Endpoint); err != nil {
			staticClusters[clusterName] = struct{}{}
			staticClusters[inst.ServiceName] = struct{}{}
		}
	}

	routes := []Route{}
	for _, rule := range rules {
		if rule.Route != nil {
			route := buildRoute(&rule)
			// Enable auto_host_rewrite if cluster contains strict_dns hosts
			if _, ok := staticClusters[route.Cluster]; ok {
				route.AutoHostRewrite = true
			}
			for _, w := range route.WeightedClusters.Clusters {
				if _, ok := staticClusters[w.Name]; ok {
					route.AutoHostRewrite = true
				}
			}
			routes = append(routes, route)
		}
	}

	return routes
}

func buildRoute(rule *api.Rule) Route {
	var headers []Header
	if rule.Match != nil {
		headers = make([]Header, 0, len(rule.Match.Headers))
		for k, v := range rule.Match.Headers {
			headers = append(
				headers,
				Header{
					Name:  k,
					Value: v,
					Regex: true,
				},
			)
		}
	}

	var path, prefix, prefixRewrite string
	if rule.Route.URI != nil {
		path = rule.Route.URI.Path
		prefix = rule.Route.URI.Prefix
		prefixRewrite = rule.Route.URI.PrefixRewrite
	} else {
		prefix = fmt.Sprintf("/%v/", rule.Destination)
		prefixRewrite = "/"
	}

	route := Route{
		Path:          path,
		Prefix:        prefix,
		PrefixRewrite: prefixRewrite,
		Headers:       headers,
		RetryPolicy: RetryPolicy{
			Policy: "5xx,connect-failure,refused-stream",
		},
		WeightedClusters: WeightedClusters{
			Clusters: make([]WeightedCluster, 0),
		},
	}

	if rule.Route.HTTPReqTimeout > 0 {
		// convert from float sec to in ms
		route.TimeoutMS = int(rule.Route.HTTPReqTimeout * 1000)
	}
	if rule.Route.HTTPReqRetries > 0 {
		route.RetryPolicy.NumRetries = rule.Route.HTTPReqRetries
	}

	if rule.Route.HTTPReqTimeout > 0 {
		// convert from float sec to in ms
		route.TimeoutMS = int(rule.Route.HTTPReqTimeout * 1000)
	}
	if rule.Route.HTTPReqRetries > 0 {
		route.RetryPolicy.NumRetries = rule.Route.HTTPReqRetries
	}

	for _, backend := range rule.Route.Backends {
		route.WeightedClusters.Clusters = append(route.WeightedClusters.Clusters, WeightedCluster{
			Name:   BuildServiceKey(backend.Name, backend.Tags),
			Weight: int(backend.Weight * 100),
		})
	}

	return route
}

// ByPriority implement sort
type ByPriority []api.Rule

// Len length
func (s ByPriority) Len() int {
	return len(s)
}

// Swap elements
func (s ByPriority) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less compare
func (s ByPriority) Less(i, j int) bool {
	return s[i].Priority < s[j].Priority
}

// SanitizeRules performs sorts on rule backends and rules.  Also calculates remaining weights
func SanitizeRules(ruleList []api.Rule) {
	for i := range ruleList {
		rule := &ruleList[i]
		if rule.Route != nil {
			var sum float64

			undefined := 0
			for j := range rule.Route.Backends {
				backend := &ruleList[i].Route.Backends[j]
				if backend.Name == "" {
					backend.Name = rule.Destination
				}

				if backend.Weight == 0.0 {
					undefined++
				} else {
					sum += backend.Weight
				}

				sort.Strings(backend.Tags)
			}

			if undefined > 0 {
				w := (1.0 - sum) / float64(undefined)
				for j := range rule.Route.Backends {
					backend := &ruleList[i].Route.Backends[j]
					if backend.Weight == 0 {
						backend.Weight = w
					}
				}
			}
		}
	}

	sort.Sort(sort.Reverse(ByPriority(ruleList))) // Descending order
}

// AddDefaultRouteRules adds a route rule for a service that currently does not have one
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
			Destination: service,
		})
	}

	return append(ruleList, defaults...)
}

// FIXME: doesn't check for name conflicts
// TODO: could be improved by using the full possible set of filenames.
func randFilename(prefix string) string {
	data := make([]byte, 16)
	for i := range data {
		data[i] = '0' + byte(rand.Intn(10))
	}

	return fmt.Sprintf("%s%s", prefix, data)
}

func buildFS(workingDir string) error {
	if err := os.MkdirAll(filepath.Dir(workingDir+runtimePath), configDirPerm); err != nil { // FIXME: hack
		return err
	}

	if err := os.MkdirAll(workingDir+runtimeVersionsPath, configDirPerm); err != nil {
		return err
	}

	return nil
}

func updateFS(workingDir string, instances []api.ServiceInstance, ruleList []api.Rule) error {
	SanitizeRules(ruleList)

	ptrInstances := make([]*api.ServiceInstance, 0, len(instances))
	for _, inst := range instances {
		ptrInstances = append(ptrInstances, &inst)
	}
	rules := AddDefaultRouteRules(ruleList, ptrInstances)

	type weightSpec struct {
		Service string
		Cluster string
		Weight  int
	}

	var weights []weightSpec
	for _, rule := range rules {
		if rule.Route != nil {
			w := 0
			for _, backend := range rule.Route.Backends {
				w += int(100 * backend.Weight)
				weight := weightSpec{
					Service: backend.Name,
					Cluster: BuildServiceKey("_", backend.Tags),
					Weight:  w,
				}
				weights = append(weights, weight)
			}
		}
	}

	dirName, err := ioutil.TempDir(workingDir+runtimeVersionsPath, "")
	if err != nil {
		return err
	}

	success := false
	defer func() {
		if !success {
			os.RemoveAll(dirName)
		}
	}()

	for _, weight := range weights {
		if err := os.MkdirAll(filepath.Join(dirName, "/traffic_shift/", weight.Service), configDirPerm); err != nil {
			return err
		} // FIXME: filemode?

		filename := filepath.Join(dirName, "/traffic_shift/", weight.Service, weight.Cluster)
		data := []byte(fmt.Sprintf("%v", weight.Weight))
		if err := ioutil.WriteFile(filename, data, configFilePerm); err != nil {
			return err
		}
	}

	oldRuntime, err := os.Readlink(workingDir + runtimePath)
	if err != nil && !os.IsNotExist(err) { // Ignore error from symlink not existing.
		return err
	}

	tmpName := randFilename(workingDir + "/")

	if err := os.Symlink(dirName, tmpName); err != nil {
		return err
	}

	// Atomically replace the runtime symlink
	if err := os.Rename(tmpName, workingDir+runtimePath); err != nil {
		return err
	}

	success = true

	// Clean up the old config FS if necessary
	// TODO: make this safer
	if oldRuntime != "" {
		oldRuntimeDir := filepath.Dir(oldRuntime)
		if filepath.Clean(oldRuntimeDir) == filepath.Clean(workingDir+runtimeVersionsPath) {
			toDelete := filepath.Join(workingDir+runtimeVersionsPath, filepath.Base(oldRuntime))
			if err := os.RemoveAll(toDelete); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildFaults(ctlrRules []api.Rule, serviceName string, tags []string) []Filter {
	var filters []Filter

	tagMap := make(map[string]struct{})
	for _, tag := range tags {
		tagMap[tag] = struct{}{}
	}

	for _, rule := range ctlrRules {
		var headers []Header
		if rule.Match != nil {
			headers = make([]Header, 0, len(rule.Match.Headers))
			for key, val := range rule.Match.Headers {
				headers = append(headers, Header{
					Name:  key,
					Value: val,
					Regex: true,
				})
			}

			if rule.Match.Source != nil && rule.Match.Source.Name == serviceName {
				isSubset := true
				for _, tag := range rule.Match.Source.Tags {
					if _, exists := tagMap[tag]; !exists {
						isSubset = false
						break
					}
				}

				if isSubset {
					for _, action := range rule.Actions {
						switch action.Action {
						case "delay":
							filter := Filter{
								Type: "decoder",
								Name: "fault",
								Config: &FilterFaultConfig{
									Delay: &DelayFilter{
										Type:     "fixed",
										Percent:  int(action.Probability * 100),
										Duration: int(action.Duration * 1000),
									},
									Headers: headers,
								},
							}
							filters = append(filters, filter)
						case "abort":
							filter := Filter{
								Type: "decoder",
								Name: "fault",
								Config: &FilterFaultConfig{
									Abort: &AbortFilter{
										Percent:    int(action.Probability * 100),
										HTTPStatus: action.ReturnCode,
									},
									Headers: headers,
								},
							}
							filters = append(filters, filter)
						}
					}
				}
			}
		}
	}

	filters = append(filters, Filter{
		Type:   "decoder",
		Name:   "router",
		Config: FilterRouterConfig{},
	})

	return filters
}

func buildGrpcHttp1BridgeFilter() Filter {
	// Construct http1_grpc_bridge filter
	return Filter{
		Type:   "both",
		Name:   "grpc_http1_bridge",
		Config: &GrpcHttp1BridgeFilter{},
	}
}

func buildSourceName(service string, tags []string) string {
	return fmt.Sprintf("%v:%v", service, strings.Join(tags, ","))

}
