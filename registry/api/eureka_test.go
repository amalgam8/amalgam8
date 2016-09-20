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

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/amalgam8/amalgam8/registry/api/protocol/eureka"
	"github.com/amalgam8/amalgam8/registry/store"
)

//-----------
// instances
//-----------
var eurekaInstances = []mockInstance{
	{data: store.ServiceInstance{ID: "http:http-1", ServiceName: "http",
		Endpoint: &store.Endpoint{Value: "192.168.0.1:80", Type: "http"}, Status: "STARTING", TTL: 30 * time.Second, Metadata: metadata,
		Extension: map[string]interface{}{"eureka": "{\"hostName\": \"localhost\", \"vipAddress\": \"http-vip\", \"ipAddr\": \"192.168.0.1\", \"port\": {\"@enabled\": true, \"$\": 80}, \"CountryId\": 1}"}}},
	{data: store.ServiceInstance{ID: "http:http-2", ServiceName: "http",
		Endpoint: &store.Endpoint{Value: "192.168.0.2:80", Type: "http"}, Status: "STARTING", TTL: 30 * time.Second, Metadata: metadata,
		Extension: map[string]interface{}{"eureka": "{\"hostName\": \"localhost\", \"vipAddress\": \"http-vip\", \"ipAddr\": \"192.168.0.2\", \"port\": {\"@enabled\": true, \"$\": 80}, \"CountryId\": 1}"}}},
}

// instances:create
type createEurekaTestCase struct {
	instance eureka.Instance
	expected int
}

func newCreateEurekaTestCase(hostname, appid, ipaddr, vipaddr, port string, metadata json.RawMessage, httpStatus int) createEurekaTestCase {
	testCase := createEurekaTestCase{
		instance: eureka.Instance{
			HostName:    hostname,
			Application: appid,
			IPAddr:      ipaddr,
			VIPAddr:     vipaddr,
			Port:        &eureka.Port{Enabled: "true", Value: port},
			Lease:       &eureka.LeaseInfo{DurationInt: 30},
			Metadata:    metadata,
		},
		expected: httpStatus,
	}
	return testCase
}

func (s createEurekaTestCase) toByteswithFaultyMetadata() []byte {
	b, err := json.Marshal(&eureka.InstanceWrapper{Inst: &s.instance})
	if err != nil {
		return nil
	}
	b[len(b)-2] = 0
	return b
}

func TestEurekaInstancesCreate(t *testing.T) {
	invalidMetadata := json.RawMessage("{\"INVALID\":\"INVALID\"}")

	cases := []createEurekaTestCase{
		newCreateEurekaTestCase("", "http", "192.168.1.1", "http-vip", "8080", metadata, http.StatusBadRequest),                 // empty hostname
		newCreateEurekaTestCase("localhost", "http", "", "http-vip", "8080", metadata, http.StatusBadRequest),                   // empty IP address
		newCreateEurekaTestCase("localhost", "http", "192.168.1.1", "", "8080", metadata, http.StatusBadRequest),                // empty VIPaddr
		newCreateEurekaTestCase("localhost", "http", "192.168.1.1", "http-vip", "8080", invalidMetadata, http.StatusBadRequest), // invalid metadata
		newCreateEurekaTestCase("localhost", "http", "192.168.1.1", "http-vip", "8080", metadata, http.StatusNoContent),         // valid
	}

	c := defaultServerConfig()
	handler, err := setupServer(c)
	assert.Nil(t, err)

	url := serverURL + eureka.ApplicationURL("", "http")

	for _, tc := range cases {
		recorder := httptest.NewRecorder()
		b, err := json.Marshal(&eureka.InstanceWrapper{Inst: &tc.instance})
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
			reply := &eureka.Instance{}

			err = json.Unmarshal(recorder.Body.Bytes(), &reply)
			assert.NoError(t, err)
		}
	}
}

// instance:delete
func TestEurekaInstanceDelete(t *testing.T) {
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
	c.CatalogMap.(*mockCatalog).prepopulateInstances(eurekaInstances)
	handler, err := setupServer(c)
	assert.Nil(t, err)

	for _, tc := range cases {
		recorder := httptest.NewRecorder()

		req, err := http.NewRequest("DELETE", serverURL+eureka.InstanceURL("", "http", tc.iid), nil)
		assert.Nil(t, err)
		req.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(recorder, req)
		assert.Equal(t, tc.expected, recorder.Code, string(tc.iid))
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", serverURL+eureka.ApplicationURL("", "http"), nil)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
}

// instance:heartbeat
func TestEurekaInstanceHeartbeat(t *testing.T) {
	cases := []struct {
		iid      string // input service identifier
		expected int    // expected result
	}{
		{"http-1", http.StatusOK},
		{"http-2", http.StatusOK},
		{"http-3", http.StatusGone}, // unknown instance id should fail
	}

	c := defaultServerConfig()
	c.CatalogMap.(*mockCatalog).prepopulateInstances(eurekaInstances)
	handler, err := setupServer(c)
	assert.Nil(t, err)

	for _, tc := range cases {
		recorder := httptest.NewRecorder()
		req, err := http.NewRequest("PUT", serverURL+eureka.InstanceURL("", "http", tc.iid), nil)
		assert.Nil(t, err)
		req.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(recorder, req)
		assert.Equal(t, tc.expected, recorder.Code, string(tc.iid))
	}
}

// instance:status
func TestEurekaInstanceStatus(t *testing.T) {
	cases := []struct {
		iid      string // input service identifier
		status   string // input instance status
		expected int    // expected result
	}{
		{"http-1", "UP", http.StatusOK},
		{"http-2", "DOWN", http.StatusOK},
		{"http-3", "UNKNOWN", http.StatusNotFound}, // unknown instance id should fail
	}

	c := defaultServerConfig()
	c.CatalogMap.(*mockCatalog).prepopulateInstances(eurekaInstances)
	handler, err := setupServer(c)
	assert.Nil(t, err)

	for _, tc := range cases {
		recorder := httptest.NewRecorder()
		req, err := http.NewRequest("PUT", serverURL+eureka.InstanceStatusURL("", "http", tc.iid)+"?value="+tc.status, nil)
		assert.Nil(t, err)
		req.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(recorder, req)
		assert.Equal(t, tc.expected, recorder.Code, string(tc.iid))

		if tc.expected == http.StatusOK {
			// Verify that the status was changed
			var inst eureka.InstanceWrapper
			req, err = http.NewRequest("GET", serverURL+eureka.InstanceURL("", "http", tc.iid), nil)
			assert.Nil(t, err)
			req.Header.Set("Content-Type", "application/json")
			handler.ServeHTTP(recorder, req)
			assert.Equal(t, tc.expected, recorder.Code, string(http.StatusOK))
			err = json.Unmarshal(recorder.Body.Bytes(), &inst)
			assert.Nil(t, err)
			assert.EqualValues(t, tc.status, inst.Inst.Status)
		}
	}
}

//--------------
// applications
//--------------

// apps/<name>
func TestEurekaAppInstances(t *testing.T) {
	cases := []struct {
		sname    string // input service name
		nInsts   int    // number of returned instances
		expected int    // expected result
	}{
		{"http", len(eurekaInstances), http.StatusOK},
		{"https", 0, http.StatusOK}, // non-existing service
	}

	c := defaultServerConfig()
	c.CatalogMap.(*mockCatalog).prepopulateInstances(eurekaInstances)
	handler, err := setupServer(c)
	assert.Nil(t, err)

	for _, tc := range cases {
		recorder := httptest.NewRecorder()
		req, err := http.NewRequest("GET", serverURL+eureka.ApplicationURL("", tc.sname), nil)
		assert.Nil(t, err)
		req.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(recorder, req)
		assert.Equal(t, tc.expected, recorder.Code, string(tc.sname))
		if tc.expected == http.StatusOK {
			var m map[string]eureka.Application

			err = json.Unmarshal(recorder.Body.Bytes(), &m)
			assert.NoError(t, err)
			app := m["application"]
			assert.NotNil(t, app)
			assert.Equal(t, app.Name, tc.sname)
			if tc.nInsts > 0 {
				assert.NotNil(t, app.Instances)
				assert.Equal(t, len(instances), len(app.Instances))
				for _, inst := range app.Instances {
					assert.EqualValues(t, "STARTING", inst.Status)
					assert.EqualValues(t, metadata, inst.Metadata)
				}
			} else {
				assert.Nil(t, app.Instances)
			}

		}
	}
}
