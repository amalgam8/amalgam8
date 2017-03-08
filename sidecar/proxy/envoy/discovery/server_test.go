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

package discovery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	server   *Server
	config   Config
	myClient *mySimpleServiceDiscovery
}

/********************* Mock client ***************************/
var discoveryServer Server
var returnError bool

type mySimpleServiceDiscovery struct {
	services []*api.ServiceInstance
}

// ListServices queries the registry for the list of services for which instances are currently registered.
func (m *mySimpleServiceDiscovery) ListServices() ([]string, error) {
	servicesNames := []string{}
	for _, service := range m.services {
		servicesNames = append(servicesNames, service.ServiceName)
	}
	return servicesNames, nil
}

// ListInstances queries the registry for the list of service instances currently registered.
func (m *mySimpleServiceDiscovery) ListInstances() ([]*api.ServiceInstance, error) {
	servicesToReturn := []*api.ServiceInstance{}
	servicesToReturn = append(servicesToReturn, m.services...)
	return servicesToReturn, nil
}

// ListServiceInstances queries the registry for the list of service instances with status 'UP' currently
// registered for the given service.
func (m *mySimpleServiceDiscovery) ListServiceInstances(serviceName string) ([]*api.ServiceInstance, error) {

	servicesToReturn := []*api.ServiceInstance{}
	if returnError {
		return servicesToReturn, fmt.Errorf("Some error has occurred")
	}

	for _, service := range m.services {
		if service.ServiceName == serviceName {
			servicesToReturn = append(servicesToReturn, service)

		}

	}
	return servicesToReturn, nil
}

// create the discovery server with a client , initialize registry with service instances.
func (suite *TestSuite) SetupTest() {
	var err error
	suite.myClient = new(mySimpleServiceDiscovery)

	suite.config = Config{
		HTTPAddressSpec: ":6500",
		Discovery:       suite.myClient,
	}

	discoveryServer, err = NewDiscoveryServer(&suite.config)
	suite.NoError(err)
	suite.server = &discoveryServer

	suite.myClient.services = append(suite.myClient.services, &api.ServiceInstance{ServiceName: "shoppingCart",
		ID: "1", Endpoint: api.ServiceEndpoint{Type: "http", Value: "http://amalgam8/shopping/cart"}})

	suite.myClient.services = append(suite.myClient.services, &api.ServiceInstance{ServiceName: "shoppingCart",
		ID: "2", Endpoint: api.ServiceEndpoint{Type: "tcp", Value: "127.0.0.5:5050"}})

	suite.myClient.services = append(suite.myClient.services, &api.ServiceInstance{Tags: []string{"first", "second"},
		ServiceName: "shoppingCart", ID: "3", Endpoint: api.ServiceEndpoint{Type: "tcp", Value: "127.0.0.4:3050"}})

	suite.myClient.services = append(suite.myClient.services, &api.ServiceInstance{ServiceName: "Orders",
		ID: "4", Endpoint: api.ServiceEndpoint{Type: "tcp", Value: "127.0.0.10:3050"}})

	suite.myClient.services = append(suite.myClient.services, &api.ServiceInstance{ServiceName: "Orders",
		ID: "6", Endpoint: api.ServiceEndpoint{Type: "http", Value: "http://127.0.0.11"}})

	suite.myClient.services = append(suite.myClient.services, &api.ServiceInstance{ServiceName: "Orders",
		ID: "7", Endpoint: api.ServiceEndpoint{Type: "tcp", Value: "132.68.5.6:1010"}})

	suite.myClient.services = append(suite.myClient.services, &api.ServiceInstance{ServiceName: "Reviews",
		ID: "8", Endpoint: api.ServiceEndpoint{Type: "tcp", Value: "132.68.5.6:1010"}})

	suite.myClient.services = append(suite.myClient.services, &api.ServiceInstance{ServiceName: "httpsService",
		ID: "9", Endpoint: api.ServiceEndpoint{Type: "https", Value: "https://127.0.0.12"}})

	go discoveryServer.Start()
	time.Sleep((200) * time.Millisecond)

}

func (suite *TestSuite) TearDownTest() {
	discoveryServer.Stop()
	time.Sleep((200) * time.Millisecond)
}

func (suite *TestSuite) TestGetRegistrationCart() {
	baseURL := "http://localhost:6500/v1/registration"
	handler, _ := discoveryServer.(*server).setup()
	recorder := httptest.NewRecorder()

	// Get shoppingCart service
	req, err := http.NewRequest("GET", baseURL+"/shoppingCart", nil)
	suite.NoError(err)
	handler.ServeHTTP(recorder, req)
	suite.Equal(recorder.Code, http.StatusOK)

	var registration Hosts

	err = json.Unmarshal(recorder.Body.Bytes(), &registration)
	suite.NoError(err)
	// The url endpoint will be filtered out
	suite.Len(registration.Hosts, 2, "Should be two records for shoppingCart")
}

func (suite *TestSuite) TestGetRegistrationOrders() {
	baseURL := "http://localhost:6500/v1/registration"
	handler, _ := discoveryServer.(*server).setup()
	recorder := httptest.NewRecorder()

	// Get Orders service instances
	req, err := http.NewRequest("GET", baseURL+"/Orders", nil)
	suite.NoError(err)
	handler.ServeHTTP(recorder, req)
	suite.Equal(recorder.Code, http.StatusOK)

	var registration Hosts

	err = json.Unmarshal(recorder.Body.Bytes(), &registration)
	suite.NoError(err)
	suite.Len(registration.Hosts, 3, "Should be three records for Orders")
}

func (suite *TestSuite) TestGetRegistrationReviews() {
	baseURL := "http://localhost:6500/v1/registration"
	handler, _ := discoveryServer.(*server).setup()
	recorder := httptest.NewRecorder()

	// Get Orders service instances
	req, err := http.NewRequest("GET", baseURL+"/Reviews", nil)
	suite.NoError(err)
	handler.ServeHTTP(recorder, req)
	suite.Equal(recorder.Code, http.StatusOK)

	var registration Hosts

	err = json.Unmarshal(recorder.Body.Bytes(), &registration)
	suite.NoError(err)
	suite.Len(registration.Hosts, 1, "Should be one record for Reviews")
}

func (suite *TestSuite) TestGetRegistrationHttpsService() {
	baseURL := "http://localhost:6500/v1/registration"
	handler, _ := discoveryServer.(*server).setup()
	recorder := httptest.NewRecorder()

	// Get Orders service instances
	req, err := http.NewRequest("GET", baseURL+"/httpsService", nil)
	suite.NoError(err)
	handler.ServeHTTP(recorder, req)
	suite.Equal(recorder.Code, http.StatusOK)

	var registration Hosts

	err = json.Unmarshal(recorder.Body.Bytes(), &registration)
	suite.NoError(err)
	suite.Len(registration.Hosts, 1, "Should be one record for httpsService")
	// Default https port
	suite.Equal(registration.Hosts[0].Port, uint16(443))
}

func (suite *TestSuite) TestGetRegistrationServiceNotFound() {
	baseURL := "http://localhost:6500/v1/registration"
	handler, _ := discoveryServer.(*server).setup()
	recorder := httptest.NewRecorder()

	// Get Orders service instances
	req, err := http.NewRequest("GET", baseURL+"/noservice", nil)
	suite.NoError(err)
	handler.ServeHTTP(recorder, req)
	suite.Equal(recorder.Code, http.StatusOK)

	var registration Hosts

	err = json.Unmarshal(recorder.Body.Bytes(), &registration)
	suite.NoError(err)
	suite.Len(registration.Hosts, 0, "Should be no records")
}

func (suite *TestSuite) TestGetRegistrationServiceNotSpecified() {
	baseURL := "http://localhost:6500/v1/registration"
	handler, _ := discoveryServer.(*server).setup()
	recorder := httptest.NewRecorder()

	// Get Orders service instances
	req, err := http.NewRequest("GET", baseURL+"/", nil)
	suite.NoError(err)
	handler.ServeHTTP(recorder, req)
	suite.Equal(recorder.Code, http.StatusNotFound)
}

func (suite *TestSuite) TestGetRegistrationError() {
	returnError = true

	baseURL := "http://localhost:6500/v1/registration"
	handler, _ := discoveryServer.(*server).setup()
	recorder := httptest.NewRecorder()

	// Get Orders service instances
	req, err := http.NewRequest("GET", baseURL+"/Reviews", nil)
	suite.NoError(err)
	handler.ServeHTTP(recorder, req)
	suite.Equal(recorder.Code, http.StatusServiceUnavailable)

	returnError = false
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestTestSuite(t *testing.T) {

	suite.Run(t, new(TestSuite))
}
