package nginx

import (
	"bytes"
	"github.com/amalgam8/controller/checker"
	"github.com/amalgam8/controller/proxyconfig"
	"github.com/amalgam8/controller/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NGINX", func() {

	var (
		gen     Generator
		manager *proxyconfig.MockManager
		chker   *checker.MockChecker
		writer  *bytes.Buffer
	)

	Context("NGINX", func() {

		BeforeEach(func() {
			writer = bytes.NewBuffer([]byte{})

			manager = new(proxyconfig.MockManager)
			manager.GetVal = resources.ProxyConfig{
				Filters: resources.Filters{
					Rules:    []resources.Rule{},
					Versions: []resources.Version{},
				},
				Port:        6379,
				LoadBalance: "round_robin",
			}

			chker = new(checker.MockChecker)
			chker.GetVal = resources.ServiceCatalog{
				Services: []resources.Service{},
			}

			var err error
			gen, err = NewGenerator(Config{
				Catalog:      chker,
				ProxyManager: manager,
				Path:         "nginx.conf.tmpl",
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("generates base config", func() {

			Expect(gen.Generate(writer, "abcdef")).ToNot(HaveOccurred())

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
			chker.GetVal.Services = services

			manager.GetVal.Filters.Rules = []resources.Rule{
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
			manager.GetVal.Filters.Versions = []resources.Version{
				resources.Version{
					Selectors: "{v2={weight=0.25}}",
					Service:   "ServiceA",
					Default:   "v1",
				},
			}

			// FIXME need to set aborts/delay to something and test it!
			//manager.GetVal.Filters

			// Generate the NGINX conf
			buf := new(bytes.Buffer)
			gen.Generate(buf, "abcdef")
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
