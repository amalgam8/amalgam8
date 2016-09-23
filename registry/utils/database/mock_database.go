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
	"path/filepath"
	"time"
)

type mockDB struct {
	record map[string]string
}

// NewMockDB returns an instance of a mock database
func NewMockDB() Database {
	data := make(map[string]string)
	db := &mockDB{
		record: data,
	}

	return db
}

func (mdb *mockDB) ReadKeys(match string) ([]string, error) {
	var keyList []string
	for key := range mdb.record {
		matched, _ := filepath.Match(match, key)
		if matched {
			keyList = append(keyList, key)
		}
	}

	return keyList, nil
}

func (mdb *mockDB) ReadEntry(key string) ([]byte, error) {
	return []byte(mdb.record[key]), nil
}

func (mdb *mockDB) ReadAllEntries(match string) (map[string]string, error) {
	entryList := make(map[string]string)
	for key, value := range mdb.record {
		matched, _ := filepath.Match(match, key)
		if matched {
			entryList[key] = value
		}
	}

	return entryList, nil
}

func (mdb *mockDB) InsertEntry(key string, entry []byte) error {
	mdb.record[key] = string(entry[:])
	return nil
}

func (mdb *mockDB) DeleteEntry(key string) (int, error) {
	if _, ok := mdb.record[key]; ok {
		delete(mdb.record, key)
		return 1, nil
	}

	return 0, nil
}

func (mdb *mockDB) Expire(key string, ttl time.Duration) error {
	return nil
}
