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

var _ = Describe("route-list", func() {
	fmt.Println()
	utils.LoadLocales("../locales")
	T := utils.Language(common.DefaultLanguage)
	var cmd *cmds.RouteListCommand
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
		cmd = cmds.NewRouteListCommand(term)
		app.Commands = []cli.Command{cmd.GetMetadata()}
		app.Before = config.Before
		app.Action = config.DefaultAction
		app.OnUsageError = config.OnUsageError

		response["reviews"] = []byte(`
			"reviews": [
      {
        "id": "e3f2346a-9460-4a1c-b202-bd183da17d59",
        "priority": 2,
        "destination": "reviews",
        "match": {
          "headers": {
            "Foo": "bar"
          }
        },
        "route": {
          "backends": [
            {
              "tags": [
                "v2"
              ]
            }
          ]
        }
      },
      {
        "id": "08e8c2f7-aa57-486f-8438-233595d6478b",
        "priority": 1,
        "destination": "reviews",
        "route": {
          "backends": [
            {
              "weight": 0.5,
              "tags": [
                "v3"
              ]
            },
            {
              "tags": [
                "v2"
              ]
            }
          ]
        }
      }
    ]`)

		response["ratings"] = []byte(`
			"ratings": [
      {
        "id": "7a497641-2485-480f-9011-c3bf3ec19e46",
        "priority": 1,
        "destination": "ratings",
        "route": {
          "backends": [
            {
              "tags": [
                "v2"
              ]
            }
          ]
        }
      },
      {
        "id": "ccb0421a-9933-4bf6-a4b8-b9c7905a4fe7",
        "priority": 2,
        "destination": "ratings",
        "match": {
          "headers": {
            "Cookie": ".*?user=jason"
          }
        },
        "route": {
          "backends": [
            {
              "tags": [
                "v2"
              ]
            }
          ]
        }
      }
    ]`)

		response["services"] = []byte(
			`{
				    "services": [ "other" ]
				}`)

		allRoutes := fmt.Sprintf("{ \"services\": {%s,%s}}", response["reviews"], response["ratings"])
		reviewsRoutes := fmt.Sprintf("{ \"services\": {%s}}", response["reviews"])
		response["all_routes"] = []byte(allRoutes)
		response["reviews_routes"] = []byte(reviewsRoutes)
	})

	Describe("List of Routes", func() {

		Describe("Validate Controller URL", func() {

			JustBeforeEach(func() {
				app.Writer = bytes.NewBufferString("")
			})

			It("should exit with ErrControllerURLInvalid error", func() {
				err := app.Run([]string{"app", "--controller_url=123", "--registry_url=http://localhost", "route-list"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrControllerURLInvalid.Error()))
			})

			It("should exit with ErrControllerURLNotFound error", func() {
				err := app.Run([]string{"app", "--controller_url=", "--registry_url=http://localhost", "route-list"})
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
				err := app.Run([]string{"app", "--registry_url=123", "--controller_url=http://localhost", "route-list"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrRegistryURLInvalid.Error()))
			})

			It("should exit with ErrRegistryURLNotFound error", func() {
				err := app.Run([]string{"app", "--registry_url=", "--controller_url=http://localhost", "route-list"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrRegistryURLNotFound.Error()))
			})

			It("should error", func() {
				err := app.Run([]string{"app", "--registry_url=http://localhost", "--controller_url=http://localhost", "--x"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(app.Name))
			})

		})

		Describe("On usage error: [route-list bad]", func() {

			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v1/rules/routes"),
						ghttp.RespondWith(http.StatusOK, response["all_routes"]),
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
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "--registry_url=" + server.URL(), "route-list", "--bad"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(cmd.GetMetadata().Usage))
			})

		})

		Describe("List Routes: [route-list]", func() {

			BeforeEach(func() {

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v1/rules/routes"),
						ghttp.RespondWith(http.StatusOK, response["all_routes"]),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/services"),
						ghttp.RespondWith(http.StatusOK, response["services"]),
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
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "--registry_url=" + server.URL(), "route-list"})
				Expect(err).ToNot(HaveOccurred())
				// TODO: Validate output
			})

			It("should print a JSON", func() {
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "--registry_url=" + server.URL(), "route-list", "-o=json"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`"v2(header=\"Foo:bar\")"`))

			})

			It("should print a YAML", func() {
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "--registry_url=" + server.URL(), "route-list", "-o=yaml"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`- v2(header="Foo:bar")`))

			})

		})

		Describe("List Routes by service: [route-list]", func() {

			const service = "reviews"
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v1/rules/routes", "destination="+service),
						ghttp.RespondWith(http.StatusOK, response["reviews_routes"]),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/services"),
						ghttp.RespondWith(http.StatusOK, response["services"]),
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
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "--registry_url=" + server.URL(), "route-list", "-s=" + service})
				Expect(err).ToNot(HaveOccurred())
				// TODO: Validate output
			})

			It("should print a JSON", func() {
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "--registry_url=" + server.URL(), "route-list", "-o=json", "-s=" + service})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`"v2(header=\"Foo:bar\")"`))

			})

			It("should print a YAML", func() {
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "--registry_url=" + server.URL(), "route-list", "-o=yaml", "-s=" + service})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`- v2(header="Foo:bar")`))

			})

		})

	})
})
