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

// RegistrationConfig options
type RegistrationConfig struct {
	Client          client.Client
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
		registeredInstance, err := agent.config.Client.Register(agent.config.ServiceInstance)
		if err == nil {
			go agent.renew(registeredInstance)
			return
		}
		logrus.WithError(err).WithField("service_name", agent.config.ServiceInstance.ServiceName).Warn("Registration failed")

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
			err := agent.config.Client.Renew(instance.ID)
			if cErr, ok := err.(client.Error); ok && cErr.Code == client.ErrorCodeUnknownInstance {
				logrus.WithError(cErr).WithField("service_name", instance.ServiceName).Warn("Heartbeat failed")
				go agent.register()
				return
			}
		case <-agent.stop:
			agent.deregister(instance)
			agent.stop <- struct{}{}
			return
		}
	}
}

func (agent *RegistrationAgent) deregister(instance *client.ServiceInstance) {
	logrus.WithField("service_name", instance.ServiceName).Info("Deregistered")
	agent.config.Client.Deregister(instance.ID)
}
