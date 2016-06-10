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

// Catalog client
type Catalog interface {
	Create(catalog resources.ServiceCatalog) error
	Read(id string) (resources.ServiceCatalog, error)
	Update(catalog resources.ServiceCatalog) error
	Delete(id string) error
	List(ids []string) ([]resources.ServiceCatalog, error)
}

type catalog struct {
	db CloudantDB
}

// NewCatalog creates catalog instance
func NewCatalog(db CloudantDB) Catalog {
	return &catalog{
		db: db,
	}
}

// Create database entry
func (c *catalog) Create(catalog resources.ServiceCatalog) error {
	return c.db.InsertEntry(&catalog)
}

// Read databse entry
func (c *catalog) Read(id string) (resources.ServiceCatalog, error) {
	serviceCatalog := resources.ServiceCatalog{}
	err := c.db.ReadEntry(id, &serviceCatalog)
	return serviceCatalog, err
}

// Update database entry
func (c *catalog) Update(catalog resources.ServiceCatalog) error {
	return c.db.InsertEntry(&catalog)
}

// Delete database entry
func (c *catalog) Delete(id string) error {
	return c.db.DeleteEntry(id)
}

// List all database IDs
func (c *catalog) List(ids []string) ([]resources.ServiceCatalog, error) {
	all := AllServiceCatalogs{}
	err := c.db.ReadAllDocsContent(&all)
	if err != nil {
		return []resources.ServiceCatalog{}, err
	}

	catalogs := []resources.ServiceCatalog{}
	for _, row := range all.Rows {
		catalogs = append(catalogs, row.Doc)
	}

	return catalogs, nil
}

// AllServiceCatalogs struct
type AllServiceCatalogs struct {
	Rows []struct {
		Doc resources.ServiceCatalog `json:"doc"`
	} `json:"rows"`
	TotalRows int `json:"total_rows"`
}

// GetEntries returns all database entries
func (at *AllServiceCatalogs) GetEntries() []Entry {
	entries := make([]Entry, len(at.Rows))
	for i := 0; i < len(at.Rows); i++ {
		entries[i] = &at.Rows[i].Doc
	}
	return entries
}
