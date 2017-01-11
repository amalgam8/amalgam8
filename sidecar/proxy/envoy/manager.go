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

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/pkg/api"
)

// EnvoyConfigPath path to envoy config file
const EnvoyConfigPath = "/etc/envoy/envoy.json"

// Manager for updating envoy proxy configuration.
type Manager interface {
	Update(instances []api.ServiceInstance, rules []api.Rule) error
}

// NewManager creates new instance
func NewManager(serviceName string, tags []string) Manager {
	return &manager{
		serviceName: serviceName,
		tags:        tags,
		service: NewService(ServiceConfig{
			DrainTimeSeconds:          3,
			ParentShutdownTimeSeconds: 5,
			EnvoyConfig:               EnvoyConfigPath,
		}),
	}
}

type manager struct {
	serviceName string
	tags        []string
	service     Service
}

func (m *manager) Update(instances []api.ServiceInstance, rules []api.Rule) error {
	conf, err := generateConfig(rules, instances, m.serviceName, m.tags)
	if err != nil {
		return err
	}

	if err := writeConfigFile(conf); err != nil {
		return err
	}

	if err := m.service.Reload(); err != nil {
		return err
	}

	return nil
}

func writeConfigFile(conf Config) error {
	file, err := os.Create(EnvoyConfigPath)
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

func generateConfig(rules []api.Rule, instances []api.ServiceInstance, serviceName string, tags []string) (Config, error) {
	sanitizeRules(rules)
	rules = addDefaultRouteRules(rules, instances)

	clusters := buildClusters(rules)
	routes := buildRoutes(rules)

	filters := buildFaults(rules, serviceName, tags)

	if err := buildFS(rules); err != nil {
		return Config{}, err
	}

	return Config{
		RootRuntime: RootRuntime{
			SymlinkRoot:  runtimePath,
			Subdirectory: "traffic_shift",
		},
		Listeners: []Listener{
			{
				Port: 6379,
				Filters: []NetworkFilter{
					{
						Type: "read",
						Name: "http_connection_manager",
						Config: NetworkFilterConfig{
							CodecType:  "auto",
							StatPrefix: "ingress_http",
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
									Path: "/var/log/envoy_access.log",
								},
							},
						},
					},
				},
			},
		},
		Admin: Admin{
			AccessLogPath: "/var/log/envoy_admin.log",
			Port:          8001,
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
							URL: "tcp://127.0.0.1:6500",
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
	ctrl      = '_' // Control character
	ctrlSplit = 's' // Split control character
)

// buildServiceKey returns key containing service name and its tags
func buildServiceKey(service string, tags []string) string {
	sort.Strings(tags) // FIXME: by reference

	// Guesstimate the required buffer capacity by assuming the typical individual tag length is 10 or less and that
	// the output will at most double in size.
	c := 2 * (len(service) + 10*len(tags))
	buf := bytes.NewBuffer(make([]byte, 0, c))

	// Writes an escaped version of the input string to the buffer.
	escape := func(s string, buf *bytes.Buffer) {
		data := []byte(s)
		for i := range data {
			if data[i] == ctrl {
				buf.WriteByte(ctrl)
			}
			buf.WriteByte(data[i])
		}
	}

	// Write escaped service and tags to the buffer separated by split control characters.
	escape(service, buf)
	for i := range tags {
		buf.Write([]byte{ctrl, ctrlSplit})
		escape(tags[i], buf)
	}

	return buf.String()
}

// ParseServiceKey returns service name and its tags
func ParseServiceKey(key string) (string, []string) {
	res := make([]string, 0, 6) // We guesstimate that most keys are composed of at most 1 service name + 5 tags.
	buf := bytes.NewBuffer(make([]byte, 0, len(key)))
	data := []byte(key)

	i := 0
	for i = 0; i < len(data)-1; i++ {
		if data[i] == ctrl {
			switch data[i+1] {
			case ctrl:
				buf.WriteByte(ctrl)
			case ctrlSplit:
				res = append(res, buf.String())
				buf = bytes.NewBuffer(make([]byte, 0, len(key)))
			default:
				// FIXME: behavior?
				logrus.WithField("character", data[i+1]).Warn("Unrecognized control character")
			}
			i++
		} else {
			buf.WriteByte(data[i])
		}
	}

	// If the 2nd to last byte was not a control character we need to write the last byte.
	if i == len(data)-1 {
		buf.WriteByte(data[i])
	}
	res = append(res, buf.String())

	service := res[0]
	tags := res[1:]

	return service, tags
}

func buildClusters(rules []api.Rule) []Cluster {
	clusterMap := make(map[string]struct{})
	for _, rule := range rules {
		if rule.Route != nil {
			for _, backend := range rule.Route.Backends {
				key := buildServiceKey(backend.Name, backend.Tags)
				clusterMap[key] = struct{}{}
			}
		}
	}

	clusters := make([]Cluster, 0, len(clusterMap))
	for clusterName := range clusterMap {
		cluster := Cluster{
			Name:             clusterName,
			ServiceName:      clusterName,
			Type:             "sds",
			LbType:           "round_robin",
			ConnectTimeoutMs: 1000,
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
				clusterName := buildServiceKey(backend.Name, backend.Tags)

				runtime := &Runtime{
					Key:     buildWeightKey(backend.Name, backend.Tags),
					Default: 0,
				}

				route := Route{
					Runtime:       runtime,
					Prefix:        "/" + backend.Name + "/",
					PrefixRewrite: "/",
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

const (
	runtimePath         = "/etc/envoy/runtime/routing"
	runtimeVersionsPath = "/etc/envoy/routing_versions"

	configDirPerm  = 0775
	configFilePerm = 0664
)

// FIXME: doesn't check for name conflicts
// TODO: could be improved by using the full possible set of filenames.
func randFilename(prefix string) string {
	data := make([]byte, 16)
	for i := range data {
		data[i] = '0' + byte(rand.Intn(10))
	}

	return fmt.Sprintf("%s%s", prefix, data)
}

func buildFS(ruleList []api.Rule) error {
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

	if err := os.MkdirAll(filepath.Dir(runtimePath), configDirPerm); err != nil { // FIXME: hack
		return err
	}

	if err := os.MkdirAll(runtimeVersionsPath, configDirPerm); err != nil {
		return err
	}

	dirName, err := ioutil.TempDir(runtimeVersionsPath, "")
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

	oldRuntime, err := os.Readlink(runtimePath)
	if err != nil && !os.IsNotExist(err) { // Ignore error from symlink not existing.
		return err
	}

	tmpName := randFilename("./")

	if err := os.Symlink(dirName, tmpName); err != nil {
		return err
	}

	// Atomically replace the runtime symlink
	if err := os.Rename(tmpName, runtimePath); err != nil {
		return err
	}

	success = true

	// Clean up the old config FS if necessary
	// TODO: make this safer
	if oldRuntime != "" {
		oldRuntimeDir := filepath.Dir(oldRuntime)
		if filepath.Clean(oldRuntimeDir) == filepath.Clean(runtimeVersionsPath) {
			toDelete := filepath.Join(runtimeVersionsPath, filepath.Base(oldRuntime))
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
				})
			}

			if rule.Match.Source != nil && rule.Match.Source.Name == serviceName {
				isSubset := true
				for _, tag := range rule.Tags {
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
