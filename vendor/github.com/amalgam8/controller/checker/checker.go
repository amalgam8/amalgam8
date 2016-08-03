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
	"reflect"
	"sort"
	"time"

	"net/http"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/controller/nginx"
	"github.com/amalgam8/controller/notification"
	"github.com/amalgam8/controller/resources"
	"github.com/amalgam8/registry/client"
)

// Checker client
type Checker interface {
	Check(ids []string) error
}

type checker struct {
	db            database.Tenant
	producerCache notification.TenantProducerCache
	generator     nginx.Generator
	factory       RegistryFactory
}

// Config options
type Config struct {
	Database      database.Tenant
	ProducerCache notification.TenantProducerCache
	Generator     nginx.Generator
	Factory       RegistryFactory
}

// New instantiates new instance
func New(conf Config) Checker {
	return &checker{
		db:            conf.Database,
		producerCache: conf.ProducerCache,
		generator:     conf.Generator,
		factory:       conf.Factory,
	}
}

// Check registered tenants for Registry catalog changes. Each registered tenant's catalog is retrieved from
// the database and compared against the current Registry catalog for that tenant. If a difference exists
// between the two, the database is updated with the most recent version of the catalog and any listeners are notified
// of the change.
//
// ids must be a subset of registered tenant IDs or nil or empty. If ids is nil or empty all registered IDs are checked.
//
// TODO: make this asynch
func (c *checker) Check(ids []string) error {
	// Get registered tenant catalogs
	entries, err := c.db.List(ids)
	if err != nil {
		// log failure
		return err
	}

	for _, entry := range entries {
		creds := entry.ProxyConfig.Credentials

		// Get newest catalog from Registry
		latestCatalog, err := c.getLatestCatalog(creds.Registry)
		if err != nil {
			logrus.WithError(err).Warn("Could not get latest catalog from registry")

			continue
		}

		// Check for differences
		if !c.catalogsEqual(entry.ServiceCatalog, latestCatalog) {
			// Update database
			entry.ServiceCatalog.Services = latestCatalog.Services
			entry.ServiceCatalog.LastUpdate = time.Now()

			if err = c.updateCatalog(entry); err != nil {
				// error during update, do not send event
				continue
			}

			// Notify tenants
			templ := c.generator.TemplateConfig(entry.ServiceCatalog, entry.ProxyConfig)
			if err = c.producerCache.SendEvent(entry.TenantToken, creds.Kafka, templ); err != nil {
				logrus.WithFields(logrus.Fields{
					"err":       err,
					"tenant_id": entry.ID,
				}).Error("Failed to notify tenant of rules change")
			}

		}
	}

	return nil
}

// catalogsEqual
func (c *checker) catalogsEqual(a, b resources.ServiceCatalog) bool {
	equal := reflect.DeepEqual(a.Services, b.Services)
	logrus.WithFields(logrus.Fields{
		"a":     a,
		"b":     b,
		"equal": equal,
	}).Debug("Comparing catalogs")
	return equal
}

// getLatestCatalog
func (c *checker) getLatestCatalog(sd resources.Registry) (resources.ServiceCatalog, error) {
	catalog := resources.ServiceCatalog{}

	registry, err := c.factory.NewRegistryClient(sd.Token, sd.URL)
	if err != nil {
		logrus.WithError(err).Error("Failed to initialize registry client")
		return catalog, err
	}

	instances, err := registry.ListInstances(client.InstanceFilter{})
	if err != nil {
		return catalog, err
	}

	// Convert
	serviceMap := make(map[string]*resources.Service)
	for _, instance := range instances {
		if serviceMap[instance.ServiceName] == nil {
			serviceMap[instance.ServiceName] = &resources.Service{
				Name:      instance.ServiceName,
				Endpoints: []resources.Endpoint{},
			}
		}

		metadata := map[string]string{}
		err = json.Unmarshal(instance.Metadata, &metadata)

		endpoint := resources.Endpoint{
			Type:     instance.Endpoint.Type,
			Value:    instance.Endpoint.Value,
			Metadata: resources.MetaData{Version: metadata["version"]},
		}

		serviceMap[instance.ServiceName].Endpoints = append(serviceMap[instance.ServiceName].Endpoints, endpoint)
	}

	for _, service := range serviceMap {
		catalog.Services = append(catalog.Services, *service)
	}

	// Sort for comparisons since registry does not guarantee any ordering
	sort.Sort(resources.ByService(catalog.Services))
	for _, service := range catalog.Services {
		sort.Sort(resources.ByEndpoint(service.Endpoints))
	}

	return catalog, nil
}

// updateCatalog updates the stored catalog. If there is a database conflict (I.E., the entry has changed between being
// read and being written) we attempt to re-read the entry and update the catalog.
func (c *checker) updateCatalog(entry resources.TenantEntry) error {
	var err error
	if err = c.db.Update(entry); err != nil {
		if ce, ok := err.(*database.DBError); ok {
			if ce.StatusCode == http.StatusConflict {
				newerEntry, err := c.db.Read(entry.ID)
				if err == nil {
					newerEntry.ServiceCatalog.Services = entry.ServiceCatalog.Services
					newerEntry.ServiceCatalog.LastUpdate = entry.ServiceCatalog.LastUpdate
					if err = c.db.Update(newerEntry); err != nil {
						logrus.WithFields(logrus.Fields{
							"err": err,
							"id":  entry.ID,
						}).Error("Failed to resolve document update conflict")
						return err
					}
					logrus.WithFields(logrus.Fields{
						"id": entry.ID,
					}).Debug("Succesfully resolved document update conflict")
					return nil

				}
				logrus.WithFields(logrus.Fields{
					"err": err,
					"id":  entry.ID,
				}).Error("Failed to retrieve latest document during conflict resolution")
				return err

			}
			logrus.WithFields(logrus.Fields{
				"err": err,
				"id":  entry.ID,
			}).Error("Database error attempting to update service catalog")
			return err
		}
		logrus.WithFields(logrus.Fields{
			"err": err,
			"id":  entry.ID,
		}).Error("Failed attempting to update service catalog")
		return err
	}

	return nil
}
