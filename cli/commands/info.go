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
	"fmt"

	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/terminal"
	"github.com/amalgam8/amalgam8/cli/utils"
	"github.com/urfave/cli"
)

// InfoCommand is ised for the info command.
type InfoCommand struct {
	term terminal.UI
}

// NewInfoCommand constructs a new Info.
func NewInfoCommand(term terminal.UI) (cmd *InfoCommand) {
	return &InfoCommand{
		term: term,
	}
}

// GetMetadata returns the metadata.
func (cmd *InfoCommand) GetMetadata() cli.Command {
	T := utils.Language(common.DefaultLanguage)
	return cli.Command{
		Name:        "info",
		Description: T("info_description"),
		Usage:       T("info_usage"),
		Aliases:     []string{"i"},
		// TODO: Complete UsageText
		UsageText:    T("info_usage"),
		Before:       cmd.Before,
		OnUsageError: cmd.OnUsageError,
		Action:       cmd.Action,
	}
}

// Before runs before the Action
// https://godoc.org/github.com/urfave/cli#BeforeFunc
func (cmd *InfoCommand) Before(ctx *cli.Context) error {
	return nil
}

// OnUsageError is executed if an usage error occurs.
func (cmd *InfoCommand) OnUsageError(ctx *cli.Context, err error, isSubcommand bool) error {
	cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
	return nil
}

// Action runs when no subcommands are specified
// https://godoc.org/github.com/urfave/cli#ActionFunc
func (cmd *InfoCommand) Action(ctx *cli.Context) error {
	if len(ctx.Args()) > 0 {
		cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
		return nil
	}
	return cmd.DefaultAction(ctx)
}

// DefaultAction prints information about the controller URL's and tokens.
// Amalgam8 info...
//
// +---------------------+----------------------------+
// | Env. Variable       | Value                      |
// +---------------------+----------------------------+
// | A8_CONTROLLER_URL   | http://192.168.0.100:31200 |
// | A8_CONTROLLER_TOKEN |                            |
// | A8_GREMLIN_URL      | http://192.168.0.100:31500 |
// | A8_GREMLIN_TOKEN    |                            |
// | A8_DEBUG            | false                      |
// +---------------------+----------------------------+
func (cmd *InfoCommand) DefaultAction(ctx *cli.Context) error {
	table := cmd.term.NewTable()
	table.SetHeader([]string{"Env. Variable", "Value"})
	table.AddMultipleRows([][]string{
		{
			common.ControllerURL.EnvVar(),
			ctx.GlobalString(common.ControllerURL.Flag()),
		},
		{
			common.ControllerToken.EnvVar(),
			ctx.GlobalString(common.ControllerToken.Flag()),
		},
		{
			common.GremlinURL.EnvVar(),
			ctx.GlobalString(common.GremlinURL.Flag()),
		},
		{
			common.GremlinToken.EnvVar(),
			ctx.GlobalString(common.GremlinToken.Flag()),
		},
		{
			common.Debug.EnvVar(),
			ctx.GlobalString(common.Debug.Flag()),
		},
	})
	fmt.Fprintf(ctx.App.Writer, "\nAmalgam8 info...\n")
	table.SortByColumnIndex(0)
	table.PrintTable()
	return nil
}
