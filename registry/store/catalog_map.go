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

// Package store defines and implements a backend store for the registry
package store

import (
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

const (
	module = "STORE"
)

// CatalogMap represents the interface of the Service store
type CatalogMap interface {
	GetCatalog(auth.Namespace) (Catalog, error)
}

type catalogMap struct {
	conf           *Config
	catalogs       map[auth.Namespace]Catalog
	catalogFactory CatalogFactory

	logger *log.Entry
	sync.Mutex
}

// New creates a new CatalogMap instance, bounded with the specified configuration
func New(conf *Config) CatalogMap {
	var lentry = logging.GetLogger(module)
	var factory CatalogFactory

	if conf == nil {
		conf = DefaultConfig
	}

	cmap := &catalogMap{
		conf:     conf,
		catalogs: make(map[auth.Namespace]Catalog),
		logger:   lentry,
	}

	if conf.Store == "redis" {
		externalConfig := &externalConfig{
			defaultTTL:        conf.DefaultTTL,
			minimumTTL:        conf.MinimumTTL,
			maximumTTL:        conf.MaximumTTL,
			namespaceCapacity: conf.NamespaceCapacity,
			store:             conf.Store,
			address:           conf.StoreAddr,
			password:          conf.StorePassword,
			database:          conf.StoreDatabase,
		}
		externalFactory := newExternalFactory(externalConfig)
		factory = externalFactory
		conf.Replication = nil
	} else {
		// The InMemory catalog is the Read-Write catalog.
		inmemConfig := &inMemoryConfig{
			defaultTTL:        conf.DefaultTTL,
			minimumTTL:        conf.MinimumTTL,
			maximumTTL:        conf.MaximumTTL,
			namespaceCapacity: conf.NamespaceCapacity,
		}
		inmemFactory := newInMemoryFactory(inmemConfig)
		factory = inmemFactory
	}

	if conf.Replication != nil {
		repConfig := &replicatedConfig{
			syncWaitTime: conf.SyncWaitTime,
			rep:          conf.Replication,
			catalogMap:   cmap,
			localFactory: factory,
		}
		repFactory := newReplicatedFactory(repConfig)
		defer repFactory.activate()
		factory = repFactory
	}

	if len(conf.Extensions) > 0 {
		factories := make([]CatalogFactory, len(conf.Extensions)+1)
		factories[0] = factory
		for i, f := range conf.Extensions {
			factories[i+1] = f
		}
		factory = newMultiFactory(&multiConfig{factories: factories})
	}

	cmap.catalogFactory = factory
	return cmap
}

func (cm *catalogMap) GetCatalog(namespace auth.Namespace) (Catalog, error) {
	cm.Lock()
	defer cm.Unlock()

	catalog, exists := cm.catalogs[namespace]
	if exists {
		return catalog, nil
	}

	catalog, err := cm.catalogFactory.CreateCatalog(namespace)
	if err != nil {
		cm.logger.WithFields(log.Fields{
			"error": err,
		}).Errorf("Failed to create a new catalog [%s]", namespace)
		return nil, err
	}

	cm.catalogs[namespace] = catalog
	cm.logger.Infof("A new catalog [%s] has beed created", namespace)

	return catalog, nil
}
