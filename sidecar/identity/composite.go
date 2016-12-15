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

package identity

import (
	"encoding/json"
	"time"

	"github.com/amalgam8/amalgam8/pkg/api"
)

// compositeProvider provides access to the identity combined from multiple identity providers.
type compositeProvider struct {
	providers  []Provider
	identities []*api.ServiceInstance
}

func newCompositeProvider(providers ...Provider) (Provider, error) {
	return &compositeProvider{
		providers: providers,
	}, nil
}

func (cp *compositeProvider) GetIdentity() (*api.ServiceInstance, error) {
	identities, err := cp.getProvidersIdentities()
	if err != nil {
		return nil, err
	}
	if len(identities) == 0 {
		return nil, nil
	}

	return &api.ServiceInstance{
		ID:            getID(identities),
		ServiceName:   getServiceName(identities),
		Endpoint:      getEndpoint(identities),
		Status:        getStatus(identities),
		Tags:          getTags(identities),
		Metadata:      getMetadata(identities),
		TTL:           getTTL(identities),
		LastHeartbeat: getLastHeartbeat(identities),
	}, nil
}

func (cp *compositeProvider) getProvidersIdentities() ([]*api.ServiceInstance, error) {
	identities := make([]*api.ServiceInstance, 0, len(cp.providers))
	for _, provider := range cp.providers {
		var err error
		identity, err := provider.GetIdentity()
		if err != nil {
			return nil, err
		}
		if identity != nil {
			identities = append(identities, identity)
		}
	}
	return identities, nil
}

func getID(identities []*api.ServiceInstance) string {
	var zeroValue string
	for _, identity := range identities {
		if identity.ID != zeroValue {
			return identity.ID
		}
	}
	return zeroValue
}

func getServiceName(identities []*api.ServiceInstance) string {
	var zeroValue string
	for _, identity := range identities {
		if identity.ServiceName != zeroValue {
			return identity.ServiceName
		}
	}
	return zeroValue
}

func getEndpoint(identities []*api.ServiceInstance) api.ServiceEndpoint {
	var zeroValue api.ServiceEndpoint
	for _, identity := range identities {
		if identity.Endpoint != zeroValue {
			return identity.Endpoint
		}
	}
	return zeroValue
}

func getStatus(identities []*api.ServiceInstance) string {
	var zeroValue string
	for _, identity := range identities {
		if identity.Status != zeroValue {
			return identity.Status
		}
	}
	return zeroValue
}

func getTags(identities []*api.ServiceInstance) []string {
	var zeroValue []string
	for _, identity := range identities {
		if identity.Tags != nil {
			return identity.Tags
		}
	}
	return zeroValue
}

func getMetadata(identities []*api.ServiceInstance) json.RawMessage {
	var zeroValue json.RawMessage
	for _, identity := range identities {
		if identity.Metadata != nil {
			return identity.Metadata
		}
	}
	return zeroValue
}

func getTTL(identities []*api.ServiceInstance) int {
	var zeroValue int
	for _, identity := range identities {
		if identity.TTL != zeroValue {
			return identity.TTL
		}
	}
	return zeroValue
}

func getLastHeartbeat(identities []*api.ServiceInstance) time.Time {
	var zeroValue time.Time
	for _, identity := range identities {
		if identity.LastHeartbeat != zeroValue {
			return identity.LastHeartbeat
		}
	}
	return zeroValue
}
