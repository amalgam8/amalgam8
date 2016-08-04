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

package manager

import (
	"net/http"

	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/controller/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manager", func() {

	var (
		manager    Manager
		tenantInfo resources.TenantInfo
		id         string
		db         database.Tenant
	)

	Context("Manager", func() {

		BeforeEach(func() {
			db = database.NewTenant(database.NewMemoryCloudantDB())
			manager = NewManager(Config{
				Database: db,
			})

			id = "abcdef"
			tenantInfo = resources.TenantInfo{
				LoadBalance: "round_robin",
				Filters: resources.Filters{
					Rules:    []resources.Rule{},
					Versions: []resources.Version{},
				},
			}
		})

		It("nothing has been registered in database", func() {
			list, err := db.List(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(list).To(HaveLen(0))
		})

		It("registers an ID", func() {
			Expect(manager.Create(id, tenantInfo)).ToNot(HaveOccurred())
		})

		It("delete invalid id returns error", func() {
			Expect(manager.Delete(id)).To(HaveOccurred())
		})

		It("cannot get non-exisitant id", func() {
			_, err := manager.Get(id)
			Expect(err).To(HaveOccurred())
		})

		Context("entry has been added", func() {
			BeforeEach(func() {
				Expect(manager.Create(id, tenantInfo)).ToNot(HaveOccurred())
			})

			It("ID in database", func() {
				list, err := db.List(nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(list).To(HaveLen(1))
			})

			It("database entry has default fields", func() {
				entryFromDB, err := manager.Get(id)
				Expect(err).ToNot(HaveOccurred())

				Expect(entryFromDB.ProxyConfig.Filters.Rules).ToNot(BeNil())
				Expect(entryFromDB.ProxyConfig.Filters.Rules).To(HaveLen(0))

			})

			It("can deregister a valid ID", func() {
				Expect(manager.Delete(id)).ToNot(HaveOccurred())
			})

			Context("ID has been deregistered", func() {
				BeforeEach(func() {
					Expect(manager.Delete(id)).ToNot(HaveOccurred())
				})

				It("cannot get non-exisitant ID", func() {
					_, err := manager.Get(id)
					Expect(err).To(HaveOccurred())
				})

				It("no entries are in database", func() {
					list, err := db.List(nil)
					Expect(err).ToNot(HaveOccurred())
					Expect(list).To(HaveLen(0))
				})
			})

			It("updates config", func() {
				Expect(manager.Set(id, tenantInfo)).ToNot(HaveOccurred())
			})

			Context("config has been updated", func() {
				BeforeEach(func() {
					tenantInfo.Filters.Rules = []resources.Rule{
						resources.Rule{
							Destination:      "A",
							Source:           "B",
							AbortProbability: 0.3,
							Header:           "test",
							Pattern:          "test",
							ReturnCode:       http.StatusServiceUnavailable,
						},
						resources.Rule{
							Source:           "A",
							Destination:      "B",
							DelayProbability: 0.1,
							Header:           "test",
							Pattern:          "test",
							Delay:            0.2,
						},
					}
					Expect(manager.Set(id, tenantInfo)).ToNot(HaveOccurred())
				})

				It("database entry has correct values", func() {
					configFromDB, err := manager.Get(id)
					Expect(err).ToNot(HaveOccurred())
					Expect(configFromDB.ID).To(Equal(id))
					Expect(configFromDB.ProxyConfig.Filters.Rules).ToNot(BeNil())
					Expect(configFromDB.ProxyConfig.Filters.Rules).To(HaveLen(len(tenantInfo.Filters.Rules)))
					for i := range configFromDB.ProxyConfig.Filters.Rules {

						// TODO order may not be guaranteed?

						Expect(configFromDB.ProxyConfig.Filters.Rules[i].Source).To(Equal(tenantInfo.Filters.Rules[i].Source))
						Expect(configFromDB.ProxyConfig.Filters.Rules[i].Destination).To(Equal(tenantInfo.Filters.Rules[i].Destination))
						Expect(configFromDB.ProxyConfig.Filters.Rules[i].ReturnCode).To(Equal(tenantInfo.Filters.Rules[i].ReturnCode))
						Expect(configFromDB.ProxyConfig.Filters.Rules[i].Pattern).To(Equal(tenantInfo.Filters.Rules[i].Pattern))
						Expect(configFromDB.ProxyConfig.Filters.Rules[i].AbortProbability).To(Equal(tenantInfo.Filters.Rules[i].AbortProbability))

						Expect(configFromDB.ProxyConfig.Filters.Rules[i].Source).To(Equal(tenantInfo.Filters.Rules[i].Source))
						Expect(configFromDB.ProxyConfig.Filters.Rules[i].Destination).To(Equal(tenantInfo.Filters.Rules[i].Destination))
						Expect(configFromDB.ProxyConfig.Filters.Rules[i].DelayProbability).To(Equal(tenantInfo.Filters.Rules[i].DelayProbability))
						Expect(configFromDB.ProxyConfig.Filters.Rules[i].Pattern).To(Equal(tenantInfo.Filters.Rules[i].Pattern))
						Expect(configFromDB.ProxyConfig.Filters.Rules[i].Delay).To(Equal(tenantInfo.Filters.Rules[i].Delay))
					}
				})

				It("database entry updated, not re-created", func() {
					list, err := db.List(nil)
					Expect(err).ToNot(HaveOccurred())
					Expect(list).To(HaveLen(1))
				})
			})
		})
	})
})
