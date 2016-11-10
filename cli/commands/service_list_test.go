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
	cmds "github.com/amalgam8/amalgam8/cli/commands"
	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/flags"
	"github.com/amalgam8/amalgam8/cli/terminal"
	"github.com/amalgam8/amalgam8/cli/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/urfave/cli"
	// "io/ioutil"
	"net/http"
	"os"
)

var _ = Describe("Service-List", func() {
	fmt.Println()
	utils.LoadLocales()
	T := utils.Language(common.DefaultLanguage)
	var cmd *cmds.ServiceListCommand
	var app *cli.App
	var server *ghttp.Server
	response := make(map[string][]byte)

	BeforeEach(func() {
		app = cli.NewApp()
		app.Name = T("app_name")
		app.Usage = T("app_usage")
		app.Version = T("app_version")
		app.Flags = flags.GlobalFlags()
		server = ghttp.NewServer()
		term := terminal.NewUI(os.Stdin, os.Stdout)
		cmd = cmds.NewServiceListCommand(term)
		app.Commands = []cli.Command{cmd.GetMetadata()}
		app.Setup()

		response["services"] = []byte(
			`{
			    "services": [ "ratings" ]
			}`)
		response["ratings"] = []byte(
			`{
  			  "service_name": "ratings",
  			  "instances": [
  			    {
  			      "id": "asdfghjkl",
  			      "service_name": "ratings",
  			      "endpoint": {
  			        "type": "http",
  			        "value": "localhost:9080"
  			      },
  			      "ttl": 60,
  			      "status": "UP",
  			      "last_heartbeat": "2016-10-10T00:28:24.483613521Z",
  			      "tags": [
  			        "v1"
  			      ]
  			    },
  					{
  			      "id": "asdfghjkl",
  			      "service_name": "ratings",
  			      "endpoint": {
  			        "type": "http",
  			        "value": "localhost:9080"
  			      },
  			      "ttl": 60,
  			      "status": "UP",
  			      "last_heartbeat": "2016-10-10T00:28:24.483613521Z",
  			      "tags": [
  			        "v2"
  			      ]
  			    }
  			  ]
  			}`)
	})

	Describe("List Services", func() {

		Describe("Validate Registry URL", func() {

			JustBeforeEach(func() {
				app.Writer = bytes.NewBufferString("")
			})

			It("should exit with ErrRegistryURLInvalid error", func() {
				err := app.Run([]string{"app", "--registry_url=123", "service-list"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrRegistryURLInvalid.Error()))
			})

			It("should exit with ErrRegistryURLNotFound error", func() {
				err := app.Run([]string{"app", "--registry_url=", "service-list"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrRegistryURLNotFound.Error()))
			})

			It("should error", func() {
				err := app.Run([]string{"app", "--registry_url=http://localhost", "--x"})
				Expect(err).To(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("Incorrect Usage"))
			})

		})

		Describe("On usage error: [service-list bad]", func() {

			BeforeEach(func() {
				server.AppendHandlers(
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

			It("should print the command help", func() {
				app.Writer = bytes.NewBufferString("")
				err := app.Run([]string{"app", "--registry_url=" + server.URL(), "service-list", "--bad"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(cmd.GetMetadata().Usage))
			})

		})

		Describe("List Services: [service-list]", func() {

			BeforeEach(func() {

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/services"),
						ghttp.RespondWith(http.StatusOK, response["services"]),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/services/ratings"),
						ghttp.RespondWith(http.StatusOK, response["ratings"]),
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
				err := app.Run([]string{"app", "--registry_url=" + server.URL(), "--debug=true", "service-list"})
				Expect(err).ToNot(HaveOccurred())
				// TODO: Validate output
			})

			It("should print a JSON", func() {
				err := app.Run([]string{"app", "--registry_url=" + server.URL(), "--debug=true", "service-list", "-o=json"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`"service": "ratings"`))
				fmt.Println(app.Writer)
			})

			It("should print a YAML", func() {
				err := app.Run([]string{"app", "--registry_url=" + server.URL(), "--debug=true", "service-list", "-o=yaml"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`- service: ratings`))
				fmt.Println(app.Writer)
			})

		})

	})
})
