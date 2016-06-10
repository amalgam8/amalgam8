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

package proxyconfig

import (
	"net/http"

	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/controller/notification"
	"github.com/amalgam8/controller/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProxyConfigManager", func() {

	var (
		manager Manager
		config  resources.ProxyConfig
		id      string
		db      database.Rules
		cache   *notification.MockTenantProducerCache
	)

	Context("ProxyConfigManager", func() {

		BeforeEach(func() {
			db = database.NewRules(database.NewMemoryCloudantDB())
			cache = new(notification.MockTenantProducerCache)
			manager = NewManager(Config{
				Database:      db,
				ProducerCache: cache,
			})

			id = "abcdef"
			config = resources.ProxyConfig{
				BasicEntry: resources.BasicEntry{
					ID:  id,
					Rev: "rev",
					IV:  "iv",
				},
				Port:        6379,
				LoadBalance: "round_robin",
				Filters: resources.Filters{
					Rules:    []resources.Rule{},
					Versions: []resources.Version{},
				},
			}
		})

		It("nothing has been registered in database", func() {
			list, err := db.List()
			Expect(err).ToNot(HaveOccurred())
			Expect(list).To(HaveLen(0))
		})

		It("registers an ID", func() {
			Expect(manager.Set(config)).ToNot(HaveOccurred())
		})

		It("delete invalid id returns error", func() {
			Expect(manager.Delete(id)).To(HaveOccurred())
		})

		It("cannot get non-exisitant id", func() {
			_, err := manager.Get(id)
			Expect(err).To(HaveOccurred())
		})

		Context("config has been added", func() {
			BeforeEach(func() {
				Expect(manager.Set(config)).ToNot(HaveOccurred())
			})

			It("ID in database", func() {
				list, err := db.List()
				Expect(err).ToNot(HaveOccurred())
				Expect(list).To(HaveLen(1))
			})

			It("database entry has default fields", func() {
				configFromDB, err := manager.Get(id)
				Expect(err).ToNot(HaveOccurred())

				Expect(configFromDB.Port).To(Equal(config.Port))
				Expect(configFromDB.Filters.Rules).ToNot(BeNil())
				Expect(configFromDB.Filters.Rules).To(HaveLen(0))

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
					list, err := db.List()
					Expect(err).ToNot(HaveOccurred())
					Expect(list).To(HaveLen(0))
				})
			})

			It("updates config", func() {
				Expect(manager.Set(config)).ToNot(HaveOccurred())
			})

			Context("config has been updated", func() {
				BeforeEach(func() {
					config.Filters.Rules = []resources.Rule{
						resources.Rule{
							Destination:      "A",
							Source:           "B",
							AbortProbability: 0.3,
							Pattern:          "test",
							ReturnCode:       http.StatusServiceUnavailable,
						},
						resources.Rule{
							Source:           "A",
							Destination:      "B",
							DelayProbability: 0.1,
							Pattern:          "test",
							Delay:            0.2,
						},
					}
					Expect(manager.Set(config)).ToNot(HaveOccurred())
				})

				It("database entry has correct values", func() {
					configFromDB, err := manager.Get(id)
					Expect(err).ToNot(HaveOccurred())
					Expect(configFromDB.ID).To(Equal(id))
					Expect(configFromDB.Rev).ToNot(Equal(config.Rev))
					Expect(configFromDB.IV).To(Equal(config.IV))
					Expect(configFromDB.Port).To(Equal(config.Port))
					Expect(configFromDB.Filters.Rules).ToNot(BeNil())
					Expect(configFromDB.Filters.Rules).To(HaveLen(len(config.Filters.Rules)))
					for i := range configFromDB.Filters.Rules {

						// TODO order may not be guaranteed?

						Expect(configFromDB.Filters.Rules[i].Source).To(Equal(config.Filters.Rules[i].Source))
						Expect(configFromDB.Filters.Rules[i].Destination).To(Equal(config.Filters.Rules[i].Destination))
						Expect(configFromDB.Filters.Rules[i].ReturnCode).To(Equal(config.Filters.Rules[i].ReturnCode))
						Expect(configFromDB.Filters.Rules[i].Pattern).To(Equal(config.Filters.Rules[i].Pattern))
						Expect(configFromDB.Filters.Rules[i].AbortProbability).To(Equal(config.Filters.Rules[i].AbortProbability))

						Expect(configFromDB.Filters.Rules[i].Source).To(Equal(config.Filters.Rules[i].Source))
						Expect(configFromDB.Filters.Rules[i].Destination).To(Equal(config.Filters.Rules[i].Destination))
						Expect(configFromDB.Filters.Rules[i].DelayProbability).To(Equal(config.Filters.Rules[i].DelayProbability))
						Expect(configFromDB.Filters.Rules[i].Pattern).To(Equal(config.Filters.Rules[i].Pattern))
						Expect(configFromDB.Filters.Rules[i].Delay).To(Equal(config.Filters.Rules[i].Delay))
					}
				})

				It("database entry updated, not re-created", func() {
					list, err := db.List()
					Expect(err).ToNot(HaveOccurred())
					Expect(list).To(HaveLen(1))
				})
			})
		})
	})
})
