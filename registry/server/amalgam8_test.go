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

package server

// Consider adding a REST client class to abstract away some of the http details?

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/amalgam8/amalgam8/registry/server/protocol/amalgam8"
	"github.com/amalgam8/amalgam8/registry/store"
	"github.com/amalgam8/amalgam8/registry/utils/i18n"
)

func init() {
	i18n.SupressTestingErrorMessages()
}

//-----------
// instances
//-----------
var metadata = []byte("{\"key1\":\"value1\"}")
var instances = []mockInstance{
	{data: store.ServiceInstance{ID: "http-1", ServiceName: "http",
		Endpoint: &store.Endpoint{Value: "192.168.0.1:80", Type: "http"}, Status: "UP", TTL: 30 * time.Second, Metadata: metadata}},
	{data: store.ServiceInstance{ID: "http-2", ServiceName: "http",
		Endpoint: &store.Endpoint{Value: "192.168.0.2:80", Type: "http"}, Status: "UP", TTL: 30 * time.Second, Metadata: metadata}},
}

// instances:methods
func TestInstancesMethodInvalid(t *testing.T) {
	const fake = "fakeid"
	var methods = []string{"CONNECT", "DELETE", "HEAD", "OPTIONS", "PATCH", "PUT", "TRACE"}
	var urls = []string{
		serverURL + amalgam8.InstanceCreateURL(),
		serverURL + amalgam8.InstanceURL(fake),
	}

	c := defaultServerConfig()
	handler, err := setupServer(c)
	assert.Nil(t, err)

	assert.Nil(t, err)
	for _, url := range urls {
		for _, method := range methods {
			if method == "DELETE" && url == serverURL+amalgam8.InstanceURL(fake) {
				continue // this is a valid combination
			}
			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(method, url, nil)
			assert.Nil(t, err)
			handler.ServeHTTP(recorder, req)
			assert.Equal(t, http.StatusMethodNotAllowed, recorder.Code, method+":"+url)
		}
	}
}

type createTestCase struct {
	instance amalgam8.InstanceRegistration
	expected int
}

func newCreateTestCase(sn string, endpointVal string, endpointType string, status string, ttl uint32, metadata json.RawMessage, httpStatus int) createTestCase {
	testCase := createTestCase{
		instance: amalgam8.InstanceRegistration{
			ServiceName: sn,
			Status:      status,
			TTL:         ttl,
			Metadata:    metadata,
			Tags:        []string{},
			Endpoint:    &amalgam8.InstanceAddress{Value: endpointVal, Type: endpointType}},
		expected: httpStatus}
	return testCase
}

func newCreateTestCaseWithTags(sn string, endpointVal string, endpointType string, status string, ttl uint32, metadata json.RawMessage, httpStatus int, tags []string) createTestCase {
	testCase := createTestCase{
		instance: amalgam8.InstanceRegistration{
			ServiceName: sn,
			Status:      status,
			TTL:         ttl,
			Metadata:    metadata,
			Tags:        tags,
			Endpoint:    &amalgam8.InstanceAddress{Value: endpointVal, Type: endpointType}},
		expected: httpStatus}
	return testCase
}

func (s createTestCase) toByteswithFaultyMetadata() []byte {
	b, err := json.Marshal(&s.instance)
	if err != nil {
		return nil
	}
	b[len(b)-2] = 0
	return b
}

// instances:create
func TestInstancesCreate(t *testing.T) {
	invalidMetadata := json.RawMessage("{\"INVALID\":\"INVALID\"}")

	cases := []createTestCase{
		newCreateTestCase("", "192.168.1.1:8081", "tcp", "UP", 0, metadata, http.StatusBadRequest),                               // empty service name
		newCreateTestCase("http", "", "tcp", "UP", 0, metadata, http.StatusBadRequest),                                           // empty endpoint value
		newCreateTestCase("http", "192.168.1.1:8081", "", "UP", 0, metadata, http.StatusBadRequest),                              // empty endpoint type
		newCreateTestCase("http", "192.168.1.1:8082", "icmp", "UP", 30, metadata, http.StatusBadRequest),                         // invalid endpoint type
		newCreateTestCase("http", "192.168.1.1:8083", "tcp", "UP", 0, invalidMetadata, http.StatusBadRequest),                    // invalid metadata
		newCreateTestCase("http", "192.168.1.1:8084", "tcp", "UP", 0, []byte("1"), http.StatusCreated),                           // valid metadata - int
		newCreateTestCase("http", "192.168.1.1:8085", "tcp", "UP", 0, []byte("true"), http.StatusCreated),                        // valid metadata - bool
		newCreateTestCase("http", "192.168.1.1:8086", "tcp", "UP", 0, []byte("\"string\""), http.StatusCreated),                  // valid metadata - string
		newCreateTestCase("http", "192.168.1.1:8087", "tcp", "UP", 0, metadata, http.StatusCreated),                              // valid metadata - object
		newCreateTestCase("http", "192.168.1.1:8088", "tcp", "UP", 30, metadata, http.StatusCreated),                             // valid, duplicate
		newCreateTestCase("http", "192.168.1.1:8089", "tcp", "STARTING", 30, metadata, http.StatusCreated),                       // valid, STARTING status
		newCreateTestCase("http", "192.168.1.1:8090", "tcp", "OUT_OF_SERVICE", 30, metadata, http.StatusCreated),                 // valid, OUT_OF_SERVICE status
		newCreateTestCase("http", "192.168.1.1:8091", "tcp", "blah", 30, metadata, http.StatusBadRequest),                        // invalid status
		newCreateTestCaseWithTags("http", "192.168.1.1:8088", "tcp", "UP", 30, metadata, http.StatusCreated, []string{"a", "b"}), // valid, with tags
		newCreateTestCaseWithTags("http", "192.168.1.1:8088", "tcp", "UP", 30, metadata, http.StatusCreated, []string{}),         // valid, with empty tags
	}

	url := serverURL + amalgam8.InstanceCreateURL()
	c := defaultServerConfig()
	handler, err := setupServer(c)
	assert.Nil(t, err)

	for _, tc := range cases {
		recorder := httptest.NewRecorder()
		b, err := json.Marshal(&tc.instance)

		assert.NoError(t, err)

		if reflect.DeepEqual(tc.instance.Metadata, invalidMetadata) {
			b = tc.toByteswithFaultyMetadata()
		}

		req, err := http.NewRequest("POST", url, bytes.NewReader(b))

		assert.Nil(t, err)
		req.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(recorder, req)
		assert.Equal(t, tc.expected, recorder.Code, string(b), "\nResponse:", string(recorder.Body.Bytes()))
		if recorder.Code == http.StatusCreated { // verify links
			reply := &amalgam8.ServiceInstance{}

			err = json.Unmarshal(recorder.Body.Bytes(), &reply)
			assert.NoError(t, err)
			assert.Nil(t, reply.Endpoint)
			assert.NotNil(t, reply.Links)
			assert.NotEmpty(t, reply.TTL)
		}
	}
}

func TestInstanceCreateMissingEndpoint(t *testing.T) {
	url := serverURL + amalgam8.InstanceCreateURL()
	c := defaultServerConfig()
	handler, err := setupServer(c)
	assert.Nil(t, err)

	var buggyReq = []byte(`{ "service_name": "service", "host": "whatnot.example.org", "port": 80, "ttl" : 25}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(buggyReq))
	assert.Nil(t, err)

	recorder := httptest.NewRecorder()
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code, string(buggyReq))
}

// instance:delete
func TestInstanceDelete(t *testing.T) {
	cases := []struct {
		iid      string // input service identifier
		expected int    // expected result
	}{
		{"http-1", http.StatusOK},
		{"http-2", http.StatusOK},
		{"http-2", http.StatusGone}, // duplicate delete should fail
		{"http-3", http.StatusGone}, // unknown instance id should fail
	}

	c := defaultServerConfig()
	c.CatalogMap.(*mockCatalog).prepopulateInstances(instances)
	handler, err := setupServer(c)
	assert.Nil(t, err)

	for _, tc := range cases {
		recorder := httptest.NewRecorder()

		req, err := http.NewRequest("DELETE", serverURL+amalgam8.InstanceURL(tc.iid), nil)
		assert.Nil(t, err)
		req.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(recorder, req)
		assert.Equal(t, tc.expected, recorder.Code, string(tc.iid))
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", serverURL+amalgam8.ServiceInstancesURL("http"), nil)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

// instance:heartbeat
func TestInstanceHeartbeat(t *testing.T) {
	cases := []struct {
		iid      string // input service identifier
		expected int    // expected result
	}{
		{"http-1", http.StatusOK},
		{"http-2", http.StatusOK},
		{"http-3", http.StatusGone}, // unknown instance id should fail
	}

	c := defaultServerConfig()
	c.CatalogMap.(*mockCatalog).prepopulateInstances(instances)
	handler, err := setupServer(c)
	assert.Nil(t, err)

	for _, tc := range cases {
		recorder := httptest.NewRecorder()
		// Consider getting the heartbeat URL of instances programmatically?
		req, err := http.NewRequest("PUT", serverURL+amalgam8.InstanceHeartbeatURL(tc.iid), nil)
		assert.Nil(t, err)
		req.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(recorder, req)
		assert.Equal(t, tc.expected, recorder.Code, string(tc.iid))
	}
}

//----------
// services
//----------
// services:methods
func TestServiceInstancesMethods(t *testing.T) {
	var methods = []string{"CONNECT", "HEAD", "OPTIONS", "PATCH", "POST", "PUT", "TRACE"}

	url := serverURL + amalgam8.ServiceInstancesURL("fakeservice")
	c := defaultServerConfig()
	handler, err := setupServer(c)
	assert.Nil(t, err)

	for _, method := range methods {
		recorder := httptest.NewRecorder()
		req, err := http.NewRequest(method, url, nil)
		assert.Nil(t, err)
		handler.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusMethodNotAllowed, recorder.Code, method+":"+url)
	}
}

func TestServiceInstancesFilteringByFieldValues(t *testing.T) {
	c := defaultServerConfig()
	c.CatalogMap.(*mockCatalog).prepopulateServices(services)
	serviceInstance1 := store.ServiceInstance{ID: "http-1", ServiceName: "http-1",
		Endpoint: &store.Endpoint{Value: "192.168.0.1", Type: "tcp"}, Status: "UP", TTL: 30 * time.Second, Metadata: metadata, Tags: []string{"DB", "NoSQL"}}
	serviceInstance2 := store.ServiceInstance{ID: "http-2", ServiceName: "http-1",
		Endpoint: &store.Endpoint{Value: "192.168.0.1", Type: "tcp"}, Status: "UP", TTL: 30 * time.Second, Metadata: metadata, Tags: []string{"DB"}}
	serviceInstance3 := store.ServiceInstance{ID: "http-3", ServiceName: "http-1",
		Endpoint: &store.Endpoint{Value: "192.168.0.1", Type: "tcp"}, Status: "STARTING", TTL: 30 * time.Second, Metadata: metadata, Tags: []string{"DB"}}
	serviceInstance4 := store.ServiceInstance{ID: "http-4", ServiceName: "http-1",
		Endpoint: &store.Endpoint{Value: "192.168.0.1", Type: "tcp"}, Status: "OUT_OF_SERVICE", TTL: 30 * time.Second, Metadata: metadata, Tags: []string{"DB", "NoSQL"}}
	serviceInstance5 := store.ServiceInstance{ID: "http-5", ServiceName: "http-1",
		Endpoint: &store.Endpoint{Value: "192.168.0.1", Type: "tcp"}, Status: "user_defined", TTL: 30 * time.Second, Metadata: metadata, Tags: []string{"DB", "NoSQL"}}
	c.CatalogMap.(*mockCatalog).prepopulateInstances([]mockInstance{
		mockInstance{serviceInstance1},
		mockInstance{serviceInstance2},
		mockInstance{serviceInstance3},
		mockInstance{serviceInstance4},
		mockInstance{serviceInstance5},
	})

	handler, err := setupServer(c)
	assert.Nil(t, err)

	filterParams := []string{
		"?tags=DB,NoSQL",             // 2 tags                                                     // 0
		"?tags=DB",                   // just 1 tag                                                 // 1
		"?tags=DB,MySQL",             // filter out, because we do an AND			    // 2
		"?status=UP",                 // status is up, should be included                           // 3
		"?status=DOWN",               // status is down, should not be included                     // 4
		"?tags=DB,NoSQL&status=UP",   // check multiple field values true                           // 5
		"?tags=DB,NoSQL&status=DOWN", // check multiple field values false                          // 6
		"?ttl=5",                     // check fields of non string/[]string types are bad request  // 7
		"?BLABLA=5",                  // check invalid field is bad request                         // 8
		"?status=up",                 // check status ignore case                                   // 9
		"?status=ALL",                // all the instances with any status                          // 10
		"?status=STARTING",           // all the instances with STARTING status                     // 11
		"?status=OUT_OF_SERVICE",     // all the instances with OUT_OF_SERVICE status               // 12
		"?status=ALL&tags=DB",        // all the instances with any status and DB tag               // 13
		"?status=STARTING&tags=DB",   // all the instances with STARTING status and DB tag          // 14
		"?tags=DB&status=STARTING",   // all the instances with STARTING status and DB tag          // 15
		"?status=user_defined",       // all the instances with user_defined status                 // 16
	}

	assert0 := func(instances []*amalgam8.ServiceInstance, responseStatus int) {
		assert.Equal(t, http.StatusOK, responseStatus)
		assert.Equal(t, 3, len(instances))
	}

	assert1 := func(instances []*amalgam8.ServiceInstance, responseStatus int) {
		assert.Equal(t, http.StatusOK, responseStatus)
		assert.Equal(t, 5, len(instances))
	}

	assert2 := func(instances []*amalgam8.ServiceInstance, responseStatus int) {
		assert.Equal(t, http.StatusOK, responseStatus)
		assert.Equal(t, 0, len(instances))
	}

	assert3 := func(instances []*amalgam8.ServiceInstance, responseStatus int) {
		assert.Equal(t, http.StatusOK, responseStatus)
		assert.Equal(t, 2, len(instances))
	}

	assert4 := func(instances []*amalgam8.ServiceInstance, responseStatus int) {
		assert.Equal(t, http.StatusOK, responseStatus)
		assert.Equal(t, 0, len(instances))
	}

	assert5 := func(instances []*amalgam8.ServiceInstance, responseStatus int) {
		assert.Equal(t, http.StatusOK, responseStatus)
		assert.Equal(t, 1, len(instances))
	}
	assert6 := func(instances []*amalgam8.ServiceInstance, responseStatus int) {
		assert.Equal(t, http.StatusOK, responseStatus)
		assert.Equal(t, 0, len(instances))
	}

	assert7 := func(instances []*amalgam8.ServiceInstance, responseStatus int) {
		assert.Equal(t, http.StatusBadRequest, responseStatus)
	}

	assert8 := func(instances []*amalgam8.ServiceInstance, responseStatus int) {
		assert.Equal(t, http.StatusBadRequest, responseStatus)
	}

	assert9 := func(instances []*amalgam8.ServiceInstance, responseStatus int) {
		assert.Equal(t, http.StatusOK, responseStatus)
		assert.Equal(t, 2, len(instances))
	}

	assert10 := func(instances []*amalgam8.ServiceInstance, responseStatus int) {
		assert.Equal(t, http.StatusOK, responseStatus)
		assert.Equal(t, 5, len(instances))
	}

	assert11 := func(instances []*amalgam8.ServiceInstance, responseStatus int) {
		assert.Equal(t, http.StatusOK, responseStatus)
		assert.Equal(t, 1, len(instances))
	}

	assert12 := func(instances []*amalgam8.ServiceInstance, responseStatus int) {
		assert.Equal(t, http.StatusOK, responseStatus)
		assert.Equal(t, 1, len(instances))
	}

	assert13 := func(instances []*amalgam8.ServiceInstance, responseStatus int) {
		assert.Equal(t, http.StatusOK, responseStatus)
		assert.Equal(t, 5, len(instances))
	}

	assert14 := func(instances []*amalgam8.ServiceInstance, responseStatus int) {
		assert.Equal(t, http.StatusOK, responseStatus)
		assert.Equal(t, 1, len(instances))
	}

	assert15 := func(instances []*amalgam8.ServiceInstance, responseStatus int) {
		assert.Equal(t, http.StatusOK, responseStatus)
		assert.Equal(t, 1, len(instances))
	}

	assert16 := func(instances []*amalgam8.ServiceInstance, responseStatus int) {
		assert.Equal(t, http.StatusOK, responseStatus)
		assert.Equal(t, 1, len(instances))
	}

	asserts := []func([]*amalgam8.ServiceInstance, int){
		assert0,
		assert1,
		assert2,
		assert3,
		assert4,
		assert5,
		assert6,
		assert7,
		assert8,
		assert9,
		assert10,
		assert11,
		assert12,
		assert13,
		assert14,
		assert15,
		assert16,
	}

	for i := range asserts {
		recorder := httptest.NewRecorder()
		req, err := http.NewRequest("GET", serverURL+amalgam8.InstancesURL()+filterParams[i]+"&service_name=http-1", nil)
		assert.Nil(t, err)
		req.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(recorder, req)
		insts := amalgam8.InstancesList{}
		err = json.Unmarshal(recorder.Body.Bytes(), &insts)
		asserts[i](insts.Instances, recorder.Code)
	}
}

func TestServiceInstancesFiltering(t *testing.T) {
	tc := struct {
		sname    string // input service name
		expected int    // expected result
	}{"http-1", http.StatusOK}

	c := defaultServerConfig()
	c.CatalogMap.(*mockCatalog).prepopulateServices(services)
	serviceInstance := store.ServiceInstance{ID: "http-1", ServiceName: "http-1",
		Endpoint: &store.Endpoint{Value: "192.168.0.1:80", Type: "tcp"}, Status: "UP", TTL: 30 * time.Second, Metadata: metadata}
	c.CatalogMap.(*mockCatalog).prepopulateInstances([]mockInstance{{serviceInstance}})

	handler, err := setupServer(c)
	assert.Nil(t, err)

	filterParams := []string{"?fields=status,ttl,status", // duplicate field name in request
		"?fields=status,ttl,status,endpoint", "?fields=", // an empty fields query
		"?fields=endpoint,ttl,metadata,status"} // a "sunny day" test

	assert1 := func(inst *amalgam8.ServiceInstance) {
		assert.Equal(t, uint32(30), inst.TTL)
		assert.EqualValues(t, "UP", inst.Status)
		// make sure we don't get back Last Heartbeat
		assert.Nil(t, inst.LastHeartbeat)
		// make sure we don't get back the metadata
		assert.NotEqual(t, metadata, inst.Metadata)
		// make sure endpoint and id are sent in spite of not being supplied
		assert.Equal(t, "192.168.0.1:80", inst.Endpoint.Value, "Endpoint wasn't sent back")
	}

	assert2 := func(inst *amalgam8.ServiceInstance) {
		assert.Nil(t, inst.LastHeartbeat)
		assert.NotEqual(t, metadata, inst.Metadata)
		assert.Equal(t, "192.168.0.1:80", inst.Endpoint.Value, "Endpoint wasn't sent back")
		assert.Equal(t, "", inst.Status)
	}

	sunnyDayAssert := func(inst *amalgam8.ServiceInstance) {
		assert.Equal(t, serviceInstance.Endpoint.Value, inst.Endpoint.Value)
		assert.Equal(t, int(serviceInstance.TTL), int(inst.TTL)*int(time.Second))
		assert.EqualValues(t, metadata, inst.Metadata)
		assert.Equal(t, serviceInstance.Status, inst.Status)
	}

	asserts := []func(*amalgam8.ServiceInstance){
		assert1, assert1, assert2, sunnyDayAssert,
	}

	for i := range asserts {
		recorder := httptest.NewRecorder()
		req, err := http.NewRequest("GET", serverURL+amalgam8.InstancesURL()+filterParams[i]+"&service_name="+tc.sname, nil)
		assert.Nil(t, err)
		req.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(recorder, req)
		assert.Equal(t, tc.expected, recorder.Code, string(tc.sname))
		if tc.expected == http.StatusOK {
			insts := amalgam8.InstancesList{}
			err = json.Unmarshal(recorder.Body.Bytes(), &insts)
			assert.Equal(t, 1, len(insts.Instances))
			assert.NoError(t, err)
			inst := insts.Instances[0]
			asserts[i](inst)
		}
	}
}

// services/<name>:list
func TestServiceInstances(t *testing.T) {
	cases := []struct {
		sname    string // input service name
		expected int    // expected result
	}{
		{"http", http.StatusOK},
		{"", http.StatusNotFound},      // empty service name
		{"https", http.StatusNotFound}, // non-existing service
	}

	c := defaultServerConfig()
	c.CatalogMap.(*mockCatalog).prepopulateInstances(instances)
	handler, err := setupServer(c)
	assert.Nil(t, err)

	for _, tc := range cases {
		recorder := httptest.NewRecorder()
		req, err := http.NewRequest("GET", serverURL+amalgam8.ServiceInstancesURL(tc.sname), nil)
		assert.Nil(t, err)
		req.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(recorder, req)
		assert.Equal(t, tc.expected, recorder.Code, string(tc.sname))
		if tc.expected == http.StatusOK {
			svc := amalgam8.InstanceList{}

			err = json.Unmarshal(recorder.Body.Bytes(), &svc)
			assert.NoError(t, err)
			assert.Equal(t, svc.ServiceName, tc.sname)
			assert.NotNil(t, svc.Instances)
			assert.Equal(t, len(instances), len(svc.Instances))
			for _, inst := range svc.Instances {
				assert.EqualValues(t, "UP", inst.Status)
				assert.NotNil(t, inst.LastHeartbeat)
				assert.EqualValues(t, metadata, inst.Metadata)
			}
		}
	}
}

// services/<name>:methods
func TestServicesListMethods(t *testing.T) {
	var methods = []string{"CONNECT", "HEAD", "OPTIONS", "PATCH", "POST", "PUT", "TRACE"}

	url := serverURL + amalgam8.ServiceNamesURL()
	c := defaultServerConfig()
	handler, err := setupServer(c)
	assert.Nil(t, err)

	for _, method := range methods {
		recorder := httptest.NewRecorder()
		req, err := http.NewRequest(method, url, nil)
		assert.Nil(t, err)
		handler.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusMethodNotAllowed, recorder.Code, method+":"+url)
	}
}

var services = []mockService{
	{data: store.Service{ServiceName: "http-1"}},
	{data: store.Service{ServiceName: "http-2"}},
	{data: store.Service{ServiceName: "http-3"}},
}

// /services: list name
func TestServicesList(t *testing.T) {
	cases := []struct {
		expected int // expected result
	}{
		{http.StatusOK},
	}

	url := serverURL + amalgam8.ServiceNamesURL()
	c := defaultServerConfig()
	c.CatalogMap.(*mockCatalog).prepopulateServices(services)
	handler, err := setupServer(c)
	assert.Nil(t, err)

	for _, tc := range cases {
		recorder := httptest.NewRecorder()
		req, err := http.NewRequest("GET", url, nil)
		assert.Nil(t, err)
		req.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(recorder, req)
		assert.Equal(t, tc.expected, recorder.Code)
		if tc.expected == http.StatusOK {
			list := amalgam8.ServicesList{}

			err = json.Unmarshal(recorder.Body.Bytes(), &list)
			assert.NoError(t, err)
			assert.NotNil(t, list.Services)
			assert.Equal(t, len(services), len(list.Services))
		}
	}
}

//---------------
// secure access
//---------------
func TestRequireHTTPS(t *testing.T) {
	url := serverURL + amalgam8.ServiceNamesURL()
	c := defaultServerConfig()
	handler, err := setupServer(c)
	assert.Nil(t, err)

	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("GET", url, nil)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(recorder, req)
	assert.NotEqual(t, recorder.Code, http.StatusMovedPermanently)

	// Now, set up a a new handler, which requires HTTPS
	c.RequireHTTPS = true
	secureHandler, err := setupServer(c)
	assert.Nil(t, err)

	recorder = httptest.NewRecorder()
	secureHandler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusMovedPermanently)

	req.Header.Set("X-Forwarded-Proto", "https")
	recorder = httptest.NewRecorder()
	secureHandler.ServeHTTP(recorder, req)
	assert.NotEqual(t, recorder.Code, http.StatusMovedPermanently)
}
