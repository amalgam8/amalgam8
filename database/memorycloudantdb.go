package database

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
)

// CloudantDB interfaces for working with database data
// TODO: we need to change this to a more generic version for standalone mode, since we won't be using Cloudant
type CloudantDB interface {
	ReadKeys() ([]string, error)
	ReadEntry(key string, entry Entry) error
	InsertEntry(entry Entry) error
	DeleteEntry(key string) error
	ReadAllDocsContent(allDocs AllDocs) error
	DBExists(dbname string) (bool, error)
}

// memoryCloudantDB In memory mock implementation of the database client
type memoryCloudantDB struct {
	mutex   sync.RWMutex
	records map[string][]byte
}

// NewMemoryCloudantDB creates a new database
func NewMemoryCloudantDB() CloudantDB {
	db := new(memoryCloudantDB)
	db.records = make(map[string][]byte)
	return db
}

func (db *memoryCloudantDB) ReadKeys() ([]string, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	keys := make([]string, len(db.records))
	index := 0
	for key := range db.records {
		keys[index] = key
		index++
	}
	return keys, nil
}

func (db *memoryCloudantDB) ReadEntry(key string, entry Entry) error {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	data, exists := db.records[key]
	if !exists {
		return NewDatabaseError("Could not find requested key", "404-Not Found", "key not found", http.StatusNotFound)
	}

	err := json.Unmarshal(data, entry)
	if err != nil {
		return err
	}

	return nil
}

func (db *memoryCloudantDB) InsertEntry(entry Entry) error {

	db.mutex.Lock()
	defer db.mutex.Unlock()

	id, rev := entry.IDRev()

	_, exists := db.records[id]
	if exists && rev == "" {
		return NewDatabaseError("Document update conflict", "409-Update Conflict", "entry already exists", http.StatusConflict)
	}

	entry.SetRev()

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	db.records[id] = data
	return nil
}

func (db *memoryCloudantDB) DeleteEntry(key string) error {

	db.mutex.Lock()
	defer db.mutex.Unlock()
	_, exists := db.records[key]
	if !exists {
		return NewDatabaseError("Could not find requested key", "404-Not Found", "key not found", http.StatusNotFound)
	}

	delete(db.records, key)
	return nil
}

// ReadAllDocsContent return all documents with their from database
func (db *memoryCloudantDB) ReadAllDocsContent(allDocs AllDocs) error {

	db.mutex.Lock()
	defer db.mutex.Unlock()

	totRows := len(db.records)

	totRowsString := strconv.Itoa(totRows)
	var b bytes.Buffer

	// build _all_docs?include_docs=true response
	b.WriteString(`{ "total_rows": ` + totRowsString + `, "rows": [ `)

	if totRows > 0 {
		cont := 0
		for _, val := range db.records {
			b.WriteString(`{ "doc": ` + string(val) + ` }`)
			if cont < len(db.records)-1 {
				b.WriteString(", ")
			}
			cont++
		}
	}

	b.WriteString(" ] } ")
	err := json.Unmarshal(b.Bytes(), &allDocs)
	if err != nil {
		return err
	}

	return nil
}

func (db *memoryCloudantDB) DBExists(dbname string) (bool, error) {
	return true, nil
}
