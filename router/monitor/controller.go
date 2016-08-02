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

package monitor

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/resources"
	"github.com/amalgam8/sidecar/router/clients"
)

type ControllerListener interface {
	RuleChange(proxyConfig resources.ProxyConfig) error
}

type ControllerConfig struct {
	Client       clients.Controller
	Listener     ControllerListener
	PollInterval time.Duration
}

type controller struct {
	ticker       *time.Ticker
	controller   clients.Controller
	pollInterval time.Duration
	version      *time.Time
	listener     ControllerListener
}

func NewController(conf ControllerConfig) Monitor {
	return &controller{
		controller:   conf.Client,
		listener:     conf.Listener,
		pollInterval: conf.PollInterval,
	}
}

// Start begins periodic polling of Controller for the latest configuration. This is a blocking operation.
func (p *controller) Start() error {
	// Stop existing ticker if necessary
	if p.ticker != nil {
		if err := p.Stop(); err != nil {
			logrus.WithError(err).Error("Could not stop existing periodic poll")
			return err
		}
	}

	// Create new ticker
	p.ticker = time.NewTicker(p.pollInterval)

	// Do initial poll
	if err := p.poll(); err != nil {
		logrus.WithError(err).Error("Poll failed")
	}

	// Start periodic poll
	for range p.ticker.C {
		if err := p.poll(); err != nil {
			logrus.WithError(err).Error("Poll failed")
		}
	}

	return nil
}

// poll obtains the latest NGINX config from Controller and updates NGINX to use it
func (p *controller) poll() error {

	// Get latest config from Controller
	conf, err := p.controller.GetProxyConfig(p.version)
	if err != nil {
		logrus.WithError(err).Error("Call to Controller failed")
		return err
	}

	if conf == nil { // Nothing to update
		return nil
	}

	// Notify listeners of change
	if err := p.listener.RuleChange(*conf); err != nil {
		logrus.WithError(err).Error("Listener failed")
		return err
	}

	// TODO: either time should be obtained from the controller OR time should only be obtained locally, otherwise
	// there may be issues with clock synchronization.
	t := time.Now()
	p.version = &t

	return nil
}

// Stop halts the periodic poll of Controller
func (p *controller) Stop() error {
	// Stop ticker if necessary
	if p.ticker != nil {
		p.ticker.Stop()
		p.ticker = nil
	}

	return nil
}
