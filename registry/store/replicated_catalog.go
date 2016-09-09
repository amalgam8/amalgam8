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
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/replication"
	"github.com/amalgam8/amalgam8/registry/utils/channels"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

type replicatedConfig struct {
	syncWaitTime time.Duration
	catalogMap   CatalogMap
	rep          replication.Replication
	localFactory CatalogFactory
}

type replicatedFactory struct {
	conf       *replicatedConfig
	repHandler *replicationHandler
}

func newReplicatedFactory(conf *replicatedConfig) *replicatedFactory {
	rh := newReplicationHandler(conf)
	return &replicatedFactory{conf: conf, repHandler: rh}
}

func (f *replicatedFactory) CreateCatalog(namespace auth.Namespace) (Catalog, error) {
	repCatalog, err := newReplicatedCatalog(namespace, f.conf)
	if err != nil {
		return nil, err
	}
	f.repHandler.addCatalog(namespace, repCatalog)
	return repCatalog, nil
}

// activate synchronizes state with remote peer, then start serving incoming replication events and/or sync requests.
// It should be called only once, after the replicatedFactory has been created and properly initialized.
func (f *replicatedFactory) activate() error {
	return f.repHandler.activate()
}

type replicatedCatalog struct {
	replicator    replication.Replicator
	notifyChannel channels.ChannelTimeout
	local         Catalog

	logger *log.Entry
}

type replicationType int

type replicatedMsg struct {
	RepType replicationType
	Payload []byte
}

type replicatedStatus struct {
	InstanceID string
	Status     string
}

// Enumeration implementation for the replication actions types
const (
	REGISTER replicationType = iota
	DEREGISTER
	RENEW
	SETSTATUS
	READREPAIR
)

var replicationActionTypes = [...]string{
	"REGISTER",
	"DEREGISTER",
	"RENEW",
	"SETSTATUS",
	"READREPAIR",
}

func (t replicationType) String() string {
	return replicationActionTypes[t]
}

func newReplicatedCatalog(namespace auth.Namespace, conf *replicatedConfig) (*replicatedCatalog, error) {
	logger := logging.GetLogger(module).WithFields(log.Fields{"namespace": namespace})

	lc, err := conf.localFactory.CreateCatalog(namespace)
	if err != nil {
		return nil, err
	}

	replicator, err := conf.rep.GetReplicator(namespace)
	if err != nil {
		return nil, err
	}

	rpc := &replicatedCatalog{
		local:         lc,
		replicator:    replicator,
		notifyChannel: channels.NewChannelTimeout(256),
		logger:        logger,
	}
	go rpc.handleIncomingMsgs()

	rpc.logger.Infof("Replicated-Catalog creation done")
	return rpc, nil
}

func (rpc *replicatedCatalog) Register(si *ServiceInstance) (*ServiceInstance, error) {
	result, err := rpc.local.Register(si)
	if err != nil {
		return result, err
	}

	payload, _ := json.Marshal(result)
	msg, err := json.Marshal(&replicatedMsg{RepType: REGISTER, Payload: payload})
	if err != nil {
		rpc.logger.WithFields(log.Fields{
			"error": err,
		}).Errorf("Failed to marshal REGISTER message for replication. instance: %v", si)
	} else {
		if err = rpc.replicator.Broadcast(msg); err != nil {
			rpc.logger.WithFields(log.Fields{
				"error": err,
			}).Errorf("Failed to broadcast REGISTER message for replication. instance: %v", si)
		}
	}

	return result, nil
}

func (rpc *replicatedCatalog) Deregister(instanceID string) (*ServiceInstance, error) {
	instance, err := rpc.local.Deregister(instanceID)
	if err != nil {
		return nil, err
	}

	msg, err := json.Marshal(&replicatedMsg{RepType: DEREGISTER, Payload: []byte(instanceID)})
	if err != nil {
		rpc.logger.WithFields(log.Fields{
			"error": err,
		}).Errorf("Failed to marshal DEREGISTER message for replication. instanceID: %s", instanceID)
	} else {
		if err = rpc.replicator.Broadcast(msg); err != nil {
			rpc.logger.WithFields(log.Fields{
				"error": err,
			}).Errorf("Failed to broadcast DEREGISTER message for replication. instanceID: %s", instanceID)
		}
	}

	return instance, nil
}

func (rpc *replicatedCatalog) Renew(instanceID string) (*ServiceInstance, error) {
	instance, err := rpc.local.Renew(instanceID)
	if err != nil {
		return nil, err
	}

	msg, err := json.Marshal(&replicatedMsg{RepType: RENEW, Payload: []byte(instanceID)})
	if err != nil {
		rpc.logger.WithFields(log.Fields{
			"error": err,
		}).Errorf("Failed to marshal RENEW message for replication. instanceID: %s", instanceID)
	} else {
		if err := rpc.replicator.Broadcast(msg); err != nil {
			rpc.logger.WithFields(log.Fields{
				"error": err,
			}).Errorf("Failed to broadcast RENEW message for replication. instanceID: %s", instanceID)
		}
	}

	return instance, nil
}

func (rpc *replicatedCatalog) SetStatus(instanceID, status string) (*ServiceInstance, error) {
	instance, err := rpc.local.SetStatus(instanceID, status)
	if err != nil {
		return nil, err
	}

	payload, _ := json.Marshal(&replicatedStatus{instanceID, status})
	msg, err := json.Marshal(&replicatedMsg{RepType: SETSTATUS, Payload: payload})
	if err != nil {
		rpc.logger.WithFields(log.Fields{
			"error": err,
		}).Errorf("Failed to marshal SETSTATUS message for replication. instanceID: %s, status: %s", instanceID, status)
	} else {
		if err := rpc.replicator.Broadcast(msg); err != nil {
			rpc.logger.WithFields(log.Fields{
				"error": err,
			}).Errorf("Failed to broadcast SETSTATUS message for replication. instanceID: %s, status: %s", instanceID, status)
		}
	}

	return instance, nil
}

func (rpc *replicatedCatalog) Instance(instanceID string) (*ServiceInstance, error) {
	return rpc.local.Instance(instanceID)
}

func (rpc *replicatedCatalog) List(serviceName string, predicate Predicate) ([]*ServiceInstance, error) {
	return rpc.local.List(serviceName, predicate)
}

func (rpc *replicatedCatalog) ListServices(predicate Predicate) []*Service {
	return rpc.local.ListServices(predicate)
}

func (rpc *replicatedCatalog) handleIncomingMsgs() {
	var data replicatedMsg

	for msg := range rpc.notifyChannel.Channel() {
		inMsg := msg.(*replication.InMessage)
		err := json.Unmarshal(inMsg.Data, &data)
		if err != nil {
			rpc.logger.WithFields(log.Fields{
				"error": err,
			}).Errorf("Failed to unmarshal replication message. data: %v", &data)
			continue
		}

		switch data.RepType {
		case REGISTER:
			var si ServiceInstance
			if err = json.Unmarshal(data.Payload, &si); err != nil {
				rpc.logger.WithFields(log.Fields{
					"error": err,
				}).Errorf("Failed to unmarshal register replicated instance. data: %s", string(data.Payload))
				break
			}
			_, err = rpc.local.Register(&si)
			if err != nil {
				rpc.logger.WithFields(log.Fields{
					"error": err,
				}).Errorf("Failed to register replicated instance. instance: %v", &si)
			}
			break
		case DEREGISTER:
			instanceID := string(data.Payload)
			_, err := rpc.local.Deregister(instanceID)
			if err != nil {
				rpc.logger.WithFields(log.Fields{
					"error": err,
				}).Errorf("Failed to deregister replicated instance. instanceID: %s", instanceID)
			}
			break
		case RENEW:
			instanceID := string(data.Payload)
			_, err := rpc.local.Renew(instanceID)
			if err != nil {
				msg, err := json.Marshal(&replicatedMsg{RepType: READREPAIR, Payload: data.Payload})
				if err != nil {
					rpc.logger.WithFields(log.Fields{
						"error": err,
					}).Errorf("Failed to marshal READREPAIR message for replication. instanceID: %s", instanceID)
					break
				}
				rpc.replicator.Send(inMsg.MemberID, msg)
			}
			break
		case SETSTATUS:
			var repStatus replicatedStatus
			if err = json.Unmarshal(data.Payload, &repStatus); err != nil {
				rpc.logger.WithFields(log.Fields{
					"error": err,
				}).Errorf("Failed to unmarshal replicated instance status. data: %s", string(data.Payload))
				break
			}
			_, err = rpc.local.SetStatus(repStatus.InstanceID, repStatus.Status)
			if err != nil {
				rpc.logger.WithFields(log.Fields{
					"error": err,
				}).Errorf("Failed to replicate instance status. instanceID: %s, status: %s", repStatus.InstanceID, repStatus.Status)
			}
			break
		case READREPAIR:
			instanceID := string(data.Payload)
			result, err := rpc.local.Instance(instanceID)
			if err != nil {
				rpc.logger.WithFields(log.Fields{
					"error": err,
				}).Errorf("Failed to find instance for READREPAIR request. instanceID: %s", instanceID)
				break
			}
			payload, _ := json.Marshal(result)
			msg, err := json.Marshal(&replicatedMsg{RepType: REGISTER, Payload: payload})
			if err != nil {
				rpc.logger.WithFields(log.Fields{
					"error": err,
				}).Errorf("Failed to marshal REGSITER message for replication. data: %v", result)
			}
			rpc.replicator.Send(inMsg.MemberID, msg)
			break
		}
	}
}
