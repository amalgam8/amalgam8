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
	"fmt"
	"os"
)

const (
	tmpNginxConfigPath = "/tmp/nginx.conf"
	nginxConfigPath    = "/etc/nginx/nginx.conf"
	nginxConfigBackup  = "/etc/nginx/nginx.conf.bak"
)

// Config handles updates to the NGINX config
type Config interface {
	Update(config string) error
	Revert() error
}

type config struct {
}

// NewConfig creates a new config
func NewConfig() Config {
	return &config{}
}

// Update updates the NGINX configuration file
func (n *config) Update(config string) error {

	output, err := os.OpenFile(tmpNginxConfigPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.WithField("err", err).Error("Couldn't open NGINX config file")
		return err
	}

	// Write the config
	fmt.Fprintf(output, config)
	output.Close()

	// Only backup if there is an existing config
	if _, err := os.Stat(nginxConfigPath); !os.IsNotExist(err) {
		log.Info("Backing up existing NGINX config")

		// Attempt to save a backup of NGINX config. If we can't save a backup,
		// we do not continue.
		if err = os.Rename(nginxConfigPath, nginxConfigBackup); err != nil {
			log.WithField("err", err).Error("Could not create backup of NGINX config")
			return err
		}
	}

	// Move the new NGINX config into place. On failure, attempt to restore
	// the backup of the NGINX config
	if err = os.Rename(tmpNginxConfigPath, nginxConfigPath); err != nil {
		log.WithField("err", err).Error("Could not overwrite existing NGINX config, restoring original")
		// Attempt to restore the backup of the config
		if err2 := os.Rename(nginxConfigBackup, nginxConfigPath); err2 != nil {
			log.WithField("err", err2).Error("Could not restore original NGINX config")
			return err2
		}

		return err
	}

	return nil
}

// Revert reverts NGINX to the backup of the original configuration
func (n *config) Revert() error {
	// Only backup if there is an existing config
	if _, err := os.Stat(nginxConfigBackup); os.IsNotExist(err) {
		log.Error("No existing NGINX config to revert to")
		return err
	}

	// Attempt to restore the backup of the config.  .bak overwrites existing
	// TODO: ideally, do a copy of the backup instead of rename so that this function is idempotent
	if err := os.Rename(nginxConfigBackup, nginxConfigPath); err != nil {
		log.WithField("err", err).Error("Could not restore original NGINX config")
		return err
	}

	return nil
}
