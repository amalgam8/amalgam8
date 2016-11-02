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

package register

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/registry/client"
)

// DefaultHeartbeatsPerTTL default number of heartbeats per TTL
const DefaultHeartbeatsPerTTL = 3

// DefaultReregistrationDelay default delay before registering
const DefaultReregistrationDelay = time.Duration(5) * time.Second

// Lifecycle is the interface implemented by objects that can be started and stopped.
type Lifecycle interface {
	Start()
	Stop()
}

// RegistrationConfig options
type RegistrationConfig struct {
	Client          client.Registry
	ServiceInstance *client.ServiceInstance
}

// RegistrationAgent maintains a registration with registry.
type RegistrationAgent struct {
	config RegistrationConfig
	active bool
	stop   chan struct{}
	mutex  sync.Mutex
}

// NewRegistrationAgent instantiates a new instance of the agent
func NewRegistrationAgent(config RegistrationConfig) (*RegistrationAgent, error) {
	// TODO validate config

	agent := &RegistrationAgent{
		config: config,
		stop:   make(chan struct{}),
	}

	return agent, nil
}

// Start maintaining registration with registry.
// Non-blocking.
func (agent *RegistrationAgent) Start() {
	logrus.WithField("service_name", agent.config.ServiceInstance.ServiceName).
		Info("Starting Amalgam8 service registration agent")

	agent.mutex.Lock()
	defer agent.mutex.Unlock()

	if agent.active {
		return
	}
	agent.active = true

	go agent.register()
}

// Stop maintaining registration with registry.
// Blocks until deregistration attempt is complete.
func (agent *RegistrationAgent) Stop() {
	logrus.WithField("service_name", agent.config.ServiceInstance.ServiceName).
		Info("Stopping Amalgam8 service registration agent")

	agent.mutex.Lock()
	defer agent.mutex.Unlock()

	if !agent.active {
		return
	}
	agent.active = false

	agent.stop <- struct{}{}
	<-agent.stop
}

func (agent *RegistrationAgent) register() {
	for {
		logrus.WithField("service_name", agent.config.ServiceInstance.ServiceName).
			Debug("Attempting to register service with Amalgam8")

		registeredInstance, err := agent.config.Client.Register(agent.config.ServiceInstance)
		if err == nil {
			logrus.WithFields(logrus.Fields{
				"service_name": registeredInstance.ServiceName,
				"instance_id":  registeredInstance.ID,
			}).Info("Service successfully registered with Amalgam8")

			go agent.renew(registeredInstance)
			return
		}

		logrus.WithError(err).WithField("service_name", agent.config.ServiceInstance.ServiceName).
			Warnf("Service registration had failed. Re-attempting in %s", DefaultReregistrationDelay)

		select {
		case <-time.After(DefaultReregistrationDelay):
			continue
		case <-agent.stop:
			agent.stop <- struct{}{}
			return
		}
	}
}

func (agent *RegistrationAgent) renew(instance *client.ServiceInstance) {
	interval := time.Duration(instance.TTL) * time.Second / DefaultHeartbeatsPerTTL

	for {
		select {
		case <-time.After(interval):
			logrus.WithFields(logrus.Fields{
				"service_name": instance.ServiceName,
				"instance_id":  instance.ID,
			}).Debug("Attempting to renew service registration with Amalgam8")

			err := agent.config.Client.Renew(instance.ID)
			if err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"service_name": instance.ServiceName,
					"instance_id":  instance.ID,
				}).Warn("Service registration renewal had failed")

				if cErr, ok := err.(client.Error); ok && cErr.Code == client.ErrorCodeUnknownInstance {
					go agent.register()
					return
				}
			}

		case <-agent.stop:
			agent.deregister(instance)
			agent.stop <- struct{}{}
			return
		}
	}
}

func (agent *RegistrationAgent) deregister(instance *client.ServiceInstance) {
	logrus.WithFields(logrus.Fields{
		"service_name": instance.ServiceName,
		"instance_id":  instance.ID,
	}).Info("Attempting to deregister service with Amalgam8")

	err := agent.config.Client.Deregister(instance.ID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"service_name": instance.ServiceName,
			"instance_id":  instance.ID,
		}).Warn("Service deregistration had failed")
	} else {
		logrus.WithFields(logrus.Fields{
			"service_name": instance.ServiceName,
			"instance_id":  instance.ID,
		}).Info("Service successfully deregistered with Amalgam8")
	}
}
