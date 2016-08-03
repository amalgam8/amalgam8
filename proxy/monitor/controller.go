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
	"github.com/amalgam8/sidecar/proxy/clients"
)

// ControllerListener is notified of changes to controller
type ControllerListener interface {
	RuleChange(proxyConfig resources.ProxyConfig) error
}

// ControllerConfig options
type ControllerConfig struct {
	Client       clients.Controller
	Listeners    []ControllerListener
	PollInterval time.Duration
}

type controller struct {
	ticker       *time.Ticker
	controller   clients.Controller
	pollInterval time.Duration
	version      *time.Time
	listeners    []ControllerListener
}

// NewController instantiates a new instance
func NewController(conf ControllerConfig) Monitor {
	return &controller{
		controller:   conf.Client,
		listeners:    conf.Listeners,
		pollInterval: conf.PollInterval,
	}
}

// Start monitoring the A8 controller. This is a blocking operation.
func (c *controller) Start() error {
	// Stop existing ticker if necessary
	if c.ticker != nil {
		if err := c.Stop(); err != nil {
			logrus.WithError(err).Error("Could not stop existing periodic poll")
			return err
		}
	}

	// Create new ticker
	c.ticker = time.NewTicker(c.pollInterval)

	// Do initial poll
	if err := c.poll(); err != nil {
		logrus.WithError(err).Error("Poll failed")
	}

	// Start periodic poll
	for range c.ticker.C {
		if err := c.poll(); err != nil {
			logrus.WithError(err).Error("Poll failed")
		}
	}

	return nil
}

// poll the A8 controller for changes and notify listeners
func (c *controller) poll() error {

	// Get latest config from the A8 controller
	conf, err := c.controller.GetProxyConfig(c.version)
	if err != nil {
		logrus.WithError(err).Error("Call to controller failed")
		return err
	}

	if conf == nil { // Nothing to update
		return nil
	}

	// Notify listeners
	for _, listener := range c.listeners {
		if err := listener.RuleChange(*conf); err != nil {
			logrus.WithError(err).Warn("Controller listener failed")
		}
	}

	// TODO: either time should be obtained from the controller OR time should only be obtained locally, otherwise
	// there may be issues with clock synchronization.
	t := time.Now()
	c.version = &t

	return nil
}

// Stop monitoring the A8 controller
func (c *controller) Stop() error {
	// Stop ticker if necessary
	if c.ticker != nil {
		c.ticker.Stop()
		c.ticker = nil
	}

	return nil
}
