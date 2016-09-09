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

package filesystem

import (
	"fmt"
	"os"
	"time"

	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/store"
)

// Config encapsulates FileSystem configuration parameters
type Config struct {
	Dir             string
	PollingInterval time.Duration
}

type fsFactory struct {
	config *Config
}

// New creates and initializes a FileSystem catalog factory
func New(conf *Config) (store.CatalogFactory, error) {
	if conf == nil || conf.Dir == "" {
		return nil, fmt.Errorf("Failed to create FileSystem catalog factory: config is not valid")
	}

	// Check if the directory exists
	fsinfo, err := os.Stat(conf.Dir)
	if err != nil {
		return nil, fmt.Errorf("Failed to create FileSystem catalog factory: %s", err)
	}
	if !fsinfo.IsDir() {
		return nil, fmt.Errorf("Failed to create FileSystem catalog factory: %s is not a directory", conf.Dir)
	}

	if conf.PollingInterval == 0 {
		conf.PollingInterval = defaultPollingInterval
	}

	if conf.PollingInterval < minPollingInterval {
		conf.PollingInterval = minPollingInterval
	}

	return &fsFactory{config: conf}, nil
}

func (f *fsFactory) CreateCatalog(namespace auth.Namespace) (store.Catalog, error) {
	return newFileSystemCatalog(namespace, f.config)
}
