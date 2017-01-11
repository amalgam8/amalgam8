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
	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
	"github.com/amalgam8/amalgam8/sidecar/config"
	"k8s.io/client-go/kubernetes"
)

var logger = logging.GetLogger("IDENTITY")

// Provider provides access to the identity of the locally running service instance.
type Provider interface {
	GetIdentity() (*api.ServiceInstance, error)
}

// New creates an identity provider which exposes the identity specified by
// the given configuration parameters (service name, tags, endpoint, etc).
// If a pod name is specified, then missing parameters may be automatically fetched from Kubernetes API
// using the given client.
func New(conf *config.Config, kubeClient kubernetes.Interface) (Provider, error) {
	var provider Provider

	provider, err := newStaticProvider(conf)
	if err != nil {
		return nil, err
	}

	if conf.Kubernetes.PodName != "" {
		kubeProvider, err := newKubernetesProvider(conf.Kubernetes.PodName, conf.Kubernetes.Namespace, kubeClient)
		if err != nil {
			return nil, err
		}
		provider, err = newCompositeProvider(provider, kubeProvider)
		if err != nil {
			return nil, err
		}
	}

	return provider, nil
}
