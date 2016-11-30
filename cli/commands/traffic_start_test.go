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

package commands_test

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	cmds "github.com/amalgam8/amalgam8/cli/commands"
	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/config"
	"github.com/amalgam8/amalgam8/cli/terminal"
	"github.com/amalgam8/amalgam8/cli/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/urfave/cli"
)

var _ = Describe("traffic-start", func() {
	fmt.Println()
	utils.LoadLocales("../locales")
	T := utils.Language(common.DefaultLanguage)
	var cmd *cmds.TrafficStartCommand
	var app *cli.App
	var server *ghttp.Server

	BeforeEach(func() {
		app = cli.NewApp()
		app.Name = T("app_name")
		app.Usage = T("app_usage")
		app.Version = T("app_version")
		app.Flags = config.GlobalFlags()
		server = ghttp.NewServer()
		term := terminal.NewUI(os.Stdin, os.Stdout)
		cmd = cmds.NewTrafficStartCommand(term)
		app.Commands = []cli.Command{cmd.GetMetadata()}
		app.Before = config.Before
		app.Action = config.DefaultAction
		app.OnUsageError = config.OnUsageError
	})

	Describe("Traffic Start", func() {

		Describe("Validate Controller URL", func() {

			JustBeforeEach(func() {
				app.Writer = bytes.NewBufferString("")
			})

			It("should exit with ErrControllerURLInvalid error", func() {
				err := app.Run([]string{"app", "--controller_url=123", "--registry_url=http://localhost", "traffic-start"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrControllerURLInvalid.Error()))
			})

			It("should exit with ErrControllerURLNotFound error", func() {
				err := app.Run([]string{"app", "--controller_url=", "--registry_url=http://localhost", "traffic-start"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrControllerURLNotFound.Error()))
			})

			It("should error", func() {
				err := app.Run([]string{"app", "--controller_url=http://localhost", "--registry_url=http://localhost", "--x"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(app.Name))
			})

		})

		Describe("Validate Registry URL", func() {

			JustBeforeEach(func() {
				app.Writer = bytes.NewBufferString("")
			})

			It("should exit with ErrRegistryURLInvalid error", func() {
				err := app.Run([]string{"app", "--registry_url=123", "--controller_url=http://localhost", "traffic-start"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrRegistryURLInvalid.Error()))
			})

			It("should exit with ErrRegistryURLNotFound error", func() {
				err := app.Run([]string{"app", "--registry_url=", "--controller_url=http://localhost", "traffic-start"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrRegistryURLNotFound.Error()))
			})

			It("should error", func() {
				err := app.Run([]string{"app", "--registry_url=http://localhost", "--controller_url=http://localhost", "--x"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(app.Name))
			})

		})

		Describe("On usage error: [traffic-start bad]", func() {

			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v1/rules/routes/test"),
						ghttp.RespondWith(http.StatusOK, response["no_rules"]),
					),
				)
			})

			AfterEach(func() {
				//shut down the server between tests
				server.Close()
			})

			JustBeforeEach(func() {
				app.Writer = bytes.NewBufferString("")
			})

			It("should print the command help", func() {
				app.Writer = bytes.NewBufferString("")
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "--registry_url=" + server.URL(), "traffic-start", "--bad"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(cmd.GetMetadata().Usage))
			})

			It("should print the command help", func() {
				app.Writer = bytes.NewBufferString("")
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "--registry_url=" + server.URL(), "traffic-start", "-service=test", "-version"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(cmd.GetMetadata().Usage))
			})

			It("should print the command help", func() {
				app.Writer = bytes.NewBufferString("")
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "--registry_url=" + server.URL(), "traffic-start", "-service", "-version=v1"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(cmd.GetMetadata().Usage))
			})

			It("should print the command help", func() {
				app.Writer = bytes.NewBufferString("")
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "--registry_url=" + server.URL(), "traffic-start", "-amount"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(cmd.GetMetadata().Usage))
			})

		})

		Describe("Start Traffic: [traffic-start]", func() {

			Describe("When -amount is out of range", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v1/rules/routes/test"),
							ghttp.RespondWith(http.StatusOK, response["no_rules"]),
						),
					)
				})

				AfterEach(func() {
					//shut down the server between tests
					server.Close()
				})

				JustBeforeEach(func() {
					app.Writer = bytes.NewBufferString("")
				})

				It("should print ErrIncorrectAmountRange when range is incorrect", func() {
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--controller_url=" + server.URL(), "--registry_url=" + server.URL(), "traffic-start", "-service=test", "-version=v1", "-amount=200"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrIncorrectAmountRange.Error()))
				})
			})

			Describe("When the service has no rules", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v1/rules/routes/test"),
							ghttp.RespondWith(http.StatusOK, response["no_rules"]),
						),
					)
				})

				AfterEach(func() {
					//shut down the server between tests
					server.Close()
				})

				JustBeforeEach(func() {
					app.Writer = bytes.NewBufferString("")
				})

				It("should print ErrNotRulesFoundForService", func() {
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--controller_url=" + server.URL(), "--registry_url=" + server.URL(), "traffic-start", "-service=test", "-version=v1", "-amount=50"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrNotRulesFoundForService.Error()))
				})
			})

			Describe("When traffic has been already started", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v1/rules/routes/reviews"),
							ghttp.RespondWith(http.StatusOK, response["traffic_started"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v1/services/reviews"),
							ghttp.RespondWith(http.StatusOK, response["reviews"]),
						),
					)
				})

				AfterEach(func() {
					//shut down the server between tests
					server.Close()
				})

				JustBeforeEach(func() {
					app.Writer = bytes.NewBufferString("")
				})

				It("should print 'traffic is already being split'", func() {
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--controller_url=" + server.URL(), "--registry_url=" + server.URL(), "traffic-start", "-service=reviews", "-version=v1", "-amount=50"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("traffic is already being split"))
				})
			})

			Describe("When weight is not zero", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v1/rules/routes/reviews"),
							ghttp.RespondWith(http.StatusOK, response["weight_not_zero"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v1/services/reviews"),
							ghttp.RespondWith(http.StatusOK, response["reviews"]),
						),
					)
				})

				AfterEach(func() {
					//shut down the server between tests
					server.Close()
				})

				JustBeforeEach(func() {
					app.Writer = bytes.NewBufferString("")
				})

				It("should print 'traffic is already being split'", func() {
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--controller_url=" + server.URL(), "--registry_url=" + server.URL(), "traffic-start", "-service=reviews", "-version=v1", "-amount=50"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("traffic is already being split"))
				})
			})

			Describe("When service is not receiving", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v1/rules/routes/reviews"),
							ghttp.RespondWith(http.StatusOK, response["traffic_stopped"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v1/services/reviews"),
							ghttp.RespondWith(http.StatusOK, response["inactive"]),
						),
					)
				})

				AfterEach(func() {
					//shut down the server between tests
					server.Close()
				})

				JustBeforeEach(func() {
					app.Writer = bytes.NewBufferString("")
				})

				It("should print 'service is not currently receiving traffic'", func() {
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--debug", "--controller_url=" + server.URL(), "--registry_url=" + server.URL(), "traffic-start", "-service=reviews", "-version=v1", "-amount=50"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("is not currently receiving traffic"))
				})
			})

			Describe("When service has not active instance", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v1/rules/routes/reviews"),
							ghttp.RespondWith(http.StatusOK, response["traffic_stopped"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v1/services/reviews"),
							ghttp.RespondWith(http.StatusOK, response["reviews"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/v1/rules"),
						),
					)
				})

				AfterEach(func() {
					//shut down the server between tests
					server.Close()
				})

				JustBeforeEach(func() {
					app.Writer = bytes.NewBufferString("")
				})

				It("should print 'service does not have active instances'", func() {
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--debug", "--controller_url=" + server.URL(), "--registry_url=" + server.URL(), "traffic-start", "-service=reviews", "-version=v4", "-amount=50"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("does not have active instances"))
				})
			})

			Describe("When transfer complete for diverting traffic", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v1/rules/routes/reviews"),
							ghttp.RespondWith(http.StatusOK, response["traffic_stopped"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v1/services/reviews"),
							ghttp.RespondWith(http.StatusOK, response["reviews"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/v1/rules"),
						),
					)
				})

				AfterEach(func() {
					//shut down the server between tests
					server.Close()
				})

				JustBeforeEach(func() {
					app.Writer = bytes.NewBufferString("")
				})

				It("should print 'diverting 50% of traffic'", func() {
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--debug", "--controller_url=" + server.URL(), "--registry_url=" + server.URL(), "traffic-start", "-service=reviews", "-version=v2", "-amount=50"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("\"reviews\": diverting 50% of traffic from \"v1\" to \"v2\""))
				})
			})

			Describe("When transfer complete for sending traffic", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v1/rules/routes/reviews"),
							ghttp.RespondWith(http.StatusOK, response["traffic_stopped"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v1/services/reviews"),
							ghttp.RespondWith(http.StatusOK, response["reviews"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/v1/rules"),
						),
					)
				})

				AfterEach(func() {
					//shut down the server between tests
					server.Close()
				})

				JustBeforeEach(func() {
					app.Writer = bytes.NewBufferString("")
				})

				It("should print 'service does not have active instances'", func() {
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--debug", "--controller_url=" + server.URL(), "--registry_url=" + server.URL(), "traffic-start", "-service=reviews", "-version=v2", "-amount=100"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("\"reviews\": sending 100% of traffic to \"v2\""))
				})
			})

		})

	})
})
