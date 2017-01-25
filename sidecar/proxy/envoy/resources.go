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

// AbortFilter definition.
type AbortFilter struct {
	Percent    int `json:"abort_percent,omitempty"`
	HTTPStatus int `json:"http_status,omitempty"`
}

// DelayFilter definition.
type DelayFilter struct {
	Type     string `json:"type,omitempty"`
	Percent  int    `json:"fixed_delay_percent,omitempty"`
	Duration int    `json:"fixed_duration_ms,omitempty"`
}

// Header definition.
// See: https://lyft.github.io/envoy/docs/configuration/http_filters/fault_filter.html#config-http-filters-fault-injection-headers
type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Regex bool   `json:"regex"`
}

// FilterFaultConfig definition.
// See: https://lyft.github.io/envoy/docs/configuration/http_filters/fault_filter.html
type FilterFaultConfig struct {
	Abort   *AbortFilter `json:"abort,omitempty"`
	Delay   *DelayFilter `json:"delay,omitempty"`
	Headers []Header     `json:"headers,omitempty"`
}

// FilterRouterConfig definition.
type FilterRouterConfig struct {
	DynamicStats bool `json:"dynamic_stats"`
}

// Filter definition.
type Filter struct {
	Type   string      `json:"type"`
	Name   string      `json:"name"`
	Config interface{} `json:"config"`
}

// Runtime definition.
type Runtime struct {
	Key     string `json:"key"`
	Default int    `json:"default"`
}

// Route definition.
// See: https://lyft.github.io/envoy/docs/configuration/http_conn_man/route_config/route.html#config-http-conn-man-route-table-route
type Route struct {
	Runtime       *Runtime `json:"runtime,omitempty"`
	Prefix        string   `json:"prefix"`
	PrefixRewrite string   `json:"prefix_rewrite"`
	Cluster       string   `json:"cluster"`
	Headers       []Header `json:"headers,omitempty"`
}

// VirtualHost definition.
// See: https://lyft.github.io/envoy/docs/configuration/http_conn_man/route_config/vhost.html#config-http-conn-man-route-table-vhost
type VirtualHost struct {
	Name    string   `json:"name"`
	Domains []string `json:"domains"`
	Routes  []Route  `json:"routes"`
}

// RouteConfig definition.
// See: https://lyft.github.io/envoy/docs/configuration/http_conn_man/route_config/route_config.html#config-http-conn-man-route-table
type RouteConfig struct {
	VirtualHosts []VirtualHost `json:"virtual_hosts"`
}

// AccessLog definition.
type AccessLog struct {
	Path   string `json:"path"`
	Format string `json:"format,omitempty"`
	Filter string `json:"filter,omitempty"`
}

// NetworkFilterConfig definition.
type NetworkFilterConfig struct {
	CodecType         string      `json:"codec_type"`
	StatPrefix        string      `json:"stat_prefix"`
	GenerateRequestID bool        `json:"generate_request_id"`
	UserAgent         bool        `json:"add_user_agent"`
	RouteConfig       RouteConfig `json:"route_config"`
	Filters           []Filter    `json:"filters"`
	AccessLog         []AccessLog `json:"access_log"`
}

// NetworkFilter definition.
// See: https://lyft.github.io/envoy/docs/configuration/listeners/filters.html#config-listener-filters
type NetworkFilter struct {
	Type   string              `json:"type"`
	Name   string              `json:"name"`
	Config NetworkFilterConfig `json:"config"`
}

// Listener definition.
// See: https://lyft.github.io/envoy/docs/configuration/listeners/listeners.html#config-listeners
type Listener struct {
	Port    int             `json:"port"`
	Filters []NetworkFilter `json:"filters"`
}

// Admin definition.
// See: https://lyft.github.io/envoy/docs/configuration/overview/admin.html#config-admin
type Admin struct {
	AccessLogPath string `json:"access_log_path"`
	Port          int    `json:"port"`
}

// Host definition.
type Host struct {
	URL string `json:"url"`
}

// Cluster definition.
// See: https://lyft.github.io/envoy/docs/configuration/cluster_manager/cluster.html#config-cluster-manager-cluster
type Cluster struct {
	Name                     string `json:"name"`
	ServiceName              string `json:"service_name,omitempty"`
	ConnectTimeoutMs         int    `json:"connect_timeout_ms"`
	Type                     string `json:"type"`
	LbType                   string `json:"lb_type"`
	MaxRequestsPerConnection int    `json:"max_requests_per_connection,omitempty"`
	Hosts                    []Host `json:"hosts,omitempty"`
}

// ClustersByName implements name based sort for clusters.
type ClustersByName []Cluster

// Len length.
func (s ClustersByName) Len() int {
	return len(s)
}

// Swap elements.
func (s ClustersByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less compare elements.
func (s ClustersByName) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

// SDS definition.
// See: https://lyft.github.io/envoy/docs/configuration/cluster_manager/sds.html#config-cluster-manager-sds
type SDS struct {
	Cluster        Cluster `json:"cluster"`
	RefreshDelayMs int     `json:"refresh_delay_ms"`
}

// ClusterManager definition.
// See: https://lyft.github.io/envoy/docs/configuration/cluster_manager/cluster_manager.html#config-cluster-manager
type ClusterManager struct {
	Clusters []Cluster `json:"clusters"`
	SDS      SDS       `json:"sds"`
}

// RootRuntime definition.
// See: https://lyft.github.io/envoy/docs/configuration/overview/overview.html
type RootRuntime struct {
	SymlinkRoot          string `json:"symlink_root"`
	Subdirectory         string `json:"subdirectory"`
	OverrideSubdirectory string `json:"override_subdirectory,omitempty"`
}

// Config definition.
// See: https://lyft.github.io/envoy/docs/configuration/overview/overview.html
type Config struct {
	RootRuntime    RootRuntime    `json:"runtime"`
	Listeners      []Listener     `json:"listeners"`
	Admin          Admin          `json:"admin"`
	ClusterManager ClusterManager `json:"cluster_manager"`
}
