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
