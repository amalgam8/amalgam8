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
	"github.com/amalgam8/amalgam8/pkg/api"
)

// RulesListener is notified of changes to rules
type RulesListener interface {
	RuleChange([]api.Rule) error
}

// RulesConfig holds configuration options for the rules monitor.
type RulesConfig struct {
	Rules        api.RulesService
	Listeners    []RulesListener
	PollInterval time.Duration
}

type rulesMonitor struct {
	rules api.RulesService

	ticker       *time.Ticker
	pollInterval time.Duration

	revision  int64
	listeners []RulesListener
}

// NewRulesMonitor instantiates a new rules monitor
func NewRulesMonitor(conf RulesConfig) Monitor {
	return &rulesMonitor{
		rules:        conf.Rules,
		listeners:    conf.Listeners,
		pollInterval: conf.PollInterval,
		revision:     -1,
	}
}

// Start monitoring the rules service. This is a blocking operation.
func (c *rulesMonitor) Start() error {
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

// poll the rules service for changes and notify listeners
func (c *rulesMonitor) poll() error {

	// Get the latest rules from the A8 controller.
	rulesset, err := c.rules.ListRules(&api.RuleFilter{})
	if err != nil {
		logrus.WithError(err).Error("Call to rules service failed")
		return err
	}

	// Short-circuit if the controller's revision is not newer than our revision
	if c.revision >= rulesset.Revision {
		return nil
	}

	// Update our revision
	c.revision = rulesset.Revision

	// Notify listeners
	for _, listener := range c.listeners {
		if err := listener.RuleChange(rulesset.Rules); err != nil {
			logrus.WithError(err).Warn("Rules listener failed")
		}
	}

	return nil
}

// Stop monitoring the rules service
func (c *rulesMonitor) Stop() error {
	// Stop ticker if necessary
	if c.ticker != nil {
		c.ticker.Stop()
		c.ticker = nil
	}

	return nil
}
