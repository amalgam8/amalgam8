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

package rules

import (
	"encoding/json"

	"errors"

	"encoding/base64"

	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/controller/util/encryption"
	"github.com/garyburd/redigo/redis"
)

// Entry is used to encapsulate a record with an IV for encryption and decryption.
type Entry struct {
	IV      string `json:"IV"`
	Payload string `json:"payload"`
}

type redisDB struct {
	pool       *redis.Pool
	address    string
	password   string
	encryption encryption.Encryption
}

// TODO: The returns from all the redis commands need to be double checked to ensure we are detecting all the errors

// newRedisDB returns an instance of a Redis database
func newRedisDB(address string, password string) *redisDB {
	db := &redisDB{
		pool: redis.NewPool(func() (redis.Conn, error) {
			// Connect to Redis
			conn, err := redis.DialURL(
				address,
				redis.DialPassword(password),
			)
			if err != nil {
				if conn != nil {
					conn.Close()
				}
				return nil, err
			}
			return conn, nil
		}, 240),
		address:    address,
		password:   password,
		encryption: nil,
	}

	// TODO: either make configurable, or tweak this number appropriately
	db.pool.MaxActive = 30

	return db
}

func (rdb *redisDB) ReadAllEntries(namespace string) ([]string, int64, error) {
	conn := rdb.pool.Get()
	defer conn.Close()

	logrus.Debug("HGETALL ", namespace)
	entryMap, err := redis.StringMap(conn.Do("HGETALL", buildRulesKey(namespace)))
	if err != nil {
		return []string{}, 0, err
	}

	// Transform into an array
	entries := make([]string, len(entryMap))
	i := 0
	for _, entry := range entryMap {
		entries[i] = entry
		i++
	}

	entries, err = rdb.decrypt(entries)
	if err != nil {
		return []string{}, 0, err
	}

	rev, err := redis.Int64(conn.Do("GET", buildNamespaceKey(namespace, "revision"))) // FIXME: pipeline
	if err == redis.ErrNil {
		rev = 0
	}

	return entries, rev, nil
}

func (rdb *redisDB) ReadEntries(namespace string, ids []string) ([]string, int64, error) {
	args := make([]interface{}, len(ids)+1)
	args[0] = buildRulesKey(namespace)
	for i, id := range ids {
		args[i+1] = id
	}

	conn := rdb.pool.Get()
	defer conn.Close()

	logrus.Debug("HMGET ", args)
	entries, err := redis.Strings(conn.Do("HMGET", args...)) // TODO: more error checking?
	if err != nil {
		return []string{}, 0, err
	}

	entries, err = rdb.decrypt(entries)
	if err != nil {
		return []string{}, 0, err
	}

	rev, err := redis.Int64(conn.Do("GET", buildNamespaceKey(namespace, "revision"))) // FIXME: pipeline
	if err == redis.ErrNil {
		rev = 0
	}

	return entries, rev, nil
}

func (rdb *redisDB) InsertEntries(namespace string, entries map[string]string) error {
	encrypted, err := rdb.encrypt(entries)
	if err != nil {
		return err
	}

	args := buildHMSetArgs(buildRulesKey(namespace), encrypted)

	conn := rdb.pool.Get()
	defer conn.Close()

	logrus.Debug("HMSET ", args)
	_, err = redis.String(conn.Do("HMSET", args...))
	if err != nil {
		return err
	}

	_, err = conn.Do("INCR", buildNamespaceKey(namespace, "revision")) // FIXME: pipeline

	return err
}

// 1. Get all existing IDs
// 2. Ensure the new rules are a subset of the existing rules
// 3. Update the rules
func (rdb *redisDB) UpdateEntries(namespace string, entries map[string]string) error {
	key := buildRulesKey(namespace)

	conn := rdb.pool.Get()
	defer conn.Close()

	conn.Do("WATCH", key) // TODO: return codes?

	existingIDs, err := redis.Strings(conn.Do("HKEYS", key))
	if err != nil {
		return err
	}

	existingIDSet := make(map[string]bool)
	for _, id := range existingIDs {
		existingIDSet[id] = true
	}

	// TODO: build a list of all the IDs that are missing?
	for id := range entries {
		_, exists := existingIDSet[id]
		if !exists {
			return errors.New("rules: id " + id + " does not exist")
		}
	}

	entries, err = rdb.encrypt(entries)
	if err != nil {
		return err
	}

	conn.Send("MULTI")
	args := buildHMSetArgs(key, entries)
	if err := conn.Send("HMSET", args...); err != nil {
		return err
	}

	// Execute transaction
	_, err = redis.Values(conn.Do("EXEC"))

	// Nil return indicates that the transaction failed
	if err != nil {
		if err == redis.ErrNil {
			logrus.Error("Transaction failed due to conflict")
		}
		return err
	}

	_, err = conn.Do("INCR", buildNamespaceKey(namespace, "revision")) // FIXME: pipeline

	return err
}

func (rdb *redisDB) DeleteEntries(namespace string, ids []string) error {
	conn := rdb.pool.Get()
	defer conn.Close()

	args := make([]interface{}, len(ids)+1)
	args[0] = buildRulesKey(namespace)
	i := 1
	for _, id := range ids {
		args[i] = id
		i++
	}

	logrus.Debug("HDEL ", args)
	_, err := redis.Int(conn.Do("HDEL", args...))
	// TODO: more error checking?
	if err != nil {
		return err
	}

	_, err = conn.Do("INCR", buildNamespaceKey(namespace, "revision")) // FIXME: pipeline

	return err
}

func (rdb *redisDB) DeleteAllEntries(namespace string) error {
	conn := rdb.pool.Get()
	defer conn.Close()

	_, err := redis.Int(conn.Do("DEL", buildRulesKey(namespace)))
	if err != nil {
		return err
	}

	_, err = conn.Do("INCR", buildNamespaceKey(namespace, "revision")) // FIXME: pipeline

	return err
}

func (rdb *redisDB) SetByDestination(namespace string, filter Filter, rules []Rule) error {
	var err error

	key := buildRulesKey(namespace)

	entries := make(map[string]string)
	for _, rule := range rules {
		entry, err := json.Marshal(&rule)
		if err != nil {
			return err
		}
		entries[rule.ID] = string(entry)
	}

	entries, err = rdb.encrypt(entries)
	if err != nil {
		return err
	}

	conn := rdb.pool.Get()
	defer conn.Close() // Automatically calls DISCARD if necessary

	conn.Do("WATCH", key)

	// Get all rules
	existingEntryMap, err := redis.StringMap(conn.Do("HGETALL", key))
	if err != nil {
		return err
	}

	existingEntries := make([]string, len(existingEntryMap))
	i := 0
	for _, entry := range existingEntryMap {
		existingEntries[i] = entry
		i++
	}

	existingEntries, err = rdb.decrypt(existingEntries)
	if err != nil {
		return err
	}

	// Unmarshal
	existingRules := make([]Rule, len(existingEntries))
	for i, entry := range existingEntries {
		rule := Rule{}
		err := json.Unmarshal([]byte(entry), &rule)
		if err != nil {
			return err
		}

		existingRules[i] = rule
	}

	rulesToDelete := FilterRules(filter, existingRules)
	logrus.WithFields(logrus.Fields{
		"pre_filtered": existingRules,
		"filtered":     rulesToDelete,
		"filter":       filter,
	}).Debug("Filtering")

	conn.Send("MULTI")

	// Delete IDs
	if len(rulesToDelete) > 0 {
		args := make([]interface{}, len(rulesToDelete)+1)
		args[0] = key
		for i, rule := range rulesToDelete {
			args[i+1] = rule.ID
		}
		logrus.Debug("HDEL ", args)

		err = conn.Send("HDEL", args...)
		if err != nil {
			return err
		}
	}

	// Add new rules
	if len(entries) > 0 {
		args := buildHMSetArgs(key, entries)

		logrus.Debug("HMSET ", args)
		err = conn.Send("HMSET", args...)
		if err != nil {
			return err
		}
	}

	// Execute transaction
	_, err = redis.Values(conn.Do("EXEC"))

	if err != nil {
		// Nil return indicates that the transaction failed
		if err == redis.ErrNil {
			logrus.Error("Transaction failed due to conflict")
		}
		return err
	}

	_, err = redis.Int64(conn.Do("INCR", buildNamespaceKey(namespace, "revision"))) // FIXME: pipeline

	return err
}

// encrypt
func (rdb *redisDB) encrypt(entries map[string]string) (map[string]string, error) {
	// Short-circuit without encryption
	if rdb.encryption == nil {
		return entries, nil
	}

	encryptedMap := make(map[string]string)
	for id, entry := range entries {
		iv := rdb.encryption.NewIV()
		payload, err := rdb.encryption.Encrypt(iv, []byte(entry))
		if err != nil {
			logrus.Error("Encryption failed")
			return encryptedMap, err
		}

		encodedIV := base64.StdEncoding.EncodeToString(iv)
		encodedPayload := base64.StdEncoding.EncodeToString(payload)

		e := Entry{
			IV:      encodedIV,
			Payload: encodedPayload,
		}

		data, err := json.Marshal(&e)
		if err != nil {
			logrus.Error("Encryption failed")
			return encryptedMap, err
		}

		encryptedMap[id] = string(data)
	}

	return encryptedMap, nil
}

// decrypt
func (rdb *redisDB) decrypt(entries []string) ([]string, error) {
	// Short-circuit without encryption
	if rdb.encryption == nil {
		return entries, nil
	}

	e := Entry{}

	decryptedEntries := make([]string, len(entries))
	for i, entry := range entries {
		if err := json.Unmarshal([]byte(entry), &e); err != nil {
			return []string{}, err
		}

		decodedPayload, err := base64.StdEncoding.DecodeString(e.Payload)
		if err != nil {
			return []string{}, err
		}

		decodedIV, err := base64.StdEncoding.DecodeString(e.IV)
		if err != nil {
			return []string{}, err
		}

		decrypted, err := rdb.encryption.Decrypt(decodedIV, decodedPayload)
		if err != nil {
			return []string{}, err
		}

		decryptedEntries[i] = string(decrypted)
	}

	return decryptedEntries, nil
}

func buildHMSetArgs(key string, fieldMap map[string]string) []interface{} {
	args := make([]interface{}, len(fieldMap)*2+1)
	args[0] = key

	i := 1
	for id, entry := range fieldMap {
		args[i] = id
		args[i+1] = entry
		i += 2
	}

	return args
}

func buildNamespaceKey(namespace, key string) string {
	return fmt.Sprintf("controller:%v:%v", namespace, key)
}

func buildRulesKey(namespace string) string {
	return fmt.Sprintf("controller:%v:rules", namespace)
}
