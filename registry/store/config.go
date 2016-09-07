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

package store

import (
	"time"

	"github.com/amalgam8/amalgam8/registry/replication"
	"github.com/amalgam8/amalgam8/registry/utils/database"
)

const (
	defaultDefaultTTL        = time.Duration(30) * time.Second
	defaultMinimumTTL        = time.Duration(5) * time.Second
	defaultMaximumTTL        = time.Duration(10) * time.Minute
	defaultSyncTimeout       = time.Duration(30) * time.Second
	defaultNamespaceCapacity = 50
	defaultStore             = "inmem"
	defaultStoreAddr         = ""
	defaultStorePassword     = ""
)

// DefaultConfig is the default configuration parameters for the registry
var DefaultConfig = NewConfig(defaultDefaultTTL, defaultMinimumTTL, defaultMaximumTTL, defaultNamespaceCapacity, nil, nil, defaultStore, defaultStoreAddr, defaultStorePassword, nil)

// Config encapsulates the registry configuration parameters
type Config struct {
	DefaultTTL time.Duration
	MinimumTTL time.Duration
	MaximumTTL time.Duration

	NamespaceCapacity int

	SyncWaitTime time.Duration

	Extensions  []CatalogFactory
	Replication replication.Replication

	Store         string
	StoreAddr     string
	StorePassword string
	StoreDatabase database.Database
}

// NewConfig creates a new registry configuration according to the specified TTL values
func NewConfig(defaultTTL, minimumTTL, maximumTTL time.Duration, namespaceCapacity int, extensions []CatalogFactory, rep replication.Replication, store string, storeAddr string, storePassword string, storeDatabase database.Database) *Config {
	validate(defaultTTL, minimumTTL, maximumTTL, namespaceCapacity)
	return &Config{
		DefaultTTL:        defaultTTL,
		MinimumTTL:        minimumTTL,
		MaximumTTL:        maximumTTL,
		SyncWaitTime:      defaultSyncTimeout,
		NamespaceCapacity: namespaceCapacity,
		Extensions:        extensions,
		Replication:       rep,
		Store:             store,
		StoreAddr:         storeAddr,
		StorePassword:     storePassword,
		StoreDatabase:     storeDatabase,
	}
}

func validate(defaultTTL, minimumTTL, maximumTTL time.Duration, namespaceCapacity int) {
	if minimumTTL > maximumTTL {
		panic("Maximum TTL must be larger or equal to minimum TTL")
	}
	if defaultTTL < minimumTTL {
		panic("Default TTL must be larger or equal to minimum TTL")
	}
	if defaultTTL > maximumTTL {
		panic("Default TTL must be smaller or equal to maximum TTL")
	}
	if namespaceCapacity < -1 {
		panic("Namespace capacity must be greater than or equal to -1")
	}
}
