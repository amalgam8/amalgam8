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

package checker

import (
	"github.com/amalgam8/controller/clients"
	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/controller/notification"
	"github.com/amalgam8/controller/proxyconfig"
	"github.com/amalgam8/controller/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Checker", func() {

	var (
		checker     Checker
		catalog     resources.ServiceCatalog
		id          string
		proxyConfig *proxyconfig.MockManager
		sdClient    *clients.MockRegistry
		db          database.Catalog
		cache       *notification.MockTenantProducerCache
	)

	Context("Checker", func() {

		BeforeEach(func() {
			db = database.NewCatalog(database.NewMemoryCloudantDB())
			proxyConfig = new(proxyconfig.MockManager)
			sdClient = new(clients.MockRegistry)
			cache = new(notification.MockTenantProducerCache)
			checker = New(Config{
				Database:      db,
				ProxyConfig:   proxyConfig,
				Registry:      sdClient,
				ProducerCache: cache,
			})

			id = "abcdef"
			catalog = resources.ServiceCatalog{
				BasicEntry: resources.BasicEntry{
					ID:  id,
					Rev: "rev",
					IV:  "iv",
				},
			}
		})

		It("nothing has been registered in database", func() {
			list, err := db.List(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(list).To(HaveLen(0))
		})

		It("registers an ID", func() {
			Expect(checker.Register(id)).ToNot(HaveOccurred())
		})

		It("delete invalid id returns error", func() {
			Expect(checker.Deregister("not_valid_id")).To(HaveOccurred())
		})

		It("cannot get non-exisitant id", func() {
			_, err := checker.Get(id)
			Expect(err).To(HaveOccurred())
		})

		It("can check works even if nothing is registered", func() {
			Expect(checker.Check(nil)).ToNot(HaveOccurred())
		})

		Context("ID has been registered", func() {
			BeforeEach(func() {
				Expect(checker.Register(id)).ToNot(HaveOccurred())
			})

			It("ID in database", func() {
				list, err := db.List(nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(list).To(HaveLen(1))
			})

			It("database entry has default fields", func() {
				catalogFromDB, err := checker.Get(id)
				Expect(err).ToNot(HaveOccurred())

				Expect(catalogFromDB.Services).ToNot(BeNil())
				Expect(catalogFromDB.Services).To(HaveLen(0))
			})

			It("can deregister a valid ID", func() {
				Expect(checker.Deregister(id)).ToNot(HaveOccurred())
			})

			Context("ID has been registered", func() {
				BeforeEach(func() {
					Expect(checker.Deregister(id)).ToNot(HaveOccurred())
				})

				It("cannot get non-exisitant ID", func() {
					_, err := checker.Get(id)
					Expect(err).To(HaveOccurred())
				})

				It("no entries are in database", func() {
					list, err := db.List(nil)
					Expect(err).ToNot(HaveOccurred())
					Expect(list).To(HaveLen(0))
				})
			})

			It("checks registered IDs", func() {
				Expect(checker.Check(nil)).ToNot(HaveOccurred())
			})

			Context("has checked registered IDs", func() {
				BeforeEach(func() {

					sdClient.GetInstancesVal = getBaseInstances()

					Expect(checker.Check(nil)).ToNot(HaveOccurred())
				})

				It("database entry has correct values", func() {
					catalogFromDB, err := checker.Get(id)
					Expect(err).ToNot(HaveOccurred())

					Expect(catalogFromDB.Services).ToNot(BeNil())
					Expect(catalogFromDB.Services).To(HaveLen(2))
					for i := range catalogFromDB.Services {
						Expect(catalogFromDB.Services[i].Name).To(Or(Equal("A"), Equal("B")))
						Expect(catalogFromDB.Services[i].Endpoints).ToNot(BeNil())

						Expect(catalogFromDB.Services[i].Endpoints).To(Or(HaveLen(1), HaveLen(2)))
					}

					Expect(catalogFromDB.ID).To(Equal(id))
					//					Expect(catalogFromDB.Rev).To(Equal(catalog.Rev))
					//					Expect(catalogFromDB.IV).To(Equal(catalog.IV))

				})
			})
		})
	})
})

func getBaseInstances() []clients.Instance {
	endpoint := clients.Endpoint{
		Type:  "http",
		Value: "A:9999",
	}
	inst := clients.Instance{
		ServiceName: "A",
		Endpoint:    endpoint,
	}
	endpoint2 := clients.Endpoint{
		Type:  "http",
		Value: "A:9988",
	}
	inst2 := clients.Instance{
		ServiceName: "A",
		Endpoint:    endpoint2,
	}

	endpoint3 := clients.Endpoint{
		Type:  "http",
		Value: "B:9999",
	}
	inst3 := clients.Instance{
		ServiceName: "B",
		Endpoint:    endpoint3,
	}

	var instances []clients.Instance
	instances = append(instances, inst)
	instances = append(instances, inst2)
	return append(instances, inst3)
}
