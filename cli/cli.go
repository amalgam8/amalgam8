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

package cli

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/config"
	"github.com/amalgam8/amalgam8/cli/terminal"
	"github.com/amalgam8/amalgam8/cli/utils"
	"github.com/urfave/cli"
)

// Main is the entrypoint for the cli when running as an executable
func Main() {
	// Set log level
	logrus.SetLevel(logrus.ErrorLevel)

	// Load translation files
	err := utils.LoadLocales("")
	if err != nil {
		logrus.Error(err)
		return
	}

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
	app.Flags = config.GlobalFlags()
	app.Commands = config.GlobalCommands(term)
	app.Before = config.Before
	app.Action = config.DefaultAction
	app.OnUsageError = config.OnUsageError

	err = app.Run(os.Args)
	if err != nil {
		logrus.Error(err.Error())
	}
}
