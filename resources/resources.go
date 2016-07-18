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

package resources

import (
	"strconv"
	"time"
)

// BasicEntry TODO
type BasicEntry struct {
	ID      string  `json:"_id"`
	Rev     string  `json:"_rev,omitempty"`
	IV      string  `json:"iv"`
	Version float64 `json:"version"`
}

// IDRev TODO
func (e *BasicEntry) IDRev() (string, string) {
	return e.ID, e.Rev
}

// SetRev TODO
func (e *BasicEntry) SetRev() {
	if e.Rev == "" {
		e.Rev = "0"
	}
	i, _ := strconv.Atoi(e.Rev)
	i++
	e.Rev = strconv.Itoa(i)
}

// SetIV TODO
func (e *BasicEntry) SetIV(iv string) {
	e.IV = iv
}

// GetIV TODO
func (e *BasicEntry) GetIV() string {
	return e.IV
}

// MetaData service instance metadata
type MetaData struct {
	Version string
}

// ServiceCatalog TODO
type ServiceCatalog struct {
	Services   []Service
	LastUpdate time.Time
}

// Service TODO
type Service struct {
	Name      string
	Endpoints []Endpoint
}

// Endpoint TODO
type Endpoint struct {
	Type     string
	Value    string
	Metadata MetaData
}

// ByService TODO
type ByService []Service

// Len TODO
func (a ByService) Len() int {
	return len(a)
}

// Swap TODO
func (a ByService) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Less TODO
func (a ByService) Less(i, j int) bool {
	return a[i].Name < a[j].Name
}

// ByEndpoint TODO
type ByEndpoint []Endpoint

// Len TODO
func (a ByEndpoint) Len() int {
	return len(a)
}

// Swap TODO
func (a ByEndpoint) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Less TODO
func (a ByEndpoint) Less(i, j int) bool {
	if a[i].Value == a[j].Value {
		if a[i].Type == a[j].Type {
			return a[i].Metadata.Version < a[j].Metadata.Version
		}
		return a[i].Type < a[j].Type
	}
	return a[i].Value < a[j].Value
}

// Registry TODO
type Registry struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

// Kafka TODO
type Kafka struct {
	APIKey   string   `json:"api_key"`
	AdminURL string   `json:"admin_url"`
	RestURL  string   `json:"rest_url"`
	Brokers  []string `json:"brokers"`
	User     string   `json:"user"`
	Password string   `json:"password"`
	SASL     bool     `json:"sasl"`
}

// Credentials TODO
type Credentials struct {
	Kafka    Kafka    `json:"kafka"`
	Registry Registry `json:"registry"`
}

// TenantEntry TODO
type TenantEntry struct {
	BasicEntry
	TenantToken    string
	ProxyConfig    ProxyConfig
	ServiceCatalog ServiceCatalog
}

// ProxyConfig TODO
type ProxyConfig struct {
	LoadBalance string      `json:"load_balance"`
	Filters     Filters     `json:"filters"`
	Credentials Credentials `json:"credentials"`
}

// Filters TODO
type Filters struct {
	Rules    []Rule    `json:"rules"`
	Versions []Version `json:"versions"`
}

// Rule TODO
type Rule struct {
	ID               string  `json:"id"`
	Source           string  `json:"source"`
	Destination      string  `json:"destination"`
	Header           string  `json:"header"`
	Pattern          string  `json:"pattern"`
	Delay            float64 `json:"delay"`
	DelayProbability float64 `json:"delay_probability"`
	AbortProbability float64 `json:"abort_probability"`
	ReturnCode       int     `json:"return_code"`
}

// Version TODO
type Version struct {
	Service   string `json:"service"`
	Default   string `json:"default"`
	Selectors string `json:"selectors"`
}

// TenantInfo JSON object for credentials and metadata of a tenant
type TenantInfo struct {
	Credentials Credentials `json:"credentials"`
	LoadBalance string      `json:"load_balance"`
	Filters     Filters     `json:"filters"`
}

// VersionedUpstreams contains upstreams by version
type VersionedUpstreams struct {
	UpstreamName string   `json:"name"`
	Upstreams    []string `json:"upstreams"`
}

// NGINXJson TODO
type NGINXJson struct {
	Upstreams map[string]NGINXUpstream `json:"upstreams"`
	Services  map[string]NGINXService  `json:"services"`
	Faults    []NGINXFault             `json:"faults,omitempty"`
}

// NGINXService TODO
type NGINXService struct {
	Default   string `json:"default"`
	Selectors string `json:"selectors,omitempty"`
	Type      string `json:"type"`
}

// NGINXUpstream TODO
type NGINXUpstream struct {
	Upstreams []NGINXEndpoint `json:"servers"`
}

// NGINXEndpoint TODO
type NGINXEndpoint struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// NGINXFault TODO
type NGINXFault struct {
	Source           string  `json:"source"`
	Destination      string  `json:"destination"`
	Header           string  `json:"header"`
	Pattern          string  `json:"pattern"`
	Delay            float64 `json:"delay"`
	DelayProbability float64 `json:"delay_probability"`
	AbortProbability float64 `json:"abort_probability"`
	AbortCode        int     `json:"return_code"`
}
