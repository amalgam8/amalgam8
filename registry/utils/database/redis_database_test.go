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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock the connection object
type MockedConn struct {
	mock.Mock
}

func (c *MockedConn) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	margs := c.Called(commandName, args)
	return margs.Get(0), margs.Error(1)
}

func (c *MockedConn) Close() error {
	return nil
}

func (c *MockedConn) Err() error {
	return nil
}

func (c *MockedConn) Flush() error {
	return nil
}

func (c *MockedConn) Receive() (reply interface{}, err error) {
	return nil, nil
}

func (c *MockedConn) Send(commandName string, args ...interface{}) error {
	return nil
}

type Endpoint struct {
	Type  string
	Value string
}
type ServiceInstance struct {
	ID               string
	ServiceName      string
	Endpoint         *Endpoint
	Status           string
	Metadata         []byte
	RegistrationTime time.Time
	LastRenewal      time.Time
	TTL              time.Duration
	Tags             []string
	Extension        map[string]interface{}
}

func TestRedisDBReadKeys(t *testing.T) {
	var expectedKeys []interface{}
	expectedKeys = append(expectedKeys, []byte("key1"))
	expectedKeys = append(expectedKeys, []byte("key2"))
	expectedKeys = append(expectedKeys, []byte("key3"))

	expectedStrings := []string{"key1", "key2", "key3"}

	mockedConn := new(MockedConn)
	db := NewRedisDBWithConn(mockedConn, "addr", "pass")

	var s []interface{}

	b1 := []byte{'0'}
	s = append(s, b1)
	s = append(s, expectedKeys)

	mockedConn.On("Do", "SCAN", []interface{}{int64(0), "MATCH", "*:test"}).Return(s, nil)

	keys, err := db.ReadKeys("*:test")

	assert.NoError(t, err)
	assert.NotNil(t, keys)
	assert.Equal(t, expectedStrings, keys)
}

func TestRedisDBReadEntry(t *testing.T) {
	key := "key1"
	value := []byte("value1")
	expectedValue := []byte("value1")

	mockedConn := new(MockedConn)
	db := NewRedisDBWithConn(mockedConn, "addr", "pass")

	mockedConn.On("Do", "GET", []interface{}{key}).Return([]byte(value), nil)

	entry, err := db.ReadEntry(key)

	assert.NoError(t, err)
	assert.Equal(t, expectedValue, entry)
}

func TestRedisDBReadEntryErrorReturned(t *testing.T) {
	key := "key1error"
	getError := fmt.Errorf("Error calling GET")

	mockedConn := new(MockedConn)
	db := NewRedisDBWithConn(mockedConn, "addr", "pass")

	mockedConn.On("Do", "GET", []interface{}{key}).Return(nil, getError)

	_, err := db.ReadEntry(key)

	assert.Error(t, err)
	assert.Equal(t, getError, err)
}

func TestRedisDBReadAllEntries(t *testing.T) {
	// Setup the map that we expect to be returned from ReadlAllEntries
	expectedMap := make(map[string]string)
	expectedMap["key1"] = "value1"
	expectedMap["key2"] = "value2"

	mockedConn := new(MockedConn)
	db := NewRedisDBWithConn(mockedConn, "addr", "pass")

	// Setup the keys we expect to be returned from the scan
	var s, expectedKeys []interface{}
	expectedKeys = append(expectedKeys, []byte("key1"))
	expectedKeys = append(expectedKeys, []byte("key2"))

	b1 := []byte{'0'}
	s = append(s, b1)
	s = append(s, expectedKeys)

	// Mock the scan for all keys for the namespace and the gets for the value
	mockedConn.On("Do", "SCAN", []interface{}{int64(0), "MATCH", "*:test"}).Return(s, nil)

	values := []interface{}{[]byte("value1"), []byte("value2")}
	mockedConn.On("Do", "MGET", []interface{}{"key1", "key2"}).Return(values, nil)

	entries, err := db.ReadAllEntries("*:test")

	assert.NoError(t, err)
	assert.NotNil(t, entries)
	assert.Equal(t, expectedMap, entries)
}

func TestRedisDBInsertEntry(t *testing.T) {
	key := "key1"
	entry := []byte("entry1")

	mockedConn := new(MockedConn)
	db := NewRedisDBWithConn(mockedConn, "addr", "pass")

	mockedConn.On("Do", "SET", []interface{}{key, entry}).Return(123, nil)

	err := db.InsertEntry(key, entry)

	assert.NoError(t, err)
}

func TestRedisDBDeleteEntry(t *testing.T) {
	mockedConn := new(MockedConn)
	db := NewRedisDBWithConn(mockedConn, "addr", "pass")

	mockedConn.On("Do", "DEL", []interface{}{"inst-id"}).Return([]byte("1"), nil)

	del, err := db.DeleteEntry("inst-id")

	assert.NoError(t, err)
	assert.Equal(t, 1, del)
}

func TestRedisDBExpireEntry(t *testing.T) {
	mockedConn := new(MockedConn)
	db := NewRedisDBWithConn(mockedConn, "addr", "pass")

	ttl := time.Second * 120
	mockedConn.On("Do", "EXPIRE", []interface{}{"inst-id", ttl.Seconds()}).Return([]byte("1"), nil)

	err := db.Expire("inst-id", ttl)

	assert.NoError(t, err)
}

func TestRedisDBExpireEntryError(t *testing.T) {
	mockedConn := new(MockedConn)
	db := NewRedisDBWithConn(mockedConn, "addr", "pass")

	expireError := fmt.Errorf("Expire error")
	ttl := time.Second * 120
	mockedConn.On("Do", "EXPIRE", []interface{}{"inst-id", ttl.Seconds()}).Return([]byte("1"), expireError)

	err := db.Expire("inst-id", ttl)

	assert.Error(t, err)
	assert.Equal(t, expireError, err)
}
