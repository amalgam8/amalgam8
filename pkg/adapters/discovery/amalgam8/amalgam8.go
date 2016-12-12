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

package amalgam8

import (
	"time"

	"github.com/amalgam8/amalgam8/pkg/adapters/discovery/cache"
	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/registry/client"
)

// RegistryConfig stores the configurable attributes of the Amalgam8 Registry adapter.
type RegistryConfig client.Config

// NewRegistryAdapter constructs a new ServiceRegistry adapter
// for the Amalgam8 Registry using the given configuration.
func NewRegistryAdapter(config RegistryConfig) (api.ServiceRegistry, error) {
	return client.New(client.Config(config))
}

// NewDiscoveryAdapter constructs a new ServiceDiscovery adapter
// for the Amalgam8 Registry using the given configuration.
func NewDiscoveryAdapter(config RegistryConfig) (api.ServiceDiscovery, error) {
	return client.New(client.Config(config))
}

// NewCachedDiscoveryAdapter constructs a new ServiceDiscovery adapter
// for the Amalgam8 Registry using the given configuration, and a local
// cache refreshed at the frequency specified by the given poll interval.
func NewCachedDiscoveryAdapter(config RegistryConfig, pollInterval time.Duration) (api.ServiceDiscovery, error) {
	registry, err := NewDiscoveryAdapter(config)
	if err != nil {
		return nil, err
	}

	return cache.New(registry, pollInterval)
}
