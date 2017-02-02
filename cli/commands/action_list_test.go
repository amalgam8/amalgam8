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

var _ = Describe("action-list", func() {
	fmt.Println()
	utils.LoadLocales("../locales")
	T := utils.Language(common.DefaultLanguage)
	var cmd *cmds.ActionListCommand
	var app *cli.App
	var server *ghttp.Server
	response := make(map[string][]byte)

	BeforeEach(func() {
		app = cli.NewApp()
		app.Name = T("app_name")
		app.Usage = T("app_usage")
		app.Version = T("app_version")
		app.Flags = config.GlobalFlags()
		server = ghttp.NewServer()
		term := terminal.NewUI(os.Stdin, os.Stdout)
		cmd = cmds.NewActionListCommand(term)
		app.Commands = []cli.Command{cmd.GetMetadata()}
		app.Before = config.Before
		app.Action = config.DefaultAction
		app.OnUsageError = config.OnUsageError

		response["reviews"] = []byte(`
			"reviews": [
				{
					"id": "2d381a94-1796-45c3-a1d8-3965051b61b1",
					"priority": 10,
					"destination": "reviews",
					"match": {
						"source": {
							"name": "productpage",
							"tags": [
								"v1"
							]
						},
						"headers": {
							"Cookie": ".*?user=jason"
						}
					},
					"actions": [
						{
							"action": "trace",
							"tags": [
								"v2"
							],
							"log_key": "gremlin_recipe_id",
							"log_value": "9f0ea878-a1f4-11e6-b410-6c40089c9f90"
						},
						{
							"action": "abort",
							"tags": [ "v1" ],
							"return_code": 400
						}
					]
				}
			]`)

		response["ratings"] = []byte(`
			"ratings": [
      {
        "id": "454a8fb0-d260-4832-8007-5b5344c03c1f",
        "priority": 10,
        "destination": "ratings",
        "match": {
          "source": {
            "name": "reviews",
            "tags": [
              "v2"
            ]
          },
          "headers": {
            "Cookie": ".*?user=jason"
          }
        },
        "actions": [
          {
            "action": "trace",
            "tags": [
              "v1"
            ],
            "log_key": "gremlin_recipe_id",
            "log_value": "9f0ea878-a1f4-11e6-b410-6c40089c9f90"
          },
          {
            "action": "delay",
            "duration": 7,
            "probability": 1,
            "tags": [
              "v1"
            ]
          }
        ]
      },
      {
        "id": "c2d98e32-8fd0-4e0d-a363-8adff99b0692",
        "priority": 10,
        "destination": "ratings",
        "match": {
          "source": {
            "name": "reviews",
            "tags": [
              "v2"
            ]
          },
          "headers": {
            "Cookie": ".*?user=jason"
          }
        },
        "actions": [
          {
            "action": "delay",
            "duration": 7,
            "probability": 1,
            "tags": [
              "v1"
            ]
          }
        ]
      },
      {
        "id": "dc8b5ffe-d50c-4c0e-84d4-bd21be20c5a8",
        "priority": 20,
        "destination": "ratings",
        "match": {
          "source": {
            "name": "reviews",
            "tags": [
              "v2"
            ]
          },
          "headers": {
            "Cookie": ".*?user=jason"
          }
        },
        "actions": [
          {
            "action": "delay",
            "duration": 7,
            "probability": 1,
            "tags": [
              "v1"
            ]
          }
        ]
      }
    ]`)

		allActions := fmt.Sprintf("{ \"services\": {%s,%s}}", response["reviews"], response["ratings"])
		response["all_actions"] = []byte(allActions)
	})

	Describe("List of Actions", func() {

		Describe("Validate Registry URL", func() {

			JustBeforeEach(func() {
				app.Writer = bytes.NewBufferString("")
			})

			It("should exit with ErrControllerURLInvalid error", func() {
				err := app.Run([]string{"app", "--controller_url=123", "action-list"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrControllerURLInvalid.Error()))
			})

			It("should exit with ErrControllerURLInvalid error", func() {
				err := app.Run([]string{"app", "--controller_url=", "action-list"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrControllerURLNotFound.Error()))
			})

			It("should error", func() {
				err := app.Run([]string{"app", "--controller_url=http://localhost", "--x"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(app.Name))
			})

		})

		Describe("On usage error: [action-list bad]", func() {

			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v1/rules/actions"),
						ghttp.RespondWith(http.StatusOK, response["all_actions"]),
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
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "action-list", "--bad"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(cmd.GetMetadata().Usage))
			})

		})

		Describe("Action List: [action-list]", func() {

			BeforeEach(func() {

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v1/rules/actions"),
						ghttp.RespondWith(http.StatusOK, response["all_actions"]),
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

			It("should print table", func() {
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "action-list"})
				Expect(err).ToNot(HaveOccurred())
				// TODO: Validate output
			})

			It("should print a JSON", func() {
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "action-list", "-o=json"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`"destination": "ratings"`))

			})

			It("should print a YAML", func() {
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "action-list", "-o=yaml"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`destination: ratings`))

			})

		})

	})
})
