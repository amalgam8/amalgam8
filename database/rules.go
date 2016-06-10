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

package database

import (
	"github.com/amalgam8/controller/resources"
)

// Rules client
type Rules interface {
	Create(catalog resources.ProxyConfig) error
	Read(id string) (resources.ProxyConfig, error)
	Update(catalog resources.ProxyConfig) error
	Delete(id string) error
	List() ([]resources.ProxyConfig, error)
}

type rules struct {
	db CloudantDB
}

// NewRules creates Rules instance
func NewRules(db CloudantDB) Rules {
	return &rules{
		db: db,
	}
}

// Create database entry
func (r *rules) Create(proxy resources.ProxyConfig) error {

	//TODO need to do struct conversion
	return r.db.InsertEntry(&proxy)
}

// Read database entry
func (r *rules) Read(id string) (resources.ProxyConfig, error) {
	proxyConfig := resources.ProxyConfig{}
	err := r.db.ReadEntry(id, &proxyConfig)

	//TODO struct deconversion

	return proxyConfig, err
}

//Update database entry
func (r *rules) Update(proxy resources.ProxyConfig) error {

	//TODO struct conversion

	return r.db.InsertEntry(&proxy)
}

// Delete database entry
func (r *rules) Delete(id string) error {
	return r.db.DeleteEntry(id)
}

// List all database entry IDs
func (r *rules) List() ([]resources.ProxyConfig, error) {
	all := AllProxyConfigs{}
	err := r.db.ReadAllDocsContent(&all)
	if err != nil {
		return []resources.ProxyConfig{}, err
	}

	configs := []resources.ProxyConfig{}
	for _, entry := range all.GetEntries() {
		config := entry.(*resources.ProxyConfig)

		// TODO struct deconversion

		configs = append(configs, *config)
	}

	return configs, nil
}

// AllProxyConfigs struct
type AllProxyConfigs struct {
	Rows []struct {
		Doc resources.ProxyConfig `json:"doc"`
	} `json:"rows"`
	TotalRows int `json:"total_rows"`
}

// GetEntries returns all database entries
func (at *AllProxyConfigs) GetEntries() []Entry {
	entries := make([]Entry, len(at.Rows))
	for i := 0; i < len(at.Rows); i++ {
		entries[i] = &at.Rows[i].Doc
	}
	return entries
}
