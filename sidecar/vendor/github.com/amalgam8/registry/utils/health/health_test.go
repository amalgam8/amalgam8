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

package health_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/amalgam8/registry/utils/health"
	"github.com/stretchr/testify/assert"
)

const (
	COMPONENT = "component"
	FAILING   = "failing"
)

// testify.TestSuite could be used to get automatic teardown cleanup after tests
func teardown() { // clear all components from health check set
	for _, c := range health.Components() {
		health.Unregister(c)
	}
}

//-----------------------------------------------------------------------------
// registration tests
func alwaysHealthy() health.Status {
	return health.Healthy
}

func alwaysFailing() health.Status {
	return health.StatusUnhealthy(FAILING, errors.New("undefined failure"))
}

var (
	healthy = health.CheckerFunc(alwaysHealthy)
	failed  = health.CheckerFunc(alwaysFailing)
)

func TestRegister(t *testing.T) {
	health.Register(COMPONENT, healthy)
	assert.Len(t, health.Components(), 1)
	assert.Contains(t, health.Components(), COMPONENT)
	all := health.RunChecks()

	status, found := all[COMPONENT]
	assert.True(t, found)
	assert.EqualValues(t, status, healthy.Check())

	teardown()
}

func TestRegisterFunc(t *testing.T) {
	health.RegisterFunc(COMPONENT, alwaysFailing)
	assert.Len(t, health.Components(), 1)
	assert.Contains(t, health.Components(), COMPONENT)
	all := health.RunChecks()

	status, found := all[COMPONENT]
	assert.True(t, found)
	assert.EqualValues(t, status, alwaysFailing())

	teardown()
}

func TestRegisterDuplicateHealthCheck(t *testing.T) {
	health.Register(COMPONENT, healthy)
	health.RegisterFunc(COMPONENT, alwaysFailing) // replace with failing health check
	assert.Len(t, health.Components(), 1)
	assert.Contains(t, health.Components(), COMPONENT)
	all := health.RunChecks()

	status, found := all[COMPONENT]
	assert.True(t, found)
	assert.EqualValues(t, status, alwaysFailing())

	teardown()
}

func TestDeregisterHealthCheck(t *testing.T) {
	health.Register(COMPONENT, healthy)
	health.Unregister(COMPONENT)
	assert.Len(t, health.Components(), 0)
	assert.NotContains(t, health.Components(), COMPONENT)
	all := health.RunChecks()
	assert.Len(t, all, 0)

	teardown()
}

func TestMultipleRegistrations(t *testing.T) {
	testcases := []struct {
		name    string         // component
		checker health.Checker // health checker
	}{
		{"component-alive", healthy},
		{"component-up", healthy},
		{"component-failed", failed},
		{"component-dead", failed},
		{"component-ok", healthy},
	}

	for _, tc := range testcases {
		health.Register(tc.name, tc.checker)
	}
	assert.Len(t, health.Components(), len(testcases)) // all registered

	all := health.RunChecks()

	for _, tc := range testcases {
		assert.Contains(t, health.Components(), tc.name)
		_, found := all[tc.name]
		assert.True(t, found)
		assert.EqualValues(t, all[tc.name], tc.checker.Check(), tc.name)
	}

	teardown()
}

//-----------------------------------------------------------------------------
// execute health checks tests
type stubHealthCheck struct {
	healthy bool
}

func (stub *stubHealthCheck) Check() health.Status {
	if stub.healthy {
		return health.StatusHealthy("I'm a lumberjack and I'm OK")
	}
	return health.StatusUnhealthy(FAILING, nil)
}

func TestHealthCheckExecuteHealthy(t *testing.T) {
	health.Register(COMPONENT, &stubHealthCheck{healthy: true})

	checks := health.RunChecks()

	for component, hc := range checks {
		assert.Equal(t, component, COMPONENT)
		assert.True(t, hc.Healthy)
		assert.Empty(t, hc.Properties["cause"])
	}

	teardown()
}

func TestHealthCheckExecuteUnhealthy(t *testing.T) {
	health.Register(COMPONENT, &stubHealthCheck{healthy: false})

	checks := health.RunChecks()

	for component, hc := range checks {
		assert.Equal(t, component, COMPONENT)
		assert.False(t, hc.Healthy)
		assert.Equal(t, FAILING, hc.Properties["message"])
	}

	teardown()
}

func TestParallelExecution(t *testing.T) {
	start := time.Now()
	expected_end := start.Add(time.Second)

	for i := 0; i < 5; i++ {
		health.RegisterFunc(fmt.Sprint("Sleeper-", i), func() health.Status {
			time.Sleep(1 * time.Second)
			return health.Healthy
		})
	}

	_ = health.RunChecks()
	// we don't really expect it to diverge by more than ~100ms, to avoid false negatives, allow one
	// second drift - still considerably less than the 5s if running checks sequenctially
	assert.WithinDuration(t, expected_end, time.Now(), time.Second, "not running concurrently")

	teardown()
}

//-----------------------------------------------------------------------------
// recover from health check panic and mark component as failing
func TestRecoverFromHealthCheckPanic(t *testing.T) {
	cause := "oops I did it again"

	health.RegisterFunc("Will panic", func() health.Status {
		panic(cause)
	})

	checks := health.RunChecks()

	assert.Len(t, checks, 1)
	for _, hc := range checks {
		assert.False(t, hc.Healthy)
		assert.Equal(t, health.CheckPanicked, hc.Properties["message"])
		assert.Equal(t, cause, hc.Properties["cause"])
	}

	teardown()
}

//-----------------------------------------------------------------------------
// publishing health checks
var published = map[string]health.CheckerFunc{
	"pass":  func() health.Status { return health.Healthy },
	"fail":  func() health.Status { return health.StatusUnhealthy(FAILING, nil) },
	"panic": func() health.Status { panic("oops!") },
}

func TestHTTPHandlerSuccess(t *testing.T) {
	testCase := "pass"
	health.RegisterFunc(testCase, published[testCase])

	handler := health.Handler()
	req, err := http.NewRequest("GET", "http://ignored.path.to.dontcare.example.com/", nil)
	assert.Nil(t, err)
	w := httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, health.HTTPStatusCodeHealthChecksPass, w.Code)
	teardown()
}

func TestHTTPHandlerFail(t *testing.T) {
	testCase := "fail"
	health.RegisterFunc(testCase, published[testCase])

	handler := health.Handler()
	req, err := http.NewRequest("GET", "http://ignored.path.to.dontcare.example.com/", nil)
	assert.Nil(t, err)
	w := httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, health.HTTPStatusCodeHealthChecksFail, w.Code)
	teardown()
}

func TestHTTPHandlerPartialFail(t *testing.T) {
	testCase := "pass"
	health.RegisterFunc(testCase, published[testCase])
	testCase = "fail"
	health.RegisterFunc(testCase, published[testCase])

	handler := health.Handler()
	req, err := http.NewRequest("GET", "http://ignored.path.to.dontcare.example.com/", nil)
	assert.Nil(t, err)
	w := httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, health.HTTPStatusCodeHealthChecksFail, w.Code)
	teardown()
}

func TestHTTPHandlerReturnsAll(t *testing.T) {
	for name, fn := range published {
		health.RegisterFunc(name, fn)
	}

	handler := health.Handler()
	req, err := http.NewRequest("GET", "http://ignored.path.to.dontcare.example.com/", nil)
	assert.Nil(t, err)
	w := httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, health.HTTPStatusCodeHealthChecksFail, w.Code)

	var actual map[string]health.Status
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), &actual))

	var expected = map[string]health.Status{
		"pass":  health.Healthy,
		"fail":  health.StatusUnhealthy(FAILING, nil),
		"panic": health.StatusUnhealthy(health.CheckPanicked, errors.New("oops!")),
	}

	assert.Len(t, actual, len(expected))
	assert.EqualValues(t, actual, expected)

	teardown()
}
