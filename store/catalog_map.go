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

	"github.com/amalgam8/registry/auth"
	"github.com/amalgam8/registry/utils/logging"
)

const (
	module string = "STORE"
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

	inmemConfig := &inmemConfig{
		defaultTTL:        conf.DefaultTTL,
		minimumTTL:        conf.MinimumTTL,
		maximumTTL:        conf.MaximumTTL,
		namespaceCapacity: conf.NamespaceCapacity,
	}
	inmemFactory := newInmemFactory(inmemConfig)
	factory = inmemFactory

	if conf.Replication != nil {
		repConfig := &replicatedConfig{
			syncWaitTime: conf.SyncWaitTime,
			rep:          conf.Replication,
			catalogMap:   cmap,
			localFactory: inmemFactory,
		}
		factory = newReplicatedFactory(repConfig)
	}

	if len(conf.AddsOn) > 0 {
		factories := make([]CatalogFactory, len(conf.AddsOn)+1)
		factories[0] = factory
		for i, f := range conf.AddsOn {
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
		return nil, err
	}

	cm.catalogs[namespace] = catalog
	cm.logger.Infof("Add a new catalog [%s]", namespace)

	return catalog, nil
}
