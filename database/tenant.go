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

// Tenant client
type Tenant interface {
	Create(entry resources.TenantEntry) error
	Read(id string) (resources.TenantEntry, error)
	Update(entry resources.TenantEntry) error
	Delete(id string) error
	List(ids []string) ([]resources.TenantEntry, error)
}

type tenant struct {
	db CloudantDB
}

// NewTenant creates tenant instance
func NewTenant(db CloudantDB) Tenant {
	return &tenant{
		db: db,
	}
}

// Create database entry
func (c *tenant) Create(catalog resources.TenantEntry) error {
	return c.db.InsertEntry(&catalog)
}

// Read database entry
func (c *tenant) Read(id string) (resources.TenantEntry, error) {
	serviceCatalog := resources.TenantEntry{}
	err := c.db.ReadEntry(id, &serviceCatalog)
	return serviceCatalog, err
}

// Update database entry
func (c *tenant) Update(catalog resources.TenantEntry) error {
	return c.db.InsertEntry(&catalog)
}

// Delete database entry
func (c *tenant) Delete(id string) error {
	return c.db.DeleteEntry(id)
}

// List all database IDs
func (c *tenant) List(ids []string) ([]resources.TenantEntry, error) {
	all := AllTenants{}
	err := c.db.ReadAllDocsContent(&all)
	if err != nil {
		return []resources.TenantEntry{}, err
	}

	catalogs := []resources.TenantEntry{}
	for _, row := range all.Rows {
		catalogs = append(catalogs, row.Doc)
	}

	return catalogs, nil
}

// AllTenants struct
type AllTenants struct {
	Rows []struct {
		Doc resources.TenantEntry `json:"doc"`
	} `json:"rows"`
	TotalRows int `json:"total_rows"`
}

// GetEntries returns all database entries
func (at *AllTenants) GetEntries() []Entry {
	entries := make([]Entry, len(at.Rows))
	for i := 0; i < len(at.Rows); i++ {
		entries[i] = &at.Rows[i].Doc
	}
	return entries
}
