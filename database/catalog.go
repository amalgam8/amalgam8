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
