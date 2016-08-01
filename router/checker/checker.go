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

	"encoding/json"

	"time"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/resources"
	"github.com/amalgam8/registry/client"
	"github.com/amalgam8/sidecar/config"
)

// Checker client
type Checker interface {
	Start() error
	Stop() error
}

type checker struct {
	config         *config.Config
	registryClient client.Discovery
	cachedCatalog  resources.ServiceCatalog
	listener       Listener
	ticker         *time.Ticker
}

// Config options
type Config struct {
	Conf           *config.Config
	RegistryClient client.Discovery
	Listener       Listener
}

// New instantiates new instance
func New(conf Config) Checker {
	return &checker{
		listener:       conf.Listener,
		registryClient: conf.RegistryClient,
		config:         conf.Conf,
	}
}

func (c *checker) Start() error {
	// Stop existing ticker if necessary
	if c.ticker != nil {
		if err := c.Stop(); err != nil {
			logrus.WithError(err).Error("Could not stop existing periodic poll")
			return err
		}
	}

	// TODO make Registry polling interval configurable
	// Create new ticker
	c.ticker = time.NewTicker(c.config.Controller.Poll)

	// Do initial poll
	if err := c.check(); err != nil {
		logrus.WithError(err).Error("Catalog check failed")
	}

	// Start periodic poll
	for range c.ticker.C {
		if err := c.check(); err != nil {
			logrus.WithError(err).Error("Catalog check failed")
		}
	}

	return nil
}

func (c *checker) check() error {
	// Get newest catalog from Registry
	latestCatalog, err := c.getLatestCatalog(c.config.Registry)
	if err != nil {
		logrus.WithError(err).Warn("Could not get latest catalog from registry")
		return err
	}

	// Check for differences
	if !c.catalogsEqual(c.cachedCatalog, latestCatalog) {
		// Update cached copy of catalog
		c.cachedCatalog = latestCatalog

		if err = c.listener.CatalogChange(latestCatalog); err != nil {
			logrus.WithError(err).Warn("Listener failed")
			return err
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
// FIXME: is this conversion still necessary?
func (c *checker) getLatestCatalog(sd config.Registry) (resources.ServiceCatalog, error) {
	catalog := resources.ServiceCatalog{}

	instances, err := c.registryClient.ListInstances(client.InstanceFilter{})
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

// Stop halts the periodic poll of Controller
func (c *checker) Stop() error {
	// Stop ticker if necessary
	if c.ticker != nil {
		c.ticker.Stop()
		c.ticker = nil
	}

	return nil
}
