package api

import (
	"errors"
	"github.com/amalgam8/controller/checker"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/cactus/go-statsd-client/statsd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("NGINX API", func() {
	var (
		api *Poll
		ch  *checker.MockChecker
		h   http.Handler
	)

	BeforeEach(func() {
		statsdClient, err := statsd.NewClient("", "") // TODO: mock out statsd?
		Expect(err).ToNot(HaveOccurred())
		ch = &checker.MockChecker{}

		api = NewPoll(statsdClient, ch)

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
