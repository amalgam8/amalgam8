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
	"encoding/json"
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

	mockedConn.On("Do", "HKEYS", []interface{}{"test"}).Return(expectedKeys, nil)

	keys, err := db.ReadKeys("test")

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

	mockedConn.On("Do", "HGET", []interface{}{"test", key}).Return([]byte(value), nil)

	entry, err := db.ReadEntry("test", key)

	assert.NoError(t, err)
	assert.Equal(t, expectedValue, entry)
}

func TestRedisDBReadEntryErrorReturned(t *testing.T) {
	key := "key1error"
	hgetError := fmt.Errorf("Error calling HGET")

	mockedConn := new(MockedConn)
	db := NewRedisDBWithConn(mockedConn, "addr", "pass")

	mockedConn.On("Do", "HGET", []interface{}{"test", key}).Return(nil, hgetError)

	_, err := db.ReadEntry("test", key)

	assert.Error(t, err)
	assert.Equal(t, hgetError, err)
}

func TestRedisDBReadAllEntries(t *testing.T) {
	expectedMap := make(map[string]string)
	expectedMap["key1"] = "value1"
	expectedMap["key2"] = "value2"
	expectedMap["key3"] = "value3"

	var expectedValues []interface{}
	expectedValues = append(expectedValues, []byte("key1"))
	expectedValues = append(expectedValues, []byte("value1"))
	expectedValues = append(expectedValues, []byte("key2"))
	expectedValues = append(expectedValues, []byte("value2"))
	expectedValues = append(expectedValues, []byte("key3"))
	expectedValues = append(expectedValues, []byte("value3"))

	mockedConn := new(MockedConn)
	db := NewRedisDBWithConn(mockedConn, "addr", "pass")

	mockedConn.On("Do", "HGETALL", []interface{}{"test"}).Return(expectedValues, nil)

	entries, err := db.ReadAllEntries("test")

	assert.NoError(t, err)
	assert.NotNil(t, entries)
	assert.Equal(t, expectedMap, entries)
}

func TestRedisDBReadAllMatchingEntries(t *testing.T) {
	si := &ServiceInstance{
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "192.168.0.1", Type: "tcp"},
	}

	mockedConn := new(MockedConn)
	db := NewRedisDBWithConn(mockedConn, "addr", "pass")

	s, _ := generateMockHScanCommandOutput("inst-id", si)

	mockedConn.On("Do", "HSCAN", []interface{}{"test", int64(0), "MATCH", "inst-id"}).Return(s, nil)

	instance, err := db.ReadAllMatchingEntries("test", "inst-id")
	assert.NoError(t, err)

	var actualSI ServiceInstance
	err = json.Unmarshal([]byte(instance["inst-id"]), &actualSI)

	assert.Equal(t, "inst-id", actualSI.ID)
	assert.Equal(t, "Calc", actualSI.ServiceName)
}

func TestRedisDBInsertEntry(t *testing.T) {
	key := "key1"
	entry := []byte("entry1")

	mockedConn := new(MockedConn)
	db := NewRedisDBWithConn(mockedConn, "addr", "pass")

	mockedConn.On("Do", "HSET", []interface{}{"test", key, entry}).Return(123, nil)

	err := db.InsertEntry("test", key, entry)

	assert.NoError(t, err)
}

func TestRedisDBDeleteEntry(t *testing.T) {
	mockedConn := new(MockedConn)
	db := NewRedisDBWithConn(mockedConn, "addr", "pass")

	mockedConn.On("Do", "HDEL", []interface{}{"test", "inst-id"}).Return([]byte("1"), nil)

	hdel, err := db.DeleteEntry("test", "inst-id")

	assert.NoError(t, err)
	assert.Equal(t, 1, hdel)
}

func generateMockHScanCommandOutput(instID string, instance *ServiceInstance) ([]interface{}, *ServiceInstance) {
	if instID != "" {
		instance.ID = instID
	}
	instanceJSON, _ := json.Marshal(instance)

	var s, sBytes []interface{}

	b1 := []byte{'0'}
	s = append(s, b1)

	var instanceData interface{}
	instBytes := []byte(instID)
	instanceData = instBytes
	sBytes = append(sBytes, instanceData)

	instBytes = []byte(instanceJSON)
	instanceData = instBytes

	sBytes = append(sBytes, instanceData)

	s = append(s, sBytes)

	return s, instance
}
