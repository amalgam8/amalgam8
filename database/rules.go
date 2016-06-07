package database

import (
	"github.com/amalgam8/controller/resources"
)

// Rules TODO
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

// NewRules TODO
func NewRules(db CloudantDB) Rules {
	return &rules{
		db: db,
	}
}

// Create TODO
func (r *rules) Create(proxy resources.ProxyConfig) error {

	//TODO need to do struct conversion
	return r.db.InsertEntry(&proxy)
}

// Read TODO
func (r *rules) Read(id string) (resources.ProxyConfig, error) {
	proxyConfig := resources.ProxyConfig{}
	err := r.db.ReadEntry(id, &proxyConfig)

	//TODO struct deconversion

	return proxyConfig, err
}

//Update TODO
func (r *rules) Update(proxy resources.ProxyConfig) error {

	//TODO struct conversion

	return r.db.InsertEntry(&proxy)
}

// Delete
func (r *rules) Delete(id string) error {
	return r.db.DeleteEntry(id)
}

// List TODO
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

// AllProxyConfigs TODO
type AllProxyConfigs struct {
	Rows []struct {
		Doc resources.ProxyConfig `json:"doc"`
	} `json:"rows"`
	TotalRows int `json:"total_rows"`
}

// GetEntries TODO
func (at *AllProxyConfigs) GetEntries() []Entry {
	entries := make([]Entry, len(at.Rows))
	for i := 0; i < len(at.Rows); i++ {
		entries[i] = &at.Rows[i].Doc
	}
	return entries
}
