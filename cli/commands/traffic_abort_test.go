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

var _ = Describe("traffic-abort", func() {
	fmt.Println()
	utils.LoadLocales("../locales")
	T := utils.Language(common.DefaultLanguage)
	var cmd *cmds.TrafficAbortCommand
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
		cmd = cmds.NewTrafficAbortCommand(term)
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
				err := app.Run([]string{"app", "--controller_url=123", "traffic-abort"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrControllerURLInvalid.Error()))
			})

			It("should exit with ErrControllerURLNotFound error", func() {
				err := app.Run([]string{"app", "--controller_url=", "traffic-abort"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrControllerURLNotFound.Error()))
			})

			It("should error", func() {
				err := app.Run([]string{"app", "--controller_url=http://localhost", "--x"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(app.Name))
			})

		})

		Describe("On usage error: [traffic-abort bad]", func() {

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
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "traffic-abort", "--bad"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(cmd.GetMetadata().Usage))
			})

			It("should print the command help", func() {
				app.Writer = bytes.NewBufferString("")
				err := app.Run([]string{"app", "--debug", "--controller_url=" + server.URL(), "traffic-abort", "-service"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(cmd.GetMetadata().Usage))
			})

		})

		Describe("Start Traffic: [traffic-abort]", func() {

			Describe("When the service has no rules", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v1/rules/routes", "destination=test"),
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
					err := app.Run([]string{"app", "--controller_url=" + server.URL(), "traffic-abort", "-service=test"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrNotRulesFoundForService.Error()))
				})
			})

			Describe("When the service has 2 or more rules", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v1/rules/routes", "destination=reviews"),
							ghttp.RespondWith(http.StatusOK, response["reviews_two_rules"]),
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
					err := app.Run([]string{"app", "--controller_url=" + server.URL(), "traffic-abort", "-service=reviews"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrInvalidStateForTrafficStep.Error()))
				})
			})

			Describe("When traffic has not been started", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v1/rules/routes", "destination=reviews"),
							ghttp.RespondWith(http.StatusOK, response["reviews_traffic_stopped"]),
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
					err := app.Run([]string{"app", "--controller_url=" + server.URL(), "traffic-abort", "-service=reviews"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrInvalidStateForTrafficStep.Error()))
				})
			})

			Describe("When transfer aborted", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v1/rules/routes", "destination=reviews"),
							ghttp.RespondWith(http.StatusOK, response["reviews_traffic_started"]),
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
					err := app.Run([]string{"app", "--debug", "--controller_url=" + server.URL(), "traffic-abort", "-service=reviews"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("Transfer aborted for \"reviews\": all traffic reverted to \"v1\""))
				})
			})

		})

	})
})
