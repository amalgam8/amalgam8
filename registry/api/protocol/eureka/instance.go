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

package eureka

import (
	"encoding/json"
	"fmt"
)

const (
	defaultDurationInt uint32 = 90
	extEureka                 = "eureka"
	extVIP                    = "vipAddress"
)

// Port encapsulates information needed for a port information
type Port struct {
	Enabled string      `json:"@enabled,omitempty"`
	Value   interface{} `json:"$,omitempty"`
}

// DatacenterInfo encapsulates information needed for a datacenter information
type DatacenterInfo struct {
	Class    string            `json:"@class,omitempty"`
	Name     string            `json:"name,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// LeaseInfo encapsulates information needed for a lease information
type LeaseInfo struct {
	RenewalInt     uint32 `json:"renewalIntervalInSecs,omitempty"`
	DurationInt    uint32 `json:"durationInSecs,omitempty"`
	RegistrationTs int64  `json:"registrationTimestamp,omitempty"`
	LastRenewalTs  int64  `json:"lastRenewalTimestamp,omitempty"`
}

// Instance encapsulates information needed for a service instance information
type Instance struct {
	ID            string          `json:"instanceId,omitempty"`
	HostName      string          `json:"hostName,omitempty"`
	Application   string          `json:"app,omitempty"`
	GroupName     string          `json:"appGroupName,omitempty"`
	IPAddr        string          `json:"ipAddr,omitempty"`
	VIPAddr       string          `json:"vipAddress,omitempty"`
	SecVIPAddr    string          `json:"secureVipAddress,omitempty"`
	Status        string          `json:"status,omitempty"`
	OvrStatus     string          `json:"overriddenstatus,omitempty"`
	CountryID     int             `json:"countryId,omitempty"`
	Port          *Port           `json:"port,omitempty"`
	SecPort       *Port           `json:"securePort,omitempty"`
	HomePage      string          `json:"homePageUrl,omitempty"`
	StatusPage    string          `json:"statusPageUrl,omitempty"`
	HealthCheck   string          `json:"healthCheckUrl,omitempty"`
	Datacenter    *DatacenterInfo `json:"dataCenterInfo,omitempty"`
	Lease         *LeaseInfo      `json:"leaseInfo,omitempty"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
	CordServer    interface{}     `json:"isCoordinatingDiscoveryServer,omitempty"`
	LastUpdatedTs interface{}     `json:"lastUpdatedTimestamp,omitempty"`
	LastDirtyTs   interface{}     `json:"lastDirtyTimestamp,omitempty"`
	ActionType    string          `json:"actionType,omitempty"`
}

// InstanceWrapper encapsulates information needed for a service instance registration
type InstanceWrapper struct {
	Inst *Instance `json:"instance,omitempty"`
}

// String output the structure
func (ir *Instance) String() string {
	mtlen := 0
	if ir.Metadata != nil {
		mtlen = len(ir.Metadata)
	}
	return fmt.Sprintf("vip_addres: %s, endpoint: %s:%d, hostname: %s, status: %s, metadata: %d",
		ir.VIPAddr, ir.IPAddr, ir.Port.Value, ir.HostName, ir.Status, mtlen)
}
