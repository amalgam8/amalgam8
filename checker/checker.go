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

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/clients"
	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/controller/nginx"
	"github.com/amalgam8/controller/notification"
	"github.com/amalgam8/controller/resources"
)

// Checker client
type Checker interface {
	Check(ids []string) error
}

type checker struct {
	db            database.Tenant
	registry      clients.Registry
	producerCache notification.TenantProducerCache
	generator     nginx.Generator
}

// Config options
type Config struct {
	Database      database.Tenant
	Registry      clients.Registry
	ProducerCache notification.TenantProducerCache
	Generator     nginx.Generator
}

// New instantiates new instance
func New(conf Config) Checker {
	return &checker{
		db:            conf.Database,
		registry:      conf.Registry,
		producerCache: conf.ProducerCache,
		generator:     conf.Generator,
	}
}

// TODO: make this asynch
// TODO: If ids == nil or len(ids) == 0, assume we want to check all IDs
// Check registered tenants for Registry catalog changes. Each registered tenant's catalog is retrieved from
// the database and compared against the current Registry catalog for that tenant. If a difference exists
// between the two, the database is updated with the most recent version of the catalog and any listeners are notified
// of the change.
//
// ids must be a subset of registered tenant IDs or nil or empty. If ids is nil or empty all registered IDs are checked.
func (c *checker) Check(ids []string) error {
	// Get registered tenant catalogs
	//catalogs, err := c.getStoredCatalogs(ids)
	entries, err := c.db.List(ids)
	if err != nil {
		// log failure
		return err
	}

	for _, entry := range entries {

		// Get Registry credentials from auth
		creds := entry.ProxyConfig.Credentials

		// Get newest catalog from Registry
		latestCatalog, err := c.getLatestCatalog(creds.Registry)
		if err != nil {
			// log failure

			//TODO get new token from auth and try again on 401's
			logrus.WithError(err).Warn("Could not get latest from Registry")

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

			// Notify tenant
			templ := c.generator.TemplateConfig(entry.ServiceCatalog, entry.ProxyConfig)
			if err = c.producerCache.SendEvent(entry.TenantToken, creds.Kafka, templ); err != nil {
				logrus.WithFields(logrus.Fields{
					"err":       err,
					"tenant_id": entry.ID,
					// TODO request ID logging??
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

	instances, err := c.registry.GetInstances(sd.Token, sd.URL)
	if err != nil {
		// log err
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

		endpoint := resources.Endpoint{
			Type:     instance.Endpoint.Type,
			Value:    instance.Endpoint.Value,
			Metadata: resources.MetaData{Version: instance.MetaData.Version},
		}

		serviceMap[instance.ServiceName].Endpoints = append(serviceMap[instance.ServiceName].Endpoints, endpoint)
	}

	for _, service := range serviceMap {
		catalog.Services = append(catalog.Services, *service)
	}

	// Sort
	sort.Sort(resources.ByService(catalog.Services))
	for _, service := range catalog.Services {
		sort.Sort(resources.ByEndpoint(service.Endpoints))
	}

	return catalog, nil
}

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
