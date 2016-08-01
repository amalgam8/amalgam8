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
	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/registry/client"
	. "github.com/onsi/ginkgo"
	//	. "github.com/onsi/gomega"
)

var _ = Describe("Checker", func() {

	var (
		checker   Checker
		id        string
		db        database.Tenant
		factory   *MockRegistryFactory
		regClient *MockRegistryClient
	)

	Context("Checker", func() {

		BeforeEach(func() {
			regClient = &MockRegistryClient{}
			factory = &MockRegistryFactory{
				RegClient: regClient,
			}
			db = database.NewTenant(database.NewMemoryCloudantDB())
			checker = New(Config{})

			id = "abcdef"
		})

		//It("can check works even if nothing is registered", func() {
		//	Expect(checker.Check(nil)).ToNot(HaveOccurred())
		//})
		//
		//Context("ID has been registered", func() {
		//	BeforeEach(func() {
		//		db.Create(resources.TenantEntry{
		//			BasicEntry: resources.BasicEntry{
		//				ID: id,
		//			},
		//			ServiceCatalog: resources.ServiceCatalog{
		//				Services:   []resources.Service{},
		//				LastUpdate: time.Now(),
		//			},
		//		})
		//	})
		//
		//	It("checks registered IDs", func() {
		//		Expect(checker.Check(nil)).ToNot(HaveOccurred())
		//	})
		//
		//	Context("has checked registered IDs", func() {
		//		BeforeEach(func() {
		//
		//			regClient.ListInstancesVal = getBaseInstances()
		//
		//			Expect(checker.Check(nil)).ToNot(HaveOccurred())
		//		})
		//
		//		It("database entry has correct values", func() {
		//			entry, err := db.Read(id)
		//			Expect(err).ToNot(HaveOccurred())
		//
		//			Expect(entry.ServiceCatalog.Services).ToNot(BeNil())
		//			Expect(entry.ServiceCatalog.Services).To(HaveLen(2))
		//			for i := range entry.ServiceCatalog.Services {
		//				Expect(entry.ServiceCatalog.Services[i].Name).To(Or(Equal("A"), Equal("B")))
		//				Expect(entry.ServiceCatalog.Services[i].Endpoints).ToNot(BeNil())
		//
		//				Expect(entry.ServiceCatalog.Services[i].Endpoints).To(Or(HaveLen(1), HaveLen(2)))
		//			}
		//
		//			Expect(entry.ID).To(Equal(id))
		//			//					Expect(catalogFromDB.Rev).To(Equal(catalog.Rev))
		//			//					Expect(catalogFromDB.IV).To(Equal(catalog.IV))
		//
		//		})
		//	})
		//})
	})
})

func getBaseInstances() []*client.ServiceInstance {
	endpoint := client.ServiceEndpoint{
		Type:  "http",
		Value: "A:9999",
	}
	inst := &client.ServiceInstance{
		ServiceName: "A",
		Endpoint:    endpoint,
	}
	endpoint2 := client.ServiceEndpoint{
		Type:  "http",
		Value: "A:9988",
	}
	inst2 := &client.ServiceInstance{
		ServiceName: "A",
		Endpoint:    endpoint2,
	}

	endpoint3 := client.ServiceEndpoint{
		Type:  "http",
		Value: "B:9999",
	}
	inst3 := &client.ServiceInstance{
		ServiceName: "B",
		Endpoint:    endpoint3,
	}

	var instances []*client.ServiceInstance
	instances = append(instances, inst)
	instances = append(instances, inst2)
	return append(instances, inst3)
}
