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
	"net/http"
	"net/http/httptest"

	"github.com/amalgam8/controller/checker"
	"github.com/amalgam8/controller/metrics"
	"github.com/amalgam8/controller/nginx"
	"github.com/amalgam8/controller/resources"
	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NGINX API", func() {
	var (
		api       *NGINX
		generator *nginx.MockGenerator
		h         http.Handler
		chker     *checker.MockChecker
	)

	BeforeEach(func() {
		reporter := metrics.NewReporter()
		generator = &nginx.MockGenerator{}
		chker = new(checker.MockChecker)
		chker.GetVal = resources.ServiceCatalog{
			Services: []resources.Service{},
		}

		api = NewNGINX(NGINXConfig{
			Reporter:  reporter,
			Generator: generator,
			Checker:   chker,
		})

		a := rest.NewApi()
		router, err := rest.MakeRouter(api.Routes()...)
		Expect(err).ToNot(HaveOccurred())
		a.SetApp(router)
		h = a.MakeHandler()
	})

	// TODO: specifically return 404?
	//	It("rejects requests on non-existent tenants", func() {
	//		req, err := http.NewRequest("GET", "/v2/tenants/not_a_tenant/nginx", nil)
	//		Expect(err).ToNot(HaveOccurred())
	//		req.Header.Set("Content-type", "application/json")
	//		//req.Header.Set("Authorization", token)
	//		w := httptest.NewRecorder()
	//		h.ServeHTTP(w, req)
	//		fmt.Println(string(w.Body.Bytes()))
	//		Expect(w.Code).To(Equal(http.StatusNotFound))
	//	})

	It("provides a generated NGINX config", func() {
		generator.GenerateString = "abcdef"

		req, err := http.NewRequest("GET", "/v1/tenants/abcdef/nginx", nil)
		Expect(err).ToNot(HaveOccurred())
		req.Header.Set("Content-type", "application/json")
		//req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(http.StatusOK))
		// TODO: ensure response body is what was provided by generator
		Expect(string(w.Body.Bytes())).To(Equal(generator.GenerateString))
	})

})
