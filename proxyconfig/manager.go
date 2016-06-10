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

package proxyconfig

import (
	"net/http"

	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/controller/notification"
	"github.com/amalgam8/controller/resources"
)

// Manager client
type Manager interface {
	Set(rules resources.ProxyConfig) error
	Get(id string) (resources.ProxyConfig, error)
	Delete(id string) error
}

type manager struct {
	db            database.Rules
	producerCache notification.TenantProducerCache
}

// Config options
type Config struct {
	Database      database.Rules
	ProducerCache notification.TenantProducerCache
}

// NewManager creates Manager instance
func NewManager(conf Config) Manager {
	return &manager{
		db:            conf.Database,
		producerCache: conf.ProducerCache,
	}
}

// Set database entry
func (p *manager) Set(rules resources.ProxyConfig) error {
	var err error
	if err := p.validate(rules); err != nil {
		return err
	}

	if rules.Rev == "" {
		err = p.db.Create(rules)
	} else {
		err = p.db.Update(rules)
	}

	if err != nil {
		if ce, ok := err.(*database.DBError); ok {
			if ce.StatusCode == http.StatusConflict {
				// There is an old orphan entry in the database, delete it and create a new entry
				oldRules, err := p.db.Read(rules.ID)
				if err != nil {
					return err
				}

				rules.Rev = oldRules.Rev

				if err = p.db.Update(rules); err != nil {
					return err
				}
			} else {
				return err
			}

		} else {
			return err
		}
	}

	// Send Kafka event
	if err = p.producerCache.SendEvent(rules.ID, rules.Credentials.Kafka); err != nil {
		return err
	}

	return nil
}

// Get database entry
func (p *manager) Get(id string) (resources.ProxyConfig, error) {
	return p.db.Read(id)
}

// Delete database entry
func (p *manager) Delete(id string) error {
	return p.db.Delete(id)
}

func (p *manager) validate(config resources.ProxyConfig) error {
	return nil
}
