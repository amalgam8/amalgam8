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

package checker

import (
	"time"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/sidecar/config"
	"github.com/amalgam8/sidecar/router/clients"
	"github.com/amalgam8/sidecar/router/nginx"
)

// Poller performs a periodic poll on Controller for changes to the NGINX config
type Poller interface {
	Start() error
	Stop() error
}

type poller struct {
	ticker     *time.Ticker
	controller clients.Controller
	nginx      nginx.Nginx
	config     *config.Config
	version    *time.Time
}

// NewPoller creates instance
func NewPoller(config *config.Config, rc clients.Controller, nginx nginx.Nginx) Poller {
	return &poller{
		controller: rc,
		config:     config,
		nginx:      nginx,
	}
}

// Start begins periodic polling of Controller for the latest configuration. This is a blocking operation.
func (p *poller) Start() error {
	// Stop existing ticker if necessary
	if p.ticker != nil {
		if err := p.Stop(); err != nil {
			logrus.WithError(err).Error("Could not stop existing periodic poll")
			return err
		}
	}

	// Create new ticker
	p.ticker = time.NewTicker(p.config.Controller.Poll)

	// Do initial poll
	if err := p.poll(); err != nil {
		logrus.WithError(err).Error("Poll failed")
	}

	// Start periodic poll
	for _ = range p.ticker.C {
		if err := p.poll(); err != nil {
			logrus.WithError(err).Error("Poll failed")
		}
	}

	return nil
}

// poll obtains the latest NGINX config from Controller and updates NGINX to use it
func (p *poller) poll() error {

	// Get latest config from Controller
	conf, err := p.controller.GetNGINXConfig(p.version)
	if err != nil {
		logrus.WithError(err).Error("Call to Controller failed")
		return err
	}

	if conf == nil {
		//TODO no new rules to update, do we need to do anything else?
		return nil
	}

	confBytes, err := json.Marshal(conf)

	// Update our existing NGINX config
	if err := p.nginx.Update(confBytes); err != nil {
		logrus.WithError(err).Error("Could not update NGINX config")
		return err
	}

	t := time.Now()
	p.version = &t

	return nil
}

// Stop halts the periodic poll of Controller
func (p *poller) Stop() error {
	// Stop ticker if necessary
	if p.ticker != nil {
		p.ticker.Stop()
		p.ticker = nil
	}

	return nil
}
