package api

import (
	"github.com/amalgam8/controller/checker"
	"github.com/amalgam8/controller/nginx"
	"github.com/amalgam8/controller/resources"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/cactus/go-statsd-client/statsd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("NGINX API", func() {
	var (
		api       *NGINX
		generator *nginx.MockGenerator
		h         http.Handler
		chker     *checker.MockChecker
	)

	BeforeEach(func() {
		statsdClient, err := statsd.NewClient("", "") // TODO: mock out statsd?
		Expect(err).ToNot(HaveOccurred())
		generator = &nginx.MockGenerator{}
		chker = new(checker.MockChecker)
		chker.GetVal = resources.ServiceCatalog{
			Services: []resources.Service{},
		}

		api = NewNGINX(NGINXConfig{
			Statsd:    statsdClient,
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
