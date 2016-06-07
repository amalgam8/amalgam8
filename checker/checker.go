package checker

import (
	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/clients"
	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/controller/notification"
	"github.com/amalgam8/controller/proxyconfig"
	"github.com/amalgam8/controller/resources"
	"reflect"
	"sort"
	"time"
)

// Checker TODO
type Checker interface {
	Register(id string) error
	Deregister(id string) error
	Check(ids []string) error
	Get(id string) (resources.ServiceCatalog, error)
}

type checker struct {
	db            database.Catalog
	proxyConfig   proxyconfig.Manager
	registry      clients.Registry
	producerCache notification.TenantProducerCache
}

// Config options
type Config struct {
	Database      database.Catalog
	ProxyConfig   proxyconfig.Manager
	Registry      clients.Registry
	ProducerCache notification.TenantProducerCache
}

// New instantiates new instance
func New(conf Config) Checker {
	return &checker{
		db:            conf.Database,
		proxyConfig:   conf.ProxyConfig,
		registry:      conf.Registry,
		producerCache: conf.ProducerCache,
	}
}

// Register tenant ID
func (c *checker) Register(id string) error {
	// TODO: either call Registry or set to valid default
	defaultCatalog := resources.ServiceCatalog{
		BasicEntry: resources.BasicEntry{
			ID: id,
		},
		Services:   []resources.Service{},
		LastUpdate: time.Now(),
	}

	if err := c.db.Create(defaultCatalog); err != nil {
		return err
	}

	return nil
}

// Deregister tenant ID
func (c *checker) Deregister(id string) error {
	// TODO: Cleanup caches?

	return c.db.Delete(id)
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
	catalogs, err := c.getStoredCatalogs(ids)
	if err != nil {
		// log failure
		return err
	}

	for _, catalog := range catalogs {

		// Get Registry credentials from auth
		creds, err := c.credentials(catalog.ID)
		if err != nil {
			logrus.WithError(err).Warn("Could not get credentials")
			continue
		}

		// Get newest catalog from Registry
		latestCatalog, err := c.getLatestCatalog(creds.Registry)
		if err != nil {
			// log failure

			//TODO get new token from auth and try again on 401's
			logrus.WithError(err).Warn("Could not get latest from Registry")

			continue
		}

		// Check for differences
		if !c.catalogsEqual(catalog, latestCatalog) {
			// Update database
			catalog.Services = latestCatalog.Services
			catalog.LastUpdate = time.Now()

			if err = c.db.Update(catalog); err != nil {
				// log failure
				continue // no point in notifying tenant
			}

			// Notify tenant
			if err = c.producerCache.SendEvent(catalog.ID, creds.Kafka); err != nil {
				logrus.WithFields(logrus.Fields{
					"err":       err,
					"tenant_id": catalog.ID,
					// TODO request ID logging??
				}).Error("Failed to notify tenant of rules change")
			}

		}
	}

	return nil
}

func (c *checker) credentials(id string) (resources.Credentials, error) {
	proxyConfig, err := c.proxyConfig.Get(id)
	return proxyConfig.Credentials, err
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

func (c *checker) getStoredCatalogs(ids []string) ([]resources.ServiceCatalog, error) {
	catalogs, err := c.db.List(ids)
	if err != nil {
		return catalogs, err
	}

	return catalogs, nil
}

// Get the tenant's catalog from the database
func (c *checker) Get(id string) (resources.ServiceCatalog, error) {
	return c.db.Read(id)
}
