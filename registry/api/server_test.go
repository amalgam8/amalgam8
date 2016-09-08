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

// Consider adding a REST client class to abstract away some of the http details?

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/stretchr/testify/assert"

	"github.com/amalgam8/amalgam8/registry/api/protocol/amalgam8"
	"github.com/amalgam8/amalgam8/registry/api/uptime"
	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/store"
	"github.com/amalgam8/amalgam8/registry/utils/health"
)

const (
	port      = "8080"
	serverURL = "http://localhost:" + port
)

func setupServer(c *Config) (http.Handler, error) {
	s, err := NewServer(c)
	if err != nil {
		return nil, err
	}
	return s.(*server).setup()
}

type mockService struct {
	data store.Service
}

type mockInstance struct {
	data store.ServiceInstance
}

type mockCatalog struct {
	instances map[string]*store.ServiceInstance
	services  []*store.Service
}

func createCatalogMap() store.CatalogMap {
	mc := mockCatalog{
		instances: make(map[string]*store.ServiceInstance),
		services:  []*store.Service{},
	}
	return &mc
}

func (mc *mockCatalog) prepopulateInstances(instances []mockInstance) {
	for _, element := range instances {
		if element.data.ID == "" {
			element.data.ID = generateInstanceID(&element.data)
		}
		inst := element.data
		mc.instances[element.data.ID] = &inst
	}
}

func (mc *mockCatalog) prepopulateServices(services []mockService) {
	for _, element := range services {
		svc := element.data
		mc.services = append(mc.services, &svc)
	}
}

func generateInstanceID(si *store.ServiceInstance) string {
	hash := md5.New()
	_, err := hash.Write([]byte(strings.Join([]string{si.ServiceName, si.Endpoint.Value}, "/")))
	if err != nil {
		return "" // empty
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func (mc *mockCatalog) GetCatalog(namespace auth.Namespace) (store.Catalog, error) {
	return mc, nil
}

func (mc *mockCatalog) Register(si *store.ServiceInstance) (*store.ServiceInstance, error) {
	i := si.DeepClone()
	i.ID = generateInstanceID(si)
	if i.TTL == 0 {
		i.TTL = time.Duration(60) * time.Second
	}
	mc.instances[i.ID] = i
	return i, nil
}

func (mc *mockCatalog) Deregister(iid string) (*store.ServiceInstance, error) {
	instance, ok := mc.instances[iid]
	if ok {
		delete(mc.instances, iid)
		return instance, nil
	}
	return nil, store.NewError(store.ErrorNoSuchServiceInstance, "unable to locate instance", iid)
}

func (mc *mockCatalog) Renew(iid string) (*store.ServiceInstance, error) {
	_, ok := mc.instances[iid]
	if ok { // TODO update the ttl?
		return nil, nil
	}
	return nil, store.NewError(store.ErrorNoSuchServiceInstance, "unable to locate instance", iid)
}

func (mc *mockCatalog) SetStatus(iid, status string) (*store.ServiceInstance, error) {
	inst, ok := mc.instances[iid]
	if ok {
		inst.Status = status
		return inst, nil
	}
	return nil, store.NewError(store.ErrorNoSuchServiceInstance, "unable to locate instance", iid)
}

func (mc *mockCatalog) List(sn string, predicate store.Predicate) ([]*store.ServiceInstance, error) {
	if sn == "" {
		return nil, store.NewError(store.ErrorBadRequest, "null service name", "")
	}

	collection := make([]*store.ServiceInstance, 0, len(mc.instances))
	for _, element := range mc.instances {
		if element.ServiceName == sn {
			if predicate == nil || predicate(element) == true {
				collection = append(collection, element.DeepClone())
			}
		}
	}
	return collection[:len(collection)], nil // return slice of only valid elements
}

func (mc *mockCatalog) Instance(iid string) (*store.ServiceInstance, error) {
	si, ok := mc.instances[iid]
	if ok {
		return si, nil
	}
	return nil, store.NewError(store.ErrorNoSuchServiceInstance, "unable to locate instance", iid)
}

func (mc *mockCatalog) ListServices(predicate store.Predicate) []*store.Service {
	collection := make([]*store.Service, 0, len(mc.services))
	for _, element := range mc.services {
		if element != nil {
			collection = append(collection, element)
		}
	}
	return collection[:len(collection)]
}

func defaultServerConfig() *Config {
	return &Config{
		HTTPAddressSpec: ":" + port,
		CatalogMap:      createCatalogMap(),
	}
}

//---------
// general
//---------

// server uses given configuration
func TestUsingPassedConfig(t *testing.T) {
	c := defaultServerConfig()

	s, err := NewServer(c)
	assert.Nil(t, err)
	assert.Equal(t, s.(*server).config, c)
}

// invalid paths on server
func TestInvalidPaths(t *testing.T) {
	c := defaultServerConfig()
	handler, err := setupServer(c)
	assert.Nil(t, err)

	recorder := httptest.NewRecorder()

	url := serverURL + "/hello"
	req, err := http.NewRequest("GET", url, nil) // invalid URL
	assert.Nil(t, err)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	url = serverURL + amalgam8.InstanceCreateURL()
	req, err = http.NewRequest("GET", url, nil) // valid URL, but no handler
	assert.Nil(t, err)

	handler.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

//--------
// uptime
//--------
func TestRootURL(t *testing.T) {
	url := serverURL
	c := defaultServerConfig()

	handler, err := setupServer(c)
	assert.Nil(t, err)
	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("GET", url, nil)
	assert.Nil(t, err)
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusOK)
}

func TestUptime(t *testing.T) {
	url := serverURL + uptime.URL()
	c := defaultServerConfig()

	handler, err := setupServer(c)
	assert.Nil(t, err)
	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("GET", url, nil)
	assert.Nil(t, err)
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusOK)

	var hc map[string]health.Status

	err = json.Unmarshal(recorder.Body.Bytes(), &hc)
	assert.Nil(t, err)
}

//-----------
// middleware
//-----------

type testMw struct {
	count int
}

func (mw *testMw) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(w rest.ResponseWriter, r *rest.Request) {
		handler(w, r)
		mw.count++
	}
}

func TestExtMiddleware(t *testing.T) {
	tmw := &testMw{}
	url := serverURL + uptime.URL()
	c := defaultServerConfig()
	c.Middlewares = []rest.Middleware{tmw}

	handler, err := setupServer(c)
	assert.Nil(t, err)
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", url, nil)
	assert.Nil(t, err)
	handler.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusOK)
	assert.EqualValues(t, 1, tmw.count)
}
