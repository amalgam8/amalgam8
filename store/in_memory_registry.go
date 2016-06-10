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
	"errors"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/amalgam8/registry/auth"
	"github.com/amalgam8/registry/replication"
	"github.com/amalgam8/registry/utils/logging"
)

const repTimeout = time.Duration(7) * time.Second

type inMemoryRegistry struct {
	namespaces map[auth.Namespace]Catalog
	sync.RWMutex
	rep replication.Replication

	logger *log.Entry
	conf   *Config
}

func newInMemoryRegistry(conf *Config, rep replication.Replication) Registry {
	var lentry = logging.GetLogger(module)

	if conf == nil {
		conf = DefaultConfig
	}

	registry := &inMemoryRegistry{
		namespaces: make(map[auth.Namespace]Catalog),
		rep:        rep,
		logger:     lentry,
		conf:       conf}

	if rep != nil {
		// Starts a synchronization operation with remote peers
		// Synchronization is blocking, executes once
		registry.synchronize()

		// Once Synchronization is complete, handle incoming replication events
		// Starts an end point to enable incoming replication events from remote peers to do
		// a registration to the local catalog
		go registry.replicate()
	}

	return registry
}

func (imr *inMemoryRegistry) create(namespace auth.Namespace) (Catalog, error) {
	imr.Lock()
	defer imr.Unlock()

	catalog, exists := imr.namespaces[namespace]

	if !exists {
		if imr.rep != nil {
			replicator, err := imr.rep.GetReplicator(namespace)
			if err != nil {
				return nil, err
			}
			catalog = newReplicatedCatalog(namespace, imr.conf, replicator)
		} else {
			catalog = newReplicatedCatalog(namespace, imr.conf, nil)
		}

		imr.namespaces[namespace] = catalog
		imr.logger.Infof("Add a new catalog [%s]", namespace)
	}

	return catalog, nil
}

func (imr *inMemoryRegistry) GetCatalog(namespace auth.Namespace) (Catalog, error) {
	catalog, err := imr.create(namespace)
	return catalog, err
}

func (imr *inMemoryRegistry) replicate() {
	imr.logger.Info("Start listening for replication notifications")
	if imr.rep != nil {
		// When we have replication,
		// we handle incoming message from remote peers
		// and redirecet to target catalog.
		for inMsg := range imr.rep.Notification() {
			if catalog, err := imr.create(inMsg.Namespace); err != nil {
				imr.logger.WithFields(log.Fields{
					"error": err,
				}).Errorf("Failed to replicate incoming event of %s catalog", inMsg.Namespace)
			} else {
				catalog.(*replicatedCatalog).notifyChannel.Send(inMsg, repTimeout)
			}
		}
	} else {
		imr.logger.WithFields(log.Fields{
			"error": errors.New("Replication is not supported"),
		}).Warn("Failed to start replication listener")
	}
	imr.logger.Info("Replication notifications listener has stopped")
}

func (imr *inMemoryRegistry) synchronize() {
	imr.logger.Info("Starting synchronization")
	var count int
	if imr.rep != nil {
		for inMsg := range imr.rep.Sync(imr.conf.SyncWaitTime) {
			if catalog, err := imr.create(inMsg.Namespace); err != nil {
				imr.logger.WithFields(log.Fields{
					"error": err,
				}).Errorf("Failed to synchronize incoming sync-event of %s catalog", inMsg.Namespace)
			} else {
				count++
				err = catalog.(*replicatedCatalog).notifyChannel.Send(inMsg, repTimeout)
				if err != nil {
					imr.logger.WithFields(log.Fields{
						"error": err,
					}).Errorf("Failed to synchronize incoming sync-event of %s catalog", inMsg.Namespace)
				}
			}
		}
	} else {
		imr.logger.WithFields(log.Fields{
			"error": "Replication is not supported",
		}).Warn("Can not start synchronization")
	}
	imr.logger.Infof("Synchronization of %d elements has completed", count)

	go imr.launchSyncRequestListener()
}

func (imr *inMemoryRegistry) launchSyncRequestListener() {
	imr.logger.Info("Start listening for Sync-Request")

	if imr.rep != nil {
		for outRequestChannel := range imr.rep.SyncRequest() {
			// TODO consider do-nothing if the registry is small enough
			go imr.handleSyncRequestJob(outRequestChannel)
		}
	} else {
		imr.logger.WithFields(log.Fields{
			"error": errors.New("Replication is not supported"),
		}).Warn("Can not start synchronization request end point")
	}
	imr.logger.Info("Sysnc-Request listener has stopped")
}

func (imr *inMemoryRegistry) handleSyncRequestJob(reqChannel chan<- []byte) {
	imr.logger.Info("Starting Sync-Request job")

	if imr.rep != nil {
		var wg sync.WaitGroup
		for namespace := range imr.namespaces {
			wg.Add(1)
			go func(ns auth.Namespace) {
				defer wg.Done()
				imr.logger.Infof("Starting a sync-request for %s catalog", ns)
				imr.namespaces[ns].(*replicatedCatalog).doSyncRequset(ns, reqChannel)
				imr.logger.Infof("Complete a sync-request for %s catalog", ns)
			}(namespace)
		}
		wg.Wait()
		close(reqChannel)
	}
	imr.logger.Info("Sync-Request job has completed")
}
