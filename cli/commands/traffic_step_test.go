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

var _ = Describe("traffic-step", func() {
	fmt.Println()
	utils.LoadLocales("../locales")
	T := utils.Language(common.DefaultLanguage)
	var cmd *cmds.TrafficStepCommand
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
		cmd = cmds.NewTrafficStepCommand(term)
		app.Commands = []cli.Command{cmd.GetMetadata()}
		app.Before = config.Before
		app.Action = config.DefaultAction
		app.OnUsageError = config.OnUsageError
	})

	Describe("Traffic Step", func() {

		Describe("Validate Controller URL", func() {

			JustBeforeEach(func() {
				app.Writer = bytes.NewBufferString("")
			})

			It("should exit with ErrControllerURLInvalid error", func() {
				err := app.Run([]string{"app", "--controller_url=123", "traffic-step"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrControllerURLInvalid.Error()))
			})

			It("should exit with ErrControllerURLNotFound error", func() {
				err := app.Run([]string{"app", "--controller_url=", "traffic-step"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrControllerURLNotFound.Error()))
			})

			It("should error", func() {
				err := app.Run([]string{"app", "--controller_url=http://localhost", "--x"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(app.Name))
			})

		})

		Describe("On usage error: [traffic-step bad]", func() {

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
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "traffic-step", "--bad"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(cmd.GetMetadata().Usage))
			})

			It("should print the command help", func() {
				app.Writer = bytes.NewBufferString("")
				err := app.Run([]string{"app", "--debug", "--controller_url=" + server.URL(), "traffic-step", "-service"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(cmd.GetMetadata().Usage))
			})

		})

		Describe("Start Traffic: [traffic-step]", func() {

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
					err := app.Run([]string{"app", "--controller_url=" + server.URL(), "traffic-step", "-service=test", "-amount=200"})
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
					err := app.Run([]string{"app", "--controller_url=" + server.URL(), "traffic-step", "-service=test", "-amount=50"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrNotRulesFoundForService.Error()))
				})
			})

			Describe("When the service has 2 or more rules", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v1/rules/routes/reviews"),
							ghttp.RespondWith(http.StatusOK, response["two_rules"]),
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

				It("should print ErrInvalidStateForTrafficStep", func() {
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--controller_url=" + server.URL(), "traffic-step", "-service=reviews", "-amount=50"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrInvalidStateForTrafficStep.Error()))
				})
			})

			Describe("When traffic has not been started", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v1/rules/routes/reviews"),
							ghttp.RespondWith(http.StatusOK, response["traffic_stopped"]),
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

				It("should print ErrInvalidStateForTrafficStep", func() {
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--controller_url=" + server.URL(), "traffic-step", "-service=reviews", "-amount=50"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrInvalidStateForTrafficStep.Error()))
				})
			})

			Describe("When transfer complete for diverting traffic", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v1/rules/routes/reviews"),
							ghttp.RespondWith(http.StatusOK, response["traffic_started"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/v1/rules"),
							ghttp.RespondWith(http.StatusOK, nil),
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
					err := app.Run([]string{"app", "--debug", "--controller_url=" + server.URL(), "traffic-step", "-service=reviews", "-amount=50"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("\"reviews\": diverting 50% of traffic from \"v1\" to \"v2\""))
				})
			})

			Describe("When transfer complete for sending traffic", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v1/rules/routes/reviews"),
							ghttp.RespondWith(http.StatusOK, response["traffic_started"]),
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
					err := app.Run([]string{"app", "--debug", "--controller_url=" + server.URL(), "traffic-step", "-service=reviews", "-amount=100"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("\"reviews\": sending 100% of traffic to \"v2\""))
				})
			})

		})

	})
})
