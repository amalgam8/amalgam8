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

package kubernetes

import (
	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/store"
)

// K8sConfig encapsulates K8s configuration parameters
type K8sConfig struct {
	K8sURL   string
	K8sToken string
}

type k8sFactory struct {
	client *k8sClient
}

// New creates and initializes a K8s catalog factory
func New(conf *K8sConfig) (store.CatalogFactory, error) {
	client, err := newK8sClient(conf.K8sURL, conf.K8sToken)
	if err != nil {
		return nil, err
	}

	return &k8sFactory{client: client}, nil
}

func (f *k8sFactory) CreateCatalog(namespace auth.Namespace) (store.Catalog, error) {
	return newK8sCatalog(namespace, f.client)
}
