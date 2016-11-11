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

	"github.com/amalgam8/amalgam8/cli/api"
	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/terminal"
	"github.com/amalgam8/amalgam8/cli/utils"
	"github.com/urfave/cli"
)

// RuleGetCommand is used for the rule-get command.
type RuleGetCommand struct {
	ctx        *cli.Context
	controller api.ControllerClient
	term       terminal.UI
}

// NewRuleGetCommand constructs a new Rule Get command.
func NewRuleGetCommand(term terminal.UI) (cmd *RuleGetCommand) {
	return &RuleGetCommand{
		term: term,
	}
}

// GetMetadata returns the metada.
func (cmd *RuleGetCommand) GetMetadata() cli.Command {
	T := utils.Language(common.DefaultLanguage)
	return cli.Command{
		Name:        T("rule_get_name"),
		Description: T("rule_get_description"),
		Usage:       T("rule_get_usage"),
		// TODO: Complete UsageText
		UsageText: T("rule_get_name"),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "id, i",
				Usage: T("rule_get_id_usage"),
			},
			cli.StringFlag{
				Name:  "tag, t",
				Usage: T("rule_get_tag_usage"),
			},
			cli.StringFlag{
				Name:  "destination, d",
				Usage: T("rule_get_destination_usage"),
			},
			cli.StringFlag{
				Name:  "output, o",
				Usage: T("rule_get_output_usage"),
				Value: "json",
			},
			cli.BoolFlag{
				Name:  "all, a",
				Usage: T("rule_get_destination_usage"),
			},
		},
		Before:       cmd.Before,
		OnUsageError: cmd.OnUsageError,
		Action:       cmd.Action,
	}
}

// Before runs before the Action
// https://godoc.org/github.com/urfave/cli#BeforeFunc
func (cmd *RuleGetCommand) Before(ctx *cli.Context) error {
	// Update the context
	cmd.ctx = ctx
	return nil
}

// OnUsageError is executed if an usage error occurs.
func (cmd *RuleGetCommand) OnUsageError(ctx *cli.Context, err error, isSubcommand bool) error {
	return cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
}

// Action runs when no subcommands are specified
// https://godoc.org/github.com/urfave/cli#ActionFunc
func (cmd *RuleGetCommand) Action(ctx *cli.Context) error {
	controller, err := api.NewControllerClient(ctx)
	if err != nil {
		// Exit if the controller returned an error
		return nil
	}
	// Update the controller
	cmd.controller = controller

	format := ctx.String("output")
	if ctx.IsSet("id") {
		return cmd.PrettyPrint(fmt.Sprintf("?id=%s", ctx.String("id")), format)
	}

	if ctx.IsSet("destination") {
		return cmd.PrettyPrint(fmt.Sprintf("?destination=%s", ctx.String("destination")), format)
	}

	if ctx.IsSet("tag") {
		return cmd.PrettyPrint(fmt.Sprintf("?tags=%s", ctx.String("tag")), format)
	}

	if ctx.IsSet("all") {
		return cmd.PrettyPrint("", format)
	}

	return cmd.DefaultAction(ctx)
}

// DefaultAction runs the default action.
func (cmd *RuleGetCommand) DefaultAction(ctx *cli.Context) error {
	return cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
}

// PrettyPrint prints the rules returned by the controller in given format.
func (cmd *RuleGetCommand) PrettyPrint(id string, format string) error {
	rules, err := cmd.controller.Rules(id)
	if err != nil {
		return err
	}

	return utils.MarshallReader(cmd.ctx.App.Writer, rules, format)
}
