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
	"github.com/garyburd/redigo/redis"
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

func (rdb *redisDB) ReadEntry(namespace, key string) (string, error) {
	conn := rdb.pool.Get()
	defer conn.Close()

	entry, err := redis.String(conn.Do("HGET", namespace, key))

	return entry, err
}

func (rdb *redisDB) ReadAllEntries(namespace string) (map[string]string, error) {
	conn := rdb.pool.Get()
	defer conn.Close()

	entries, err := redis.StringMap(conn.Do("HGETALL", namespace))

	return entries, err
}

// TODO: sanitize input?
func (rdb *redisDB) InsertEntry(namespace, key, entry string) error {
	conn := rdb.pool.Get()
	defer conn.Close()

	_, err := conn.Do("HSET", namespace, key, entry)
	return err
}

func (rdb *redisDB) DeleteEntry(namespace, key string) error {
	conn := rdb.pool.Get()
	defer conn.Close()

	// TODO: ensure one field has been removed?
	_, err := conn.Do("HDEL", namespace, key)
	return err
}
