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

// TenantEntry TODO
type TenantEntry struct {
	BasicEntry
	ProxyConfig ProxyConfig
	LastUpdate  time.Time
}

// ProxyConfig TODO
type ProxyConfig struct {
	LoadBalance string  `json:"load_balance"`
	Filters     Filters `json:"filters"`
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
	LoadBalance string  `json:"load_balance"`
	Filters     Filters `json:"filters"`
}
