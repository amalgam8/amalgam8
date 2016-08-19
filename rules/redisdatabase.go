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

	"github.com/Sirupsen/logrus"
	"github.com/garyburd/redigo/redis"
)

type redisDB struct {
	pool     *redis.Pool
	address  string
	password string
}

// TODO: The returns from all the redis commands need to be double checked to ensure we are detecting all the errors

// NewRedisDB returns an instance of a Redis database
func NewRedisDB(address string, password string) *redisDB {
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
		address:  address,
		password: password,
	}

	// TODO: either make configurable, or tweak this number appropriately
	db.pool.MaxActive = 30

	return db
}

func (rdb *redisDB) ReadKeys(namespace string) ([]string, error) {
	conn := rdb.pool.Get()
	defer conn.Close()

	hashKeys, err := redis.Strings(conn.Do("HKEYS", namespace))

	return hashKeys, err
}

func (rdb *redisDB) ReadAllEntries(namespace string) (map[string]string, error) {
	conn := rdb.pool.Get()
	defer conn.Close()

	entries, err := redis.StringMap(conn.Do("HGETALL", namespace))

	return entries, err
}

func (rdb *redisDB) ReadEntries(namespace string, ids []string) ([]string, error) {
	conn := rdb.pool.Get()
	defer conn.Close()

	args := make([]interface{}, len(ids)+1)
	args[0] = namespace
	for i, id := range ids {
		args[i+1] = id
	}

	return redis.Strings(conn.Do("HMGET", args...))
	// TODO: more error checking?
}

func (rdb *redisDB) InsertEntries(namespace string, entries map[string]string) error {
	conn := rdb.pool.Get()
	defer conn.Close()

	args := make([]interface{}, len(entries)*2+1)
	args[0] = namespace

	i := 1
	for id, entry := range entries {
		args[i] = id
		args[i+1] = entry
		i += 2
	}

	_, err := redis.String(conn.Do("HMSET", args...))
	// TODO: more error checking?

	return err
}

func (rdb *redisDB) DeleteEntries(namespace string, ids []string) error {
	conn := rdb.pool.Get()
	defer conn.Close()

	args := make([]interface{}, len(ids)+1)
	args[0] = namespace
	i := 1
	for id := range ids {
		args[i] = id
		i++
	}

	_, err := redis.Int(conn.Do("HDEL", args...))
	// TODO: more error checking?

	return err
}

const (
	RuleAny = iota
	RuleRoute
	RuleAction
)

func (rdb *redisDB) SetByDestination(namespace string, filter Filter, rules []Rule) error {
	entries := make([]string, len(rules))
	for i, rule := range rules {
		entry, err := json.Marshal(&rule)
		if err != nil {
			return err
		}
		entries[i] = string(entry)
	}

	conn := rdb.pool.Get()
	defer conn.Close() // Automatically calls DISCARD if necessary

	conn.Do("WATCH", namespace)

	// Get all rules
	existingEntries, err := redis.StringMap(conn.Do("HGETALL", namespace))
	if err != nil {
		return err
	}

	// Unmarshal
	existingRules := make([]Rule, 0, len(existingEntries))
	for _, entry := range existingEntries {
		rule := Rule{}
		err := json.Unmarshal([]byte(entry), &rule)
		if err != nil {
			return err
		}

		existingRules = append(existingRules, rule)
	}

	rulesToDelete := FilterRules(filter, existingRules)
	logrus.WithFields(logrus.Fields{
		"pre_filtered": existingRules,
		"filtered":     rulesToDelete,
		"filter":       filter,
	}).Debug("Filtering")

	conn.Send("MULTI")

	// Add new rules
	if len(entries) > 0 {
		args := make([]interface{}, len(entries)*2+1)
		args[0] = namespace
		i := 1
		for id, entry := range entries {
			args[i] = id
			args[i+1] = entry
			i += 2
		}

		err = conn.Send("HMSET", args...)
		if err != nil {
			return err
		}
	}

	// Delete IDs
	if len(rulesToDelete) > 0 {
		args := make([]interface{}, len(rulesToDelete)+1)
		args[0] = namespace
		for i, rule := range rulesToDelete {
			args[i+1] = rule.ID
		}
		logrus.Debug("HDEL", args)

		err = conn.Send("HDEL", args...)
		if err != nil {
			return err
		}
	}

	// Execute transaction
	_, err = redis.Values(conn.Do("EXEC"))

	// Nil return indicates that the transaction failed
	if err == redis.ErrNil {
		logrus.Error("Transaction failed due to conflict")
	}

	return err
}
