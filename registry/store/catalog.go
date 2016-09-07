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

package store

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Predicate for filtering returned instances
type Predicate func(si *ServiceInstance) bool

// Catalog for managing instances within a registry namespace
type Catalog interface {
	Register(si *ServiceInstance) (*ServiceInstance, error)
	Deregister(instanceID string) (*ServiceInstance, error)
	Renew(instanceID string) (*ServiceInstance, error)
	SetStatus(instanceID, status string) (*ServiceInstance, error)

	Instance(instanceID string) (*ServiceInstance, error)
	List(serviceName string, predicate Predicate) ([]*ServiceInstance, error)
	ListServices(predicate Predicate) []*Service
}

const (
	serviceNameMaxLength int = 64
	valueMaxLength       int = 64
	statusMaxLength      int = 32
	metadataMaxLength    int = 1024

	instancesMetricName  = "store.instances.count"
	expirationMetricName = "store.instances.expiration"
	lifetimeMetricName   = "store.instances.lifetime"

	metadataLengthMetricName    = "store.metadata.length"
	metadataInstancesMetricName = "store.metadata.instances"
	tagsLengthMetricName        = "store.tags.length"
	tagsInstancesMetricName     = "store.tags.instances"
)

func computeInstanceID(si *ServiceInstance) string {
	// The ID is deterministically computed for each catalog,
	// This is necessary to support replication, and duplicate registration request accross nodes in the sd cluster
	hash := sha256.New()
	hash.Write([]byte(strings.Join([]string{si.ServiceName, si.Endpoint.Type, si.Endpoint.Value}, "/")))
	//hash.Write([]byte(time.Now().String()))
	md := hash.Sum(nil)
	mdStr := hex.EncodeToString(md)
	return mdStr[:16]
}
