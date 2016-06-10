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
	"io"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
)

var log = logrus.WithFields(logrus.Fields{
	"module": "tenant.nginx",
})

// Nginx manages updates to NGINX configuration
type Nginx interface {
	// Update NGINX to run with the provided configuration
	Update(reader io.Reader) error
}

type nginx struct {
	config      Config
	service     Service
	serviceName string
	mutex       sync.Mutex
}

// NewNginx creates new Nginx instance
func NewNginx(serviceName string) Nginx {
	return &nginx{
		config:      NewConfig(),
		service:     NewService(),
		serviceName: serviceName,
	}
}

// Update NGINX to run with the provided configuration
func (n *nginx) Update(reader io.Reader) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	configBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		log.WithField("err", err).Error("Could not read config")
		return err
	}

	// Replace service name
	configStr := string(configBytes)
	configStr = strings.Replace(configStr, "__SERVICE_NAME__", n.serviceName, -1)

	// Update the NGINX configuration file
	if err = n.config.Update(configStr); err != nil {
		log.WithField("err", err).Error("Could not update NGINX configuration file")
		return err
	}

	// Determine if NGINX is running
	running, err := n.service.Running()
	if err != nil {
		log.WithField("err", err).Error("Could not get status of NGINX service")
		return err
	}

	var nginxErr error
	if running {
		// NGINX is already running; attempt to reload NGINX
		nginxErr = n.reloadNginx()
	} else {
		// NGINX is not running; attempt to start NGINX
		nginxErr = n.startNginx()
	}

	// log the failed nginx config
	if nginxErr != nil {
		log.WithField("config", string(configBytes)).Error("Failed NGINX config")
		return nginxErr
	}

	return nil
}

// startNginx attempts to start the NGINX service. On a failure attempt to start NGINX with the backup configuration.
func (n *nginx) startNginx() error {
	log.Debug("Starting NGINX with new configuration")
	if err := n.service.Start(); err != nil {
		log.WithField("err", err).Error("Could not start NGINX service with new configuration")
		if revertErr := n.config.Revert(); revertErr != nil {
			log.WithError(revertErr).Error("Reverting to backup NGINX configuration failed")
			return revertErr
		}

		if startErr := n.service.Start(); startErr != nil {
			log.WithField("err", startErr).Error("Could not start NGINX with backup configuration")
			return startErr
		}

		log.Warn("Reverted to old NGINX configuration")
		return err
	}

	return nil
}

// reloadNginx attempts to reload the NGINX service. On failure revert to the backup NGINX configuration.
func (n *nginx) reloadNginx() error {
	log.Debug("Reloading NGINX with new configuration")
	if err := n.service.Reload(); err != nil {
		log.WithField("err", err).Error("Could not reload NGINX with new configuration")
		if revertErr := n.config.Revert(); revertErr != nil {
			log.WithField("err", revertErr).Error("Failed to revert NGINX configuration to backup")
			return revertErr
		}

		log.Warn("Reverted to old NGINX configuration")
		return err
	}

	return nil
}
