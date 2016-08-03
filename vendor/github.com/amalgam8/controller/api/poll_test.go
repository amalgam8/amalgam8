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
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/amalgam8/controller/checker"
	"github.com/amalgam8/controller/metrics"
	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NGINX API", func() {
	var (
		api *Poll
		ch  *checker.MockChecker
		h   http.Handler
	)

	BeforeEach(func() {
		reporter := metrics.NewReporter()
		ch = &checker.MockChecker{}

		api = NewPoll(reporter, ch)

		a := rest.NewApi()
		router, err := rest.MakeRouter(api.Routes()...)
		Expect(err).ToNot(HaveOccurred())
		a.SetApp(router)
		h = a.MakeHandler()
	})

	It("polls successfully", func() {
		req, err := http.NewRequest("POST", "/v1/poll", nil)
		Expect(err).ToNot(HaveOccurred())
		req.Header.Set("Content-Type", "application/json")
		//req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(http.StatusOK))
	})

	It("reports when poll fails", func() {
		ch.CheckError = errors.New("poll failed")

		req, err := http.NewRequest("POST", "/v1/poll", nil)
		Expect(err).ToNot(HaveOccurred())
		req.Header.Set("Content-Type", "application/json")
		//req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		Expect(w.Code).ToNot(Equal(http.StatusOK))
	})

})
