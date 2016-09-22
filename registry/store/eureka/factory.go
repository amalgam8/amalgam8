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
	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/store"
)

// Config encapsulates eureka configuration parameters
type Config struct {
	EurekaURLs []string
}

type eurekaFactory struct {
	client *eurekaClient
}

// New creates and initializes a Eureka catalog factory
func New(conf *Config) (store.CatalogFactory, error) {
	client, err := newEurekaClient(conf.EurekaURLs)
	if err != nil {
		return nil, err
	}

	return &eurekaFactory{client: client}, nil
}

func (f *eurekaFactory) CreateCatalog(namespace auth.Namespace) (store.Catalog, error) {
	// Eureka is not designed for multi-tenancy.
	// Therefore, we map the eureka catalog ONLY to the default namespace
	if namespace.String() == "" || namespace.String() == "default" {
		return newEurekaCatalog(f.client)
	}
	// nil catalog means that we don't map eureka catalog to this namespace
	return nil, nil
}
