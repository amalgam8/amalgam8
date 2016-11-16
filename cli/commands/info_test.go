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

	"os"

	cmds "github.com/amalgam8/amalgam8/cli/commands"
	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/config"
	"github.com/amalgam8/amalgam8/cli/terminal"
	"github.com/amalgam8/amalgam8/cli/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/urfave/cli"
)

var _ = Describe("Info Command", func() {
	fmt.Println()
	utils.LoadLocales("../locales")
	T := utils.Language(common.DefaultLanguage)
	var cmd *cmds.InfoCommand
	var app *cli.App

	BeforeEach(func() {
		app = cli.NewApp()
		app.Name = T("app_name")
		app.Usage = T("app_usage")
		app.Version = T("app_version")
		app.Flags = config.GlobalFlags()
		term := terminal.NewUI(os.Stdin, os.Stdout)
		cmd = cmds.NewInfoCommand(term)
		app.Commands = []cli.Command{cmd.GetMetadata()}
		app.Before = config.Before
		app.Action = config.DefaultAction
		app.OnUsageError = config.OnUsageError
	})

	Describe("Info Command", func() {

		Describe("On usage error: [info bad]", func() {
			JustBeforeEach(func() {
				app.Writer = bytes.NewBufferString("")
			})

			It("should print the command help", func() {
				app.Writer = bytes.NewBufferString("")
				err := app.Run([]string{"app", "info", "--bad"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(cmd.GetMetadata().Usage))
			})

		})

		Describe("List Services: [info]", func() {

			JustBeforeEach(func() {
				app.Writer = bytes.NewBufferString("")
			})

			It("should print table", func() {
				err := app.Run([]string{"app", "info"})
				Expect(err).ToNot(HaveOccurred())
				Expect(fmt.Sprint(app.Writer)).To(ContainSubstring(`Amalgam8 info...`))
				// TODO: Validate output for table
			})

		})

	})
})
