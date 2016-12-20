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

var _ = Describe("RecipeRun Command", func() {
	fmt.Println()
	utils.LoadLocales("../locales")
	T := utils.Language(common.DefaultLanguage)
	var cmd *cmds.RecipeRunCommand
	var app *cli.App
	var server *ghttp.Server
	response := make(map[string][]byte)
	JSONFilePath := "testdata/rule.json"
	ScriptPath := "testdata/hello"

	BeforeEach(func() {
		app = cli.NewApp()
		app.Name = T("app_name")
		app.Usage = T("app_usage")
		app.Version = T("app_version")
		app.Flags = config.GlobalFlags()
		server = ghttp.NewServer()
		term := terminal.NewUI(os.Stdin, os.Stdout)
		cmd = cmds.NewRecipeRunCommand(term)
		app.Commands = []cli.Command{cmd.GetMetadata()}
		app.Before = config.Before
		app.Action = config.DefaultAction
		app.OnUsageError = config.OnUsageError

		response["recipe_id"] = []byte(`
			{
  			"recipe_id": "recipe_id"
			}`)

		response["results"] = []byte(`
			{
			  "results": [
			    {
			      "dest": "productpage:v1",
			      "errormsg": "",
			      "max_latency": "7s",
			      "name": "bounded_response_time",
			      "result": true,
			      "source": "gateway:none"
			    },
			    {
			      "dest": "productpage:v1",
			      "errormsg": "",
			      "name": "http_status",
			      "result": true,
			      "source": "gateway:none",
			      "status": [
			        200,
			        302
			      ]
			    },
			    {
			      "dest": "reviews:v2",
			      "errormsg": "unexpected connection termination",
			      "name": "http_status",
			      "result": false,
			      "source": "productpage:v1",
			      "status": 200
			    },
			    {
			      "dest": "ratings:v1",
			      "errormsg": "",
			      "name": "http_status",
			      "result": true,
			      "source": "reviews:v2",
			      "status": 200
			    }
			  ]
			}`)
	})

	Describe("RecipeRun Command", func() {

		Describe("Validate Gremlin URL", func() {

			JustBeforeEach(func() {
				app.Writer = bytes.NewBufferString("")
			})

			It("should exit with ErrGremlinURLInvalid error", func() {
				err := app.Run([]string{"app", "--gremlin_url=123", "recipe-run"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrGremlinURLInvalid.Error()))
			})

			It("should exit and print ErrGremlinURLNotFound error", func() {
				err := app.Run([]string{"app", "--gremlin_url=", "recipe-run"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrGremlinURLNotFound.Error()))
			})

			It("should error", func() {
				err := app.Run([]string{"app", "--gremlin_url=http://localhost", "--x"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(app.Name))
			})

		})

		Describe("On usage error: [recipe-run bad]", func() {
			JustBeforeEach(func() {
				app.Writer = bytes.NewBufferString("")
			})

			It("should print the command help", func() {
				app.Writer = bytes.NewBufferString("")
				err := app.Run([]string{"app", "recipe-run", "--bad"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(cmd.GetMetadata().Usage))
			})
		})

		Describe("RecipeRun: [recipe-run]", func() {

			Describe("When '-topology' or '-scenarios' have not been set", func() {

				JustBeforeEach(func() {
					app.Writer = bytes.NewBufferString("")
				})

				It("should print ErrTopologyOrScenariosNotFound", func() {
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--gremlin_url=" + server.URL(), "recipe-run", "-topology="})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrTopologyOrScenariosNotFound.Error()))
				})

				It("should print ErrTopologyOrScenariosNotFound", func() {
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--gremlin_url=" + server.URL(), "recipe-run", "-scenarios="})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrTopologyOrScenariosNotFound.Error()))
				})
			})

			Describe("When topology file is invalid", func() {
				JustBeforeEach(func() {
					app.Writer = bytes.NewBufferString("")
				})

				It("should print ErrFileNotFound", func() {
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--gremlin_url=" + server.URL(), "recipe-run", "-topology=test.json", "-scenarios=test.json", "-w=0s", "-f=true"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrFileNotFound.Error()))
				})
			})

			Describe("When scenarios file is invalid", func() {
				JustBeforeEach(func() {
					app.Writer = bytes.NewBufferString("")
				})

				It("should print ErrFileNotFound", func() {
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--gremlin_url=" + server.URL(), "recipe-run", "-topology=" + JSONFilePath, "-scenarios=test.json", "-w=0s", "-f=true"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrFileNotFound.Error()))
				})
			})

			Describe("When checks file is invalid", func() {

				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v1/recipes"),
							ghttp.RespondWith(http.StatusOK, response["recipe_id"]),
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

				It("should print ErrFileNotFound", func() {
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--gremlin_url=" + server.URL(), "recipe-run", "-topology=" + JSONFilePath, "-scenarios=" + JSONFilePath, "-checks=test.json", "-w=0s", "-f=true"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrFileNotFound.Error()))
				})
			})

			Describe("When checks and script files have been provided", func() {

				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v1/recipes"),
							ghttp.RespondWith(http.StatusOK, response["recipe_id"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v1/recipes/recipe_id"),
							ghttp.RespondWith(http.StatusOK, response["recipe_id"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("DELETE", "/api/v1/recipes/recipe_id"),
							ghttp.RespondWith(http.StatusOK, response["recipe_id"]),
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

				It("should execute script", func() {
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--gremlin_url=" + server.URL(), "--debug", "recipe-run", "-topology=" + JSONFilePath, "-scenarios=" + JSONFilePath, "-checks=" + JSONFilePath, "-run-load-script=" + ScriptPath, "-w=0s", "-f=true"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("Hello world"))
				})

			})

			Describe("When no checks file has been provided", func() {

				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v1/recipes"),
							ghttp.RespondWith(http.StatusOK, response["recipe_id"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v1/recipes/recipe_id"),
							ghttp.RespondWith(http.StatusOK, response["recipe_id"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("DELETE", "/api/v1/recipes/recipe_id"),
							ghttp.RespondWith(http.StatusOK, response["recipe_id"]),
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

				It("should no verify recipe", func() {
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--gremlin_url=" + server.URL(), "--debug", "recipe-run", "-topology=" + JSONFilePath, "-scenarios=" + JSONFilePath, "-w=0s", "-f=true"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("Recipe created but not verified"))
				})

			})

			Describe("When print results in Table", func() {

				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v1/recipes"),
							ghttp.RespondWith(http.StatusOK, response["recipe_id"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v1/recipes/recipe_id"),
							ghttp.RespondWith(http.StatusOK, response["results"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("DELETE", "/api/v1/recipes/recipe_id"),
							ghttp.RespondWith(http.StatusOK, response["recipe_id"]),
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
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--gremlin_url=" + server.URL(), "--debug", "recipe-run", "-topology=" + JSONFilePath, "-scenarios=" + JSONFilePath, "-checks=" + JSONFilePath, "-w=0s", "-f=true"})
					Expect(err).ToNot(HaveOccurred())

					// TODO: Verify table
				})

			})

			Describe("When print results as JSON", func() {

				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v1/recipes"),
							ghttp.RespondWith(http.StatusOK, response["recipe_id"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v1/recipes/recipe_id"),
							ghttp.RespondWith(http.StatusOK, response["results"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("DELETE", "/api/v1/recipes/recipe_id"),
							ghttp.RespondWith(http.StatusOK, response["recipe_id"]),
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

				It("should print JSON", func() {
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--gremlin_url=" + server.URL(), "--debug", "recipe-run", "-topology=" + JSONFilePath, "-scenarios=" + JSONFilePath, "-checks=" + JSONFilePath, "-o=json", "-w=0s", "-f=true"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("\"dest\": \"productpage:v1\","))
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("\"dest\": \"productpage:v1\","))
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("\"errormsg\": \"unexpected connection termination\","))
				})

			})

			Describe("When print results as YAML", func() {

				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v1/recipes"),
							ghttp.RespondWith(http.StatusOK, response["recipe_id"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/api/v1/recipes/recipe_id"),
							ghttp.RespondWith(http.StatusOK, response["results"]),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("DELETE", "/api/v1/recipes/recipe_id"),
							ghttp.RespondWith(http.StatusOK, response["recipe_id"]),
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

				It("should print YAML", func() {
					app.Writer = bytes.NewBufferString("")
					err := app.Run([]string{"app", "--gremlin_url=" + server.URL(), "--debug", "recipe-run", "-topology=" + JSONFilePath, "-scenarios=" + JSONFilePath, "-checks=" + JSONFilePath, "-o=yaml", "-w=0s", "-f=true"})
					Expect(err).ToNot(HaveOccurred())
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("- dest: productpage:v1"))
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("- dest: productpage:v1"))
					Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("errormsg: unexpected connection termination"))
				})

			})

		})

	})
})
