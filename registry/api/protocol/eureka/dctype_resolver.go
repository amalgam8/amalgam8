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
	"errors"
	"fmt"
	"strings"
)

// UniqueIdentifier indicates the unique ID of the datacenter information
type UniqueIdentifier interface {
	GetID(dcinfo *DatacenterInfo) string
}

var amzInfo amazonInfo
var slInfo softlayerInfo

func getInstanceID(inst *Instance) (string, error) {
	var id string

	if inst == nil {
		return "", errors.New("instance is nil")
	}

	// The default identifier is the Hostname
	id = inst.HostName
	if inst.Datacenter == nil {
		return id, nil
	}

	name := strings.ToLower(inst.Datacenter.Name)

	switch name {
	case "myown", "netflix":
	case "amazon":
		if uid := amzInfo.GetID(inst.Datacenter); uid != "" {
			id = uid
		}
	case "softlayer":
		if uid := slInfo.GetID(inst.Datacenter); uid != "" {
			id = uid
		}
	default:
		return "", fmt.Errorf("unknown datacenter name [%s]", inst.Datacenter.Name)
	}

	return id, nil
}
