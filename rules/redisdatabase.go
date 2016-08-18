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

	"github.com/garyburd/redigo/redis"
	"fmt"
)

type redisDB struct {
	pool     *redis.Pool
	address  string
	password string
}

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

	conn.Send("MULTI")
	for _, id := range ids {
		conn.Send("HGET", namespace, id)
	}
	entries, err := redis.Strings(conn.Do("EXEC")) // TODO: validate each response?
	if err != nil {
		return []string{}, err
	}

	return entries, nil
}

func (rdb *redisDB) InsertEntries(namespace string, entries map[string]string) error {
	conn := rdb.pool.Get()
	defer conn.Close()

	conn.Send("MULTI")
	for id, entry := range entries {
		fmt.Println(entry)
		conn.Send("HSET", namespace, id, entry)
	}
	_, err := conn.Do("EXEC")

	// TODO: validate each response?

	return err
}

func (rdb *redisDB) DeleteEntries(namespace string, ids []string) error {
	conn := rdb.pool.Get()
	defer conn.Close()

	conn.Send("MULTI")
	for _, id := range ids {
		conn.Send("HDEL", namespace, id)
	}
	_, err := conn.Do("EXEC")

	// TODO: more error checking?

	return err
}

const (
	RuleAny = iota
	RuleRoute
	RuleAction
)

func (rdb *redisDB) SetByDestination(namespace string, destinations []string, ruleType int, rules []Rule) error {
	// Get all rules
	entryMap, err := rdb.ReadAllEntries(namespace)
	if err != nil {
		return err
	}

	// Get IDs filtered by destination
	idsToDelete := make([]string, 0, len(entryMap))
	for _, entry := range entryMap {
		rule := Rule{}
		err := json.Unmarshal([]byte(entry), &rule)
		if err != nil {
			return err
		}

		for _, destination := range destinations {
			if rule.Destination == destination {
				if (ruleType == RuleAction && len(rule.Action) > 0) ||
					(ruleType == RuleRoute && len(rule.Route) > 0) {
					idsToDelete = append(idsToDelete, rule.ID)
				}
			}
		}
	}
	conn := rdb.pool.Get()
	defer conn.Close()

	conn.Send("MULTI")

	// Add new rules
	fmt.Println("Destination insert")
	for _, rule := range rules {
		entry, err := json.Marshal(&rule)
		if err != nil {
			return err
		}

		fmt.Println(string(entry))

		conn.Send("HSET", namespace, rule.ID, string(entry))
	}

	// Delete IDs
	for _, id := range idsToDelete {
		conn.Send("HDEL", namespace, id)
	}

	// Execute transaction
	_, err = conn.Do("EXEC")

	return err
}
