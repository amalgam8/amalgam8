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

package nginx

import (
	"bytes"

	"time"

	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/controller/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NGINX", func() {

	var (
		gen        Generator
		writer     *bytes.Buffer
		db         database.Tenant
		lastUpdate *time.Time
		entry      resources.TenantEntry
		id         string
	)

	Context("NGINX", func() {

		BeforeEach(func() {
			id = "abcdef"
			writer = bytes.NewBuffer([]byte{})
			db = database.NewTenant(database.NewMemoryCloudantDB())
			entry = resources.TenantEntry{
				ProxyConfig: resources.ProxyConfig{
					Filters: resources.Filters{
						Rules:    []resources.Rule{},
						Versions: []resources.Version{},
					},
					Port:        6379,
					LoadBalance: "round_robin",
				},
				BasicEntry: resources.BasicEntry{
					ID: id,
				},
				ServiceCatalog: resources.ServiceCatalog{
					Services:   []resources.Service{},
					LastUpdate: time.Now(),
				},
			}

			var err error
			gen, err = NewGenerator(Config{
				Database: db,
				Path:     "nginx.conf.tmpl",
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("generates base config", func() {

			Expect(db.Create(entry)).ToNot(HaveOccurred())

			Expect(gen.Generate(writer, id, lastUpdate)).ToNot(HaveOccurred())

			Expect(len(writer.String())).ToNot(BeZero())
		})

		It("generates a valid NGINX conf", func() {
			// Setup input data
			services := []resources.Service{
				resources.Service{
					Name: "ServiceA",
					Endpoints: []resources.Endpoint{
						resources.Endpoint{
							Type:  "http",
							Value: "127.0.0.1:1234",
							Metadata: resources.MetaData{
								Version: "v2",
							},
						},
						resources.Endpoint{
							Type:  "http",
							Value: "127.0.0.5:1234",
							Metadata: resources.MetaData{
								Version: "v1",
							},
						},
						resources.Endpoint{
							Type:  "https",
							Value: "127.0.0.2:1234",
						},
						resources.Endpoint{
							Type:  "tcp",
							Value: "127.0.0.3:1234",
						},
					},
				},
				resources.Service{
					Name: "ServiceB",
					Endpoints: []resources.Endpoint{
						resources.Endpoint{
							Type:  "https",
							Value: "127.0.0.4:1234",
						},
					},
				},
				resources.Service{
					Name: "ServiceC",
					Endpoints: []resources.Endpoint{
						resources.Endpoint{
							Type:  "http",
							Value: "127.0.0.5:1234",
						},
					},
				},
			}
			entry.ServiceCatalog.Services = services

			entry.ProxyConfig.Filters.Rules = []resources.Rule{
				resources.Rule{
					Source:           "source",
					Destination:      "ServiceA",
					Delay:            0.3,
					DelayProbability: 0.9,
					ReturnCode:       501,
					AbortProbability: 0.1,
					Pattern:          "header_value",
					Header:           "header_name",
				},
			}
			entry.ProxyConfig.Filters.Versions = []resources.Version{
				resources.Version{
					Selectors: "{v2={weight=0.25}}",
					Service:   "ServiceA",
					Default:   "v1",
				},
			}

			// FIXME need to set aborts/delay to something and test it!
			//manager.GetVal.Filters

			Expect(db.Create(entry)).ToNot(HaveOccurred())

			// Generate the NGINX conf
			buf := new(bytes.Buffer)
			gen.Generate(buf, id, lastUpdate)
			conf := buf.String()

			// Verify the result...

			// Make sure endpoints are present and properly filtered (only HTTP)
			Expect(conf).To(ContainSubstring("127.0.0.1:1234"))    // HTTP
			Expect(conf).NotTo(ContainSubstring("127.0.0.2:1234")) // HTTPS
			Expect(conf).NotTo(ContainSubstring("127.0.0.3:1234")) // TCP
			Expect(conf).NotTo(ContainSubstring("127.0.0.4:1234")) // HTTPS

			// Ensure proxy was generated for ServiceA and ServiceC
			Expect(conf).To(ContainSubstring("location /ServiceA/"))
			Expect(conf).To(ContainSubstring("upstream ServiceA_v2"))
			Expect(conf).To(ContainSubstring("upstream ServiceA_v1"))
			Expect(conf).To(ContainSubstring("ngx.var.target = splib.get_target(\"ServiceA\", \"v1\", {v2={weight=0.25}})"))
			Expect(conf).To(ContainSubstring("ngx.sleep(0.3)"))
			Expect(conf).To(ContainSubstring("ngx.exit(501)"))

			Expect(conf).To(ContainSubstring("location /ServiceC/"))
			Expect(conf).To(ContainSubstring("upstream ServiceC_UNVERSIONED"))
			Expect(conf).To(ContainSubstring("ngx.var.target = splib.get_target(\"ServiceC\", \"UNVERSIONED\", nil)"))

			// Ensure that no proxy configuration was generated for service with only HTTPS endpoint
			Expect(conf).NotTo(ContainSubstring("proxy_pass http://ServiceB/"))
			Expect(conf).NotTo(ContainSubstring("location /ServiceB/"))
		})
	})
})
