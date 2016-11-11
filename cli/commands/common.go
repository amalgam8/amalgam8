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

package commands

import (
	"os"
	"strings"

	"github.com/amalgam8/amalgam8/cli/terminal"
	"github.com/amalgam8/amalgam8/cli/utils"
	"github.com/urfave/cli"
)

var (
	// JSON .
	JSON = strings.ToLower(utils.JSON)
	// YAML .
	YAML = strings.ToLower(utils.YAML)
	// TABLE .
	TABLE = "table"
)

// CommandTable .
type CommandTable struct {
	header []string
	body   [][]string
}

// GlobalCommands .
func GlobalCommands(term terminal.UI) []cli.Command {
	if term == nil {
		term = terminal.NewUI(os.Stdin, os.Stdout)
	}

	return []cli.Command{
		NewServiceListCommand(term).GetMetadata(),
		NewRouteListCommand(term).GetMetadata(),
		NewActionListCommand(term).GetMetadata(),
		NewRuleCreateCommand(term).GetMetadata(),
		NewRuleGetCommand(term).GetMetadata(),
		NewRuleDeleteCommand(term).GetMetadata(),
		NewInfoCommand(term).GetMetadata(),
	}
}
