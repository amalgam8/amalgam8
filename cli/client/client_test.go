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

package client_test

import (
	"bytes"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	. "github.com/amalgam8/amalgam8/cli/client"
	"net/http"
)

var _ = Describe("Amalgam8 Client", func() {
	fmt.Println()
	var server *ghttp.Server
	var client Client

	BeforeEach(func() {

	})

	Describe("The amalgam8 client", func() {

		BeforeEach(func() {
			server = ghttp.NewServer()
			client = NewClient(server.URL(), "", nil)

		})

		Describe("set a new HTTP Client", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/instances"),
						ghttp.RespondWith(http.StatusOK, `{}`),
					),
				)
			})

			It("should not fail", func() {
				client.SetRegistryURL(server.URL())
				client.SetHTTPClient(&http.Client{
					Transport: nil,
				})
				err := client.GET("/instances", true, nil, nil)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("set a new Registry Token", func() {
			It("should not fail", func() {
				client.SetRegistryToken("")
			})
		})

		Describe("Token", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/"),
						ghttp.RespondWith(http.StatusOK, `{}`),
						ghttp.VerifyHeaderKV("Authorization", "Bearer Test"),
					),
				)
			})

			It("should not fail", func() {
				client.SetRegistryToken("Test")
				err := client.GET("/", false, nil, nil)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Describe("Headers", func() {
			BeforeEach(func() {
				header := http.Header{}
				header.Add("Name", "Test")
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/"),
						ghttp.VerifyHeader(header),
						ghttp.RespondWith(http.StatusOK, `{}`),
					),
				)
			})

			It("should not fail", func() {
				header := http.Header{}
				header.Add("Name", "Test")
				err := client.GET("/", false, header, nil)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Describe("GET", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/services"),
						ghttp.RespondWith(http.StatusOK, `{"name":"test", "version":1}`),
					),
				)
			})

			It("should not fail if object has been provided to parse the result", func() {
				err := client.GET("/services", false, nil, nil)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should parse the result correctly", func() {
				test := &struct {
					Name    string
					Version int
				}{}
				err := client.GET("/services", false, nil, test)
				Expect(err).ToNot(HaveOccurred())
				Expect(test.Name).To(Equal("test"))
				Expect(test.Version).To(Equal(1))
			})
		})

		Describe("POST", func() {
			Describe("POST nil or empty body", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/services"),
							ghttp.VerifyBody([]byte("")),
							ghttp.VerifyContentType("application/json"),
							ghttp.RespondWith(http.StatusOK, `{}`),
						),
					)
				})

				It("should not fail", func() {
					err := client.POST("/services", nil, true, nil, nil)
					Expect(err).ToNot(HaveOccurred())
				})

				It("should not fail", func() {
					body := bytes.NewReader([]byte(""))
					err := client.POST("/services", body, true, nil, nil)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Describe("POST JSON body", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/services"),
							ghttp.VerifyBody([]byte(`{"name":"test"}`)),
							ghttp.VerifyContentType("application/json"),
							ghttp.RespondWith(http.StatusOK, `{}`),
						),
					)
				})
				It("should not fail", func() {
					body := bytes.NewReader([]byte(`{"name":"test"}`))
					err := client.POST("/services", body, true, nil, nil)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Describe("PUT", func() {
			Describe("PUT nil or empty body", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/services"),
							ghttp.VerifyBody([]byte("")),
							ghttp.VerifyContentType("application/json"),
							ghttp.RespondWith(http.StatusOK, `{}`),
						),
					)
				})

				It("should not fail", func() {
					err := client.PUT("/services", nil, true, nil, nil)
					Expect(err).ToNot(HaveOccurred())
				})

				It("should not fail", func() {
					body := bytes.NewReader([]byte(""))
					err := client.PUT("/services", body, true, nil, nil)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Describe("PUT JSON body", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/services"),
							ghttp.VerifyBody([]byte(`{"name":"test"}`)),
							ghttp.VerifyContentType("application/json"),
							ghttp.RespondWith(http.StatusOK, `{}`),
						),
					)
				})
				It("should not fail", func() {
					body := bytes.NewReader([]byte(`{"name":"test"}`))
					err := client.PUT("/services", body, true, nil, nil)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Describe("DELETE", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/services"),
						ghttp.RespondWith(http.StatusOK, `{}`),
					),
				)
			})

			It("should not fail", func() {
				err := client.DELETE("/services", false, nil, nil)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Describe("set a new Registry URL", func() {
			It("should fail", func() {

				client.SetRegistryURL("")
				err := client.DELETE("/test", false, nil, nil)
				Expect(err).To(HaveOccurred())
			})
		})

		AfterEach(func() {
			//shut down the server between tests
			server.Close()
		})
	})

})
