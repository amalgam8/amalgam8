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

package nginx

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/sidecar/router/clients"
)

var log = logrus.WithFields(logrus.Fields{
	"module": "tenant.nginx",
})

// Nginx manages updates to NGINX configuration
type Nginx interface {
	// Update NGINX to run with the provided configuration
	Update(templateConf clients.NGINXJson) error
}

type nginx struct {
	service     Service
	serviceName string
	mutex       sync.Mutex
	nginxClient clients.NGINX
}

// Conf for creating new NGINX interface
type Conf struct {
	Service     Service
	NGINXClient clients.NGINX
}

// NewNginx creates new Nginx instance
func NewNginx(conf Conf) (Nginx, error) {
	return &nginx{
		service:     conf.Service,
		nginxClient: conf.NGINXClient,
	}, nil
}

// Update NGINX to run with the provided configuration
func (n *nginx) Update(templateConf clients.NGINXJson) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	var err error

	// Determine if NGINX is running
	running, err := n.service.Running()
	if err != nil {
		log.WithError(err).Error("Could not get status of NGINX service")
		return err
	}

	if !running {
		// NGINX is not running; attempt to start NGINX
		if err := n.startNginx(); err != nil {
			log.WithError(err).Error("Failed to start NGINX")
			return err
		}
	}

	if err = n.nginxClient.UpdateHTTPUpstreams(templateConf); err != nil {
		logrus.WithError(err).Error("Failed to update HTTP upstreams with NGINX")
		return err
	}

	return nil
}

// startNginx attempts to start the NGINX service. On a failure attempt to start NGINX with the backup configuration.
func (n *nginx) startNginx() error {
	log.Debug("Starting NGINX")
	if err := n.service.Start(); err != nil {
		log.WithField("err", err).Error("Could not start NGINX service")
		return err
	}

	return nil
}
