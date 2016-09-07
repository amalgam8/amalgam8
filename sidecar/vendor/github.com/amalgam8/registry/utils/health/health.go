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

package health

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

const (
	// CheckPanicked is reported when a health check panics
	CheckPanicked = "healthcheck panic"

	// HTTPStatusCodeHealthChecksPass is the HTTP Status code used to indicate success
	HTTPStatusCodeHealthChecksPass = http.StatusOK
	// HTTPStatusCodeHealthChecksFail is the HTTP Status code used to indicate health check failure
	HTTPStatusCodeHealthChecksFail = http.StatusServiceUnavailable
)

// The Checker type defines an interface with a single Check() function that determines the health of a component and
// returns its status back to the caller.
type Checker interface {
	// Check performs a health check of the component.
	// Health checks should normally return quickly, and avoid synchronous network calls or long-running computations.
	// If such operations are needed, they should be performed in the background (e.g., by a separate goroutine).
	Check() Status
}

// The CheckerFunc is an adapter to allow the use of ordinary functions as health checkers.
// If fn is a function with the appropriate signature, CheckerFunc(fn) is a Checker object that calls fn.
type CheckerFunc func() Status

// Check calls fn()
func (fn CheckerFunc) Check() Status {
	return fn()
}

// Register adds the Checker and named component to the set of monitored components.
func Register(name string, check Checker) {
	healthchecksMutex.Lock()
	defer healthchecksMutex.Unlock()

	healthchecks[name] = check
}

// RegisterFunc adds the Checker function and named component to the set of monitored components.
func RegisterFunc(name string, checker func() Status) {
	Register(name, CheckerFunc(checker))
}

// Unregister removes the health checker currently registered for the named component.
func Unregister(name string) {
	healthchecksMutex.Lock()
	defer healthchecksMutex.Unlock()

	delete(healthchecks, name)
}

// Components returns the registered components names (in arbitrary order).
func Components() []string {
	healthchecksMutex.Lock()
	defer healthchecksMutex.Unlock()

	names := make([]string, 0, len(healthchecks))
	for name := range healthchecks {
		names = append(names, name)
	}
	return names
}

// RunChecks executes all health checks, returning a mapping between registered component names and their health
// check status.
func RunChecks() map[string]Status {
	healthchecksMutex.Lock()
	defer healthchecksMutex.Unlock()

	var m sync.Mutex
	var wg sync.WaitGroup
	var c = len(healthchecks)

	wg.Add(c)

	results := make(map[string]Status, c)
	for name, check := range healthchecks {
		// run each component in its own go-routine to allow parallel execution
		go func(name string, hc Checker) {
			defer wg.Done()

			res := checkComponent(hc)

			m.Lock()
			defer m.Unlock()

			results[name] = res
		}(name, check)
	}
	wg.Wait()
	return results
}

func checkComponent(hc Checker) (s Status) {
	defer func() {
		r := recover()
		if r != nil { // panicked - attempt to convert the common cases to an error
			msg := CheckPanicked
			if _, ok := r.(error); ok {
				s = StatusUnhealthy(msg, r.(error))
			} else if _, ok = r.(string); ok {
				s = StatusUnhealthy(msg, errors.New(r.(string)))
			} else if _, ok = r.(fmt.GoStringer); ok {
				s = StatusUnhealthy(msg, errors.New(r.(fmt.GoStringer).GoString()))
			} else if _, ok = r.(fmt.Stringer); ok {
				s = StatusUnhealthy(msg, errors.New(r.(fmt.Stringer).String()))
			} else { // panicked in an unexpected way
				panic(r)
			}
		}
	}()

	return hc.Check()
}

// Handler returns an http.HandlerFunc that can be used to retrieve the JSON representation of health check statuses.
// The returned representation is a mapping from component name to its status.
func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hc := RunChecks()

		b, err := json.Marshal(hc)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		responseCode := HTTPStatusCodeHealthChecksPass
		for _, status := range hc {
			if !status.Healthy { // health check failure, set the handler's response code to indicate failure
				responseCode = HTTPStatusCodeHealthChecksFail
				break
			}
		}

		w.WriteHeader(responseCode)
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(b)
	}
}

// health checks repository.
var (
	healthchecks      = make(map[string]Checker)
	healthchecksMutex sync.Mutex
)
