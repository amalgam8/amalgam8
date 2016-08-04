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
	"github.com/amalgam8/controller/client"
	"github.com/amalgam8/controller/rules"
)

// ControllerListener is notified of changes to controller
type ControllerListener interface {
	RuleChange([]rules.Rule) error
}

// ControllerConfig options
type ControllerConfig struct {
	Client       client.Client
	Listeners    []ControllerListener
	PollInterval time.Duration
}

type controller struct {
	ticker         *time.Ticker
	controller     client.Client
	pollInterval   time.Duration
	currentVersion *time.Time
	listeners      []ControllerListener
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

	// Get the latest rules from the A8 controller.
	resp, err := c.controller.GetRules(rules.Filter{})
	if err != nil {
		logrus.WithError(err).Error("Call to controller failed")
		return err
	}

	// Check if the rules have been modified since the last poll
	if c.currentVersion != nil && !resp.LastUpdated.After(*c.currentVersion) {
		return nil
	}
	c.currentVersion = &resp.LastUpdated

	// Notify listeners
	for _, listener := range c.listeners {
		if err := listener.RuleChange(resp.Rules); err != nil {
			logrus.WithError(err).Warn("Controller listener failed")
		}
	}

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
