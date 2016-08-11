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
	"github.com/amalgam8/registry/auth"
	"github.com/garyburd/redigo/redis"
)

type redisDB struct {
	conn      redis.Conn
	address   string
	password  string
	namespace auth.Namespace
}

// NewRedisDB returns an instance of a Redis database
func NewRedisDB(namespace auth.Namespace, address string, password string) Database {
	db := &redisDB{
		conn:      nil,
		address:   address,
		password:  password,
		namespace: namespace,
	}

	return db
}

// NewRedisDBWithConn returns an instance of a Redis database using an existing connection
func NewRedisDBWithConn(conn redis.Conn, namespace auth.Namespace, address string, password string) Database {
	db := &redisDB{
		conn:      conn,
		address:   address,
		password:  password,
		namespace: namespace,
	}

	return db

}

func (rdb *redisDB) connect() redis.Conn {
	// Connect to Redis
	conn, _ := redis.Dial("tcp", rdb.address)
	_, err := conn.Do("AUTH", rdb.password)
	if err != nil {
		conn.Close()
		panic(err)
	}
	return conn
}

func (rdb *redisDB) ReadKeys() ([]string, error) {
	conn := rdb.conn
	if rdb.conn == nil {
		conn = rdb.connect()
		defer conn.Close()
	}

	hashKeys, err := redis.Strings(conn.Do("HKEYS", rdb.namespace.String()))

	return hashKeys, err
}

func (rdb *redisDB) ReadEntry(key string) (string, error) {
	conn := rdb.conn
	if rdb.conn == nil {
		conn = rdb.connect()
		defer conn.Close()
	}

	entry, err := redis.String(conn.Do("HGET", rdb.namespace.String(), key))

	return entry, err
}

func (rdb *redisDB) ReadAllEntries() (map[string]string, error) {
	conn := rdb.conn
	if rdb.conn == nil {
		conn = rdb.connect()
		defer conn.Close()
	}

	entries, err := redis.StringMap(conn.Do("HGETALL", rdb.namespace.String()))

	return entries, err
}

func (rdb *redisDB) ReadAllMatchingEntries(match string) (map[string]string, error) {
	conn := rdb.conn
	if rdb.conn == nil {
		conn = rdb.connect()
		defer conn.Close()
	}

	var (
		cursor int64
		keys   []string
	)
	var matches = make(map[string]string)

	for {
		items, err := redis.Values(conn.Do("HSCAN", rdb.namespace.String(), cursor, "MATCH", match))
		if err != nil || items == nil || len(items) == 0 {
			return matches, err
		}

		items, err = redis.Scan(items, &cursor, &keys)
		if err != nil {
			return matches, err
		}

		for i := 0; i < len(keys); i++ {
			// Make sure we don't go off the end as the values returned are:
			// key[i]=key key[i+1]=value
			if i+1 > len(keys) {
				break
			}
			matches[keys[i]] = keys[i+1]
			i++
		}
		if cursor == 0 {
			break
		}
	}
	return matches, nil
}

func (rdb *redisDB) InsertEntry(key string, entry string) error {
	conn := rdb.conn
	if rdb.conn == nil {
		conn = rdb.connect()
		defer conn.Close()
	}

	_, err := conn.Do("HSET", rdb.namespace.String(), key, entry)
	return err
}

func (rdb *redisDB) DeleteEntry(key string) (int, error) {
	conn := rdb.conn
	if rdb.conn == nil {
		conn = rdb.connect()
		defer conn.Close()
	}

	return redis.Int(conn.Do("HDEL", rdb.namespace.String(), key))
}
