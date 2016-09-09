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
	"encoding/json"
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

const repTimeout = time.Duration(7) * time.Second

type replicationHandler struct {
	conf     *replicatedConfig
	catalogs map[auth.Namespace]*replicatedCatalog
	logger   *log.Entry

	sync.Mutex
}

func newReplicationHandler(conf *replicatedConfig) *replicationHandler {
	var lentry = logging.GetLogger(module)

	handler := &replicationHandler{
		conf:     conf,
		catalogs: make(map[auth.Namespace]*replicatedCatalog),
		logger:   lentry,
	}

	return handler
}

// activate synchronizes state with remote peer, then start serving incoming replication events and/or sync requests.
// It should be called only once, after the replicationHandler has been created and properly initialized.
func (rh *replicationHandler) activate() error {
	rh.synchronize()

	go rh.launchSyncRequestListener()
	go rh.replicate()

	return nil
}

// synchronize starts a synchronization operation with remote peers.
// Synchronization is blocking, and should execute only once.
func (rh *replicationHandler) synchronize() {
	rh.logger.Info("Starting synchronization")

	var count int
	for inMsg := range rh.conf.rep.Sync(rh.conf.syncWaitTime) {
		if catalog, err := rh.getCatalog(inMsg.Namespace); err != nil {
			rh.logger.WithFields(log.Fields{
				"error": err,
			}).Errorf("Failed to synchronize incoming sync-event of %s catalog", inMsg.Namespace)
		} else {
			count++
			err = catalog.notifyChannel.Send(inMsg, repTimeout)
			if err != nil {
				rh.logger.WithFields(log.Fields{
					"error": err,
				}).Errorf("Failed to synchronize incoming sync-event of %s catalog", inMsg.Namespace)
			}
		}
	}
	rh.logger.Infof("Synchronization of %d elements has completed", count)
}

// replicate starts handling incoming replication events from remote peer
func (rh *replicationHandler) replicate() {
	rh.logger.Info("Start listening for replication notifications")

	for inMsg := range rh.conf.rep.Notification() {
		if catalog, err := rh.getCatalog(inMsg.Namespace); err != nil {
			rh.logger.WithFields(log.Fields{
				"error": err,
			}).Errorf("Failed to replicate incoming event of %s catalog", inMsg.Namespace)
		} else {
			catalog.notifyChannel.Send(inMsg, repTimeout)
		}
	}

	rh.logger.Info("Replication notifications listener has stopped")
}

func (rh *replicationHandler) launchSyncRequestListener() {
	rh.logger.Info("Start listening for Sync-Request")

	for outRequestChannel := range rh.conf.rep.SyncRequest() {
		go rh.handleSyncRequestJob(outRequestChannel)
	}
	rh.logger.Info("Sysnc-Request listener has stopped")
}

func (rh *replicationHandler) handleSyncRequestJob(reqChannel chan<- []byte) {
	rh.logger.Info("Starting Sync-Request job")

	var wg sync.WaitGroup
	for ns, catalog := range rh.catalogs {
		wg.Add(1)
		go func(namespace auth.Namespace, catalog *replicatedCatalog) {
			defer wg.Done()
			rh.logger.Infof("Starting a sync-request for %s catalog", namespace)
			services := catalog.ListServices(nil)

			for _, srv := range services {
				if instances, err := catalog.List(srv.ServiceName, nil); err != nil {
					rh.logger.WithFields(log.Fields{
						"error": err,
					}).Errorf("Sync Request with no instances for service %s", srv.ServiceName)
				} else {
					for _, inst := range instances {
						payload, _ := json.Marshal(inst)
						msg, _ := json.Marshal(&replicatedMsg{RepType: REGISTER, Payload: payload})
						out, _ := json.Marshal(map[string]interface{}{"Namespace": namespace, "Data": msg})
						reqChannel <- out
					}
				}
			}

			rh.logger.Infof("Complete a sync-request for %s catalog", namespace)
		}(ns, catalog)
	}
	wg.Wait()
	close(reqChannel)

	rh.logger.Info("Sync-Request job has completed")
}

func (rh *replicationHandler) getCatalog(namespace auth.Namespace) (*replicatedCatalog, error) {
	catalog := rh.lookupCatalog(namespace)
	if catalog != nil {
		return catalog, nil
	}

	// If there is no replicatedCatalog, we have to create a new one
	_, err := rh.conf.catalogMap.GetCatalog(namespace)
	if err != nil {
		rh.logger.WithFields(log.Fields{
			"error": err,
		}).Errorf("Failed to create a new %s catalog", namespace)
		return nil, err
	}

	catalog = rh.lookupCatalog(namespace)
	if catalog == nil {
		return nil, fmt.Errorf("Catalog %s does not exist", namespace)
	}

	return catalog, nil
}

func (rh *replicationHandler) addCatalog(namespace auth.Namespace, catalog *replicatedCatalog) {
	rh.Lock()
	defer rh.Unlock()

	rh.catalogs[namespace] = catalog
}

func (rh *replicationHandler) lookupCatalog(namespace auth.Namespace) *replicatedCatalog {
	rh.Lock()
	defer rh.Unlock()

	catalog, _ := rh.catalogs[namespace]
	return catalog
}
