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

package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/cli/api"
	"github.com/amalgam8/amalgam8/cli/commands"
	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/flags"
	"github.com/amalgam8/amalgam8/cli/terminal"
	"github.com/amalgam8/amalgam8/cli/utils"
	"github.com/urfave/cli"
	"os"
	"strings"
)

func init() {
	// Set log level
	logrus.SetLevel(logrus.ErrorLevel)

	// Load translation files
	err := utils.LoadLocales()
	if err != nil {
		logrus.Error(err)
		return
	}
}

func main() {
	// Set default translation language
	T := utils.Language(common.DefaultLanguage) // TODO: Add more languages

	// Create the CLI App
	app := cli.NewApp()
	// Terminal
	term := terminal.NewUI(nil, app.Writer)
	app.Name = T("app_name")
	app.Usage = T("app_usage")
	app.Version = T("app_version")
	app.Metadata = map[string]interface{}{}
	app.Flags = flags.GlobalFlags()
	app.Commands = commands.GlobalCommands(term)
	app.Before = before
	app.Action = defaultAction
	app.OnUsageError = onUsageError

	err := app.Run(os.Args)
	if err != nil {
		logrus.Error(err.Error())
	}
}

// before runs after the context is ready and before the Action
// https://godoc.org/github.com/urfave/cli#BeforeFunc
func before(ctx *cli.Context) error {
	return nil
}

func onUsageError(ctx *cli.Context, err error, isSubcommand bool) error {
	if err != nil {
		logrus.WithError(err).Debug("Error")

		if strings.Contains(err.Error(), common.ErrUnknowFlag.Error()) {
			cli.ShowAppHelp(ctx)
			return nil
		}

		if strings.Contains(err.Error(), common.ErrInvalidFlagArg.Error()) {
			flag := err.Error()[strings.LastIndex(err.Error(), "-")+1:]

			if flag == common.RegistryURL.Flag() {
				url, errURL := api.ValidateRegistryURL(ctx)
				if err != nil {
					fmt.Fprintf(ctx.App.Writer, fmt.Sprintf("%s: %q\n\n", errURL.Error(), url))
				}
				return nil
			}

			if flag == common.ControllerURL.Flag() {
				url, errURL := api.ValidateControllerURL(ctx)
				if err != nil {
					fmt.Fprintf(ctx.App.Writer, fmt.Sprintf("%s: %q\n\n", errURL.Error(), url))
				}
				return nil
			}
		}

		fmt.Fprintf(ctx.App.Writer, "\nError: %#v\n\n", err.Error())
		return nil
	}

	cli.ShowAppHelp(ctx)
	return nil
}

func defaultAction(ctx *cli.Context) error {

	// Validate flags if not command has been specified
	if ctx.NumFlags() > 0 && ctx.NArg() == 0 {
		url, err := api.ValidateRegistryURL(ctx)
		if err != nil {
			fmt.Fprintf(ctx.App.Writer, fmt.Sprintf("\n%s: %q\n\n", err.Error(), url))
			return nil
		}
		url, err = api.ValidateControllerURL(ctx)
		if err != nil {
			fmt.Fprintf(ctx.App.Writer, fmt.Sprintf("\n%s: %q\n\n", err.Error(), url))
			return nil
		}
	}

	cli.ShowAppHelp(ctx)
	return nil
}
