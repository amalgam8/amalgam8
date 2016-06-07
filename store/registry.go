// Package store defines and implements a backend store for the registry
package store

import (
	"github.com/amalgam8/registry/auth"
	"github.com/amalgam8/registry/replication"
)

const (
	module string = "STORE"
)

// Registry represents the interface of the Service Registry
type Registry interface {

	// Actor: Service Discovery Provider
	GetCatalog(auth.Namespace) (Catalog, error)
}

// New creates a new Registry instance, bounded with the specified configuration and replication
func New(conf *Config, rep replication.Replication) Registry {
	// This is the default implementation for now
	// TODO: Allow customization
	return newInMemoryRegistry(conf, rep)
}
