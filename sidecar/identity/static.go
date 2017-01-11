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
	"fmt"

	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/pkg/network"
	"github.com/amalgam8/amalgam8/sidecar/config"
)

// staticProvider provides access to the identity manifested in the given configuration.
type staticProvider struct {
	conf *config.Config
}

func newStaticProvider(conf *config.Config) (Provider, error) {
	return &staticProvider{
		conf: conf,
	}, nil
}

func (sp *staticProvider) GetIdentity() (*api.ServiceInstance, error) {
	ip := sp.conf.Endpoint.Host
	if ip == "" {
		available := network.WaitForPrivateNetwork()
		if available {
			ip = network.GetPrivateIP().String()
		} else {
			logger.Warnf("Could not detect local IP address")
		}
	}

	var addr string
	if sp.conf.Endpoint.Port != 0 {
		addr = fmt.Sprintf("%s:%d", ip, sp.conf.Endpoint.Port)
	} else {
		addr = ip
	}

	return &api.ServiceInstance{
		ServiceName: sp.conf.Service.Name,
		Tags:        sp.conf.Service.Tags,
		Endpoint: api.ServiceEndpoint{
			Type:  sp.conf.Endpoint.Type,
			Value: addr,
		},
	}, nil
}
