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

var _ = Describe("Rule-Update", func() {
	fmt.Println()
	utils.LoadLocales("../locales")
	T := utils.Language(common.DefaultLanguage)
	var cmd *cmds.RuleUpdateCommand
	var app *cli.App
	var server *ghttp.Server
	response := make(map[string][]byte)
	JSONFilePath := "testdata/rule.json"
	YAMLFilePath := "testdata/rule.yaml"

	BeforeEach(func() {
		app = cli.NewApp()
		app.Name = T("app_name")
		app.Usage = T("app_usage")
		app.Version = T("app_version")
		app.Flags = config.GlobalFlags()
		server = ghttp.NewServer()
		term := terminal.NewUI(os.Stdin, os.Stdout)
		cmd = cmds.NewRuleUpdateCommand(term)
		app.Commands = []cli.Command{cmd.GetMetadata()}
		app.Before = config.Before
		app.Action = config.DefaultAction
		app.OnUsageError = config.OnUsageError
		response["empty"] = []byte(`{}`)
	})

	Describe("Update Rules", func() {

		Describe("Validate Controller URL", func() {

			JustBeforeEach(func() {
				app.Writer = bytes.NewBufferString("")
			})

			It("should exit with ErrControllerURLInvalid error", func() {
				err := app.Run([]string{"app", "--controller_url=123", "rule-update"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrControllerURLInvalid.Error()))
			})

			It("should exit with ErrControllerURLInvalid error", func() {
				err := app.Run([]string{"app", "--controller_url=", "rule-update"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(common.ErrControllerURLNotFound.Error()))
			})

			It("should error", func() {
				err := app.Run([]string{"app", "--controller_url=http://localhost", "--x"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(app.Name))
			})

		})

		Describe("On usage error: [rule-update bad]", func() {

			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/rules"),
						ghttp.RespondWith(http.StatusOK, ""),
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
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "rule-update", "--bad"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(cmd.GetMetadata().Usage))
			})

		})

		Describe("Update rules using a JSON file: [rule-update -f]", func() {

			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v1/rules"),
						ghttp.RespondWith(http.StatusOK, response["empty"]),
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

			It("should read rules from the JSON file", func() {
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "rule-update", "-f=" + JSONFilePath})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("Request Completed"))
			})

		})

		Describe("Update rules using a YAML file: [rule-update -f]", func() {

			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/v1/rules"),
						ghttp.RespondWith(http.StatusOK, response["empty"]),
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

			It("should read rules from the YAML file", func() {
				err := app.Run([]string{"app", "--controller_url=" + server.URL(), "rule-update", "-f=" + YAMLFilePath})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring("Request Completed"))
			})

		})

	})
})
