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
	"time"

	"github.com/amalgam8/controller/clients"
	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/controller/nginx"
	"github.com/amalgam8/controller/notification"
	"github.com/amalgam8/controller/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Checker", func() {

	var (
		checker  Checker
		id       string
		sdClient *clients.MockRegistry
		db       database.Tenant
		cache    *notification.MockTenantProducerCache
		n        *nginx.MockGenerator
	)

	Context("Checker", func() {

		BeforeEach(func() {
			db = database.NewTenant(database.NewMemoryCloudantDB())
			sdClient = new(clients.MockRegistry)
			cache = new(notification.MockTenantProducerCache)
			n = &nginx.MockGenerator{}
			checker = New(Config{
				Database:      db,
				Registry:      sdClient,
				ProducerCache: cache,
				Generator:     n,
			})

			id = "abcdef"
		})

		It("can check works even if nothing is registered", func() {
			Expect(checker.Check(nil)).ToNot(HaveOccurred())
		})

		Context("ID has been registered", func() {
			BeforeEach(func() {
				db.Create(resources.TenantEntry{
					BasicEntry: resources.BasicEntry{
						ID: id,
					},
					ServiceCatalog: resources.ServiceCatalog{
						Services:   []resources.Service{},
						LastUpdate: time.Now(),
					},
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
					entry, err := db.Read(id)
					Expect(err).ToNot(HaveOccurred())

					Expect(entry.ServiceCatalog.Services).ToNot(BeNil())
					Expect(entry.ServiceCatalog.Services).To(HaveLen(2))
					for i := range entry.ServiceCatalog.Services {
						Expect(entry.ServiceCatalog.Services[i].Name).To(Or(Equal("A"), Equal("B")))
						Expect(entry.ServiceCatalog.Services[i].Endpoints).ToNot(BeNil())

						Expect(entry.ServiceCatalog.Services[i].Endpoints).To(Or(HaveLen(1), HaveLen(2)))
					}

					Expect(entry.ID).To(Equal(id))
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
