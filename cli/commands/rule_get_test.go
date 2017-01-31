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

var _ = Describe("Rule-Get", func() {
	fmt.Println()
	utils.LoadLocales("../locales")
	T := utils.Language(common.DefaultLanguage)
	var cmd *cmds.RuleGetCommand
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
		cmd = cmds.NewRuleGetCommand(term)
		app.Commands = []cli.Command{cmd.GetMetadata()}
		app.Before = config.Before
		app.Action = config.DefaultAction
		app.OnUsageError = config.OnUsageError

		response["abc123"] = []byte(`
			{
	      "id": "abc123",
	      "priority": 10,
	      "destination": "reviews",
	      "match": {
	        "source": {
	          "name": "ratings",
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
	    }`)

		response["xyz123"] = []byte(`
			{
	      "id": "xyz123",
	      "priority": 30,
	      "destination": "productpage",
	      "match": {
	        "source": {
	          "name": "ratings",
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
	    }`)

		response["reviews"] = response["abc123"]
		response["v1"] = response["xyz123"]
		allRules := fmt.Sprintf("{ \"rules\": [%s,%s]}", response["abc123"], response["xyz123"])
		response["all_rules"] = []byte(allRules)
	})

	Describe("Get Rules", func() {

		Describe("Validate Controller URL", func() {

			JustBeforeEach(func() {
				app.Writer = bytes.NewBufferString("")
			})

			It("should exit with ErrControllerURLInvalid error", func() {
				err := app.Run([]string{"app", "--controller_url=123", "rule-get"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrControllerURLInvalid.Error()))
			})

			It("should exit with ErrControllerURLInvalid error", func() {
				err := app.Run([]string{"app", "--controller_url=", "rule-get"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrControllerURLNotFound.Error()))
			})

			It("should error", func() {
				err := app.Run([]string{"app", "--controller_url=http://localhost", "--x"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(app.Name))
			})

		})

		Describe("On usage error: [rule-get bad]", func() {

			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v1/rules"),
						ghttp.RespondWith(http.StatusOK, response["all_rules"]),
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
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "rule-get", "--bad"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(cmd.GetMetadata().Usage))
			})

		})

		Describe("Get rules by id: [rule-get -a]", func() {

			const ID = "abc123"
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v1/rules"),
						ghttp.RespondWith(http.StatusOK, response["all_rules"]),
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

			It("should print rules as JSON", func() {
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "rule-get", "-a", "-o=json"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`"id": "abc123"`))
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`"id": "xyz123"`))

			})

			It("should print rules as YAML", func() {
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "rule-get", "-a"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("- id: abc123"))
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("- id: xyz123"))

			})

		})

		Describe("Get rules by id: [rule-get -i id]", func() {

			const ID = "abc123"
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v1/rules", "id="+ID),
						ghttp.RespondWith(http.StatusOK, `{ "rules": [`+string(response[ID])+`]}`),
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

			It("should print rules as JSON", func() {
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "rule-get", "-i=" + ID, "-o=json"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`"id": "` + ID + `"`))

			})

			It("should print rules as YAML", func() {
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "rule-get", "-i=" + ID})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`- id: ` + ID))

			})

		})

		Describe("Get rules by destination: [rule-get -d dest]", func() {

			const destination = "reviews"
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v1/rules", "destination="+destination),
						ghttp.RespondWith(http.StatusOK, `{ "rules": [`+string(response[destination])+`]}`),
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

			It("should print rules as JSON", func() {
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "rule-get", "-d=" + destination, "-o=json"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`"destination": "` + destination + `"`))

			})

			It("should print rules as YAML", func() {
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "rule-get", "-d=" + destination})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`destination: ` + destination))

			})

		})

		Describe("Get rules by tag: [rule-get -t tag]", func() {

			const tags = "v1"
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v1/rules", "tag="+tags),
						ghttp.RespondWith(http.StatusOK, `{ "rules": [`+string(response[tags])+`]}`),
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

			It("should print rules as JSON", func() {
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "rule-get", "-t=" + tags, "-o=json"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`"tags"`))
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`"` + tags + `"`))

			})

			It("should print rules as YAML", func() {
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "rule-get", "-t=" + tags})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`- ` + tags))

			})

		})

	})
})
