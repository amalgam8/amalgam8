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
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sort"

	"strings"

	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/sidecar/config"
	"github.com/amalgam8/amalgam8/sidecar/identity"
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
func NewManager(identity identity.Provider, conf *config.Config) Manager {
	return &manager{
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
	}
}

type manager struct {
	identity     identity.Provider
	service      Service
	sdsPort      int
	adminPort    int
	listenerPort int //Single listener port. TODO: Change to array, with port type Http|TCP
	workingDir   string
	loggingDir   string
}

func (m *manager) Update(instances []api.ServiceInstance, rules []api.Rule) error {
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

	sanitizeRules(rules)
	rules = addDefaultRouteRules(rules, instances)

	clusters := buildClusters(rules)
	routes := buildRoutes(rules)
	filters := buildFaults(rules, inst.ServiceName, inst.Tags)

	if err := buildFS(rules, m.workingDir); err != nil {
		return Config{}, err
	}

	traceKey := "gremlin_recipe_id"
	traceVal := "-"
	for _, rule := range rules {
		for _, action := range rule.Actions {
			if action.GetType() == "trace" {
				if rule.Match != nil && rule.Match.Source != nil && rule.Match.Source.Name == inst.ServiceName {
					trace := action.Internal().(api.TraceAction)
					traceKey = trace.LogKey
					traceVal = trace.LogValue
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
				Port: m.listenerPort, //TODO: needs to be generated based on m.listenerPort
				Filters: []NetworkFilter{
					{
						Type: "read",
						Name: "http_connection_manager",
						Config: NetworkFilterConfig{
							CodecType:         "auto",
							StatPrefix:        "ingress_http",
							UserAgent:         true,
							GenerateRequestID: true,
							RouteConfig: RouteConfig{
								VirtualHosts: []VirtualHost{
									{
										Name:    "backend",
										Domains: []string{"*"},
										Routes:  routes,
									},
								},
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
			Clusters: clusters,
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
		},
	}, nil
}

const (
	serviceDelimiter = ':'
	tagDelimiter     = ','
)

// buildServiceKey builds a service key given a service name and tags in the
// form "serviceName:tag1=value1,tag2=value2,tag3=value3" where ':' is the
// service delimiter and ',' is the tag delimiter. We assume that the service
// name and the tags do not contain either delimiter.
func buildServiceKey(service string, tags []string) string {
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

func buildClusters(rules []api.Rule) []Cluster {
	clusterMap := make(map[string]*api.Backend)
	for _, rule := range rules {
		if rule.Route != nil {
			for _, backend := range rule.Route.Backends {
				key := buildServiceKey(backend.Name, backend.Tags)
				clusterMap[key] = &backend
			}
		}
	}

	clusters := make([]Cluster, 0, len(clusterMap))
	for clusterName, backend := range clusterMap {

		cluster := Cluster{
			Name:             clusterName,
			ServiceName:      clusterName,
			Type:             "sds",
			LbType:           "round_robin",
			ConnectTimeoutMs: 1000,
		}

		if backend.Timeout > 0 {
			// convert from float to int
			cluster.ConnectTimeoutMs = int(backend.Timeout * 1000)
		}

		if backend.Retries > 0 {
			cluster.Hystrix = CircuitBreaker{
				MaxRetries:        backend.Retries,
				MaxConnections:    50,
				MaxRequests:       25,
				MaxPendingRequest: 5,
			}
		}

		clusters = append(clusters, cluster)
	}

	sort.Sort(ClustersByName(clusters))

	return clusters
}

func buildWeightKey(service string, tags []string) string {
	return fmt.Sprintf("%v.%v", service, buildServiceKey("_", tags))
}

func buildRoutes(ruleList []api.Rule) []Route {
	routes := []Route{}
	for _, rule := range ruleList {
		if rule.Route != nil {
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

				clusterName := buildServiceKey(backend.Name, backend.Tags)

				runtime := &Runtime{
					Key:     buildWeightKey(backend.Name, backend.Tags),
					Default: 0,
				}

				route := Route{
					Runtime:       runtime,
					Path:          path,
					Prefix:        prefix,
					PrefixRewrite: prefixRewrite,
					Cluster:       clusterName,
					Headers:       headers,
				}

				routes = append(routes, route)
			}
		}
	}

	return routes
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

func sanitizeRules(ruleList []api.Rule) {
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

func addDefaultRouteRules(ruleList []api.Rule, instances []api.ServiceInstance) []api.Rule {
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

// FIXME: doesn't check for name conflicts
// TODO: could be improved by using the full possible set of filenames.
func randFilename(prefix string) string {
	data := make([]byte, 16)
	for i := range data {
		data[i] = '0' + byte(rand.Intn(10))
	}

	return fmt.Sprintf("%s%s", prefix, data)
}

func buildFS(ruleList []api.Rule, workingDir string) error {
	type weightSpec struct {
		Service string
		Cluster string
		Weight  int
	}

	var weights []weightSpec
	for _, rule := range ruleList {
		if rule.Route != nil {
			w := 0
			for _, backend := range rule.Route.Backends {
				w += int(100 * backend.Weight)
				weight := weightSpec{
					Service: backend.Name,
					Cluster: buildServiceKey("_", backend.Tags),
					Weight:  w,
				}
				weights = append(weights, weight)
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(workingDir+runtimePath), configDirPerm); err != nil { // FIXME: hack
		return err
	}

	if err := os.MkdirAll(workingDir+runtimeVersionsPath, configDirPerm); err != nil {
		return err
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
						switch action.GetType() {
						case "delay":
							delay := action.Internal().(api.DelayAction)
							filter := Filter{
								Type: "decoder",
								Name: "fault",
								Config: &FilterFaultConfig{
									Delay: &DelayFilter{
										Type:     "fixed",
										Percent:  int(delay.Probability * 100),
										Duration: int(delay.Duration * 1000),
									},
									Headers: headers,
								},
							}
							filters = append(filters, filter)
						case "abort":
							abort := action.Internal().(api.AbortAction)
							filter := Filter{
								Type: "decoder",
								Name: "fault",
								Config: &FilterFaultConfig{
									Abort: &AbortFilter{
										Percent:    int(abort.Probability * 100),
										HTTPStatus: abort.ReturnCode,
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

func buildSourceName(service string, tags []string) string {
	return fmt.Sprintf("%v:%v", service, strings.Join(tags, ","))

}
