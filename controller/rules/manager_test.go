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

package rules

import (
	"errors"

	"github.com/amalgam8/amalgam8/pkg/api"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type MockValidator struct {
	Error error
}

func (m *MockValidator) Validate(r api.Rule) error {
	return m.Error
}

var _ = Describe("Manager", func() {

	var (
		validator api.Validator
		manager   Manager
		namespace string
	)

	BeforeEach(func() {
		validator = &MockValidator{}
		namespace = "test"
	})

	JustBeforeEach(func() {
		manager = NewMemoryManager(validator)
	})

	Describe("simple CRUD operations", func() {
		Describe("adding a rule", func() {
			var (
				rules    []api.Rule
				newRules NewRules
				err      error
			)

			JustBeforeEach(func() {
				newRules, err = manager.AddRules(namespace, rules)
			})

			Context("the rule is valid", func() {
				const Destination = "DestinationX"

				BeforeEach(func() {
					rules = []api.Rule{
						{
							Destination: Destination,
						},
					}
				})

				It("should not error", func() {
					Expect(err).ToNot(HaveOccurred())
				})

				It("should return an ID", func() {
					Expect(newRules.IDs).To(HaveLen(1))
					Expect(newRules.IDs[0]).ToNot(BeEmpty())
				})

				It("now contains the rule", func() {
					f := api.RuleFilter{
						IDs: newRules.IDs,
					}
					retrievedRules, err := manager.GetRules(namespace, f)
					Expect(err).ToNot(HaveOccurred())
					Expect(retrievedRules.Rules).To(HaveLen(1))
					Expect(retrievedRules.Rules[0].Destination).To(Equal(Destination))
				})

				It("updates the revision", func() {
					retrievedRules, err := manager.GetRules(namespace, api.RuleFilter{})
					Expect(err).ToNot(HaveOccurred())
					Expect(retrievedRules.Revision).To(Equal(int64(1)))
				})

				Describe("modifying a rule", func() {
					const NewDestination = "DestinationY"

					JustBeforeEach(func() {
						rules = []api.Rule{
							{
								ID:          newRules.IDs[0],
								Destination: NewDestination,
							},
						}
						err = manager.UpdateRules(namespace, rules)
					})

					Context("the modified rule is valid", func() {
						It("should not error", func() {
							Expect(err).ToNot(HaveOccurred())
						})

						Context("the modified rule is read", func() {
							var retrievedRules RetrievedRules

							JustBeforeEach(func() {
								f := api.RuleFilter{
									IDs: newRules.IDs,
								}
								retrievedRules, err = manager.GetRules(namespace, f)
							})

							It("should not error", func() {
								Expect(err).ToNot(HaveOccurred())
							})

							It("has updated the rule", func() {
								Expect(retrievedRules.Rules).To(HaveLen(1))
								Expect(retrievedRules.Rules[0].Destination).To(Equal(NewDestination))
							})

							It("updates the revision", func() {
								Expect(retrievedRules.Revision).To(Equal(int64(2)))
							})
						})
					})
				})

				Describe("deleting the rule", func() {
					JustBeforeEach(func() {
						filter := api.RuleFilter{
							IDs: newRules.IDs,
						}
						err = manager.DeleteRules(namespace, filter)
					})

					It("should not error", func() {
						Expect(err).ToNot(HaveOccurred())
					})

					Context("all rules are retrieved", func() {
						var retrievedRules RetrievedRules

						JustBeforeEach(func() {
							retrievedRules, err = manager.GetRules(namespace, api.RuleFilter{})
						})

						It("should not error", func() {
							Expect(err).ToNot(HaveOccurred())
						})

						It("no longer contains any rules", func() {
							Expect(retrievedRules.Rules).To(BeEmpty())
						})

						It("updates the revision", func() {
							Expect(retrievedRules.Revision).To(Equal(int64(2)))
						})
					})
				})
			})

			Context("the rule is invalid", func() {
				BeforeEach(func() {
					validator = &MockValidator{
						Error: errors.New("invalid rule"),
					}
				})

				It("should error", func() {
					Expect(err).To(HaveOccurred())
				})

				It("should not generate an ID", func() {
					Expect(newRules.IDs).To(BeEmpty())
				})

				Context("all rules are retrieved", func() {
					var retrievedRules RetrievedRules

					JustBeforeEach(func() {
						retrievedRules, err = manager.GetRules(namespace, api.RuleFilter{})
					})

					It("should not error", func() {
						Expect(err).ToNot(HaveOccurred())
					})

					It("should not contain any rules", func() {
						Expect(retrievedRules.Rules).To(BeEmpty())
					})

					It("should not update the revision", func() {
						Expect(retrievedRules.Revision).To(Equal(int64(0)))
					})
				})
			})
		})
	})
})
