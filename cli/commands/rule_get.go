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
	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/terminal"
	"github.com/amalgam8/amalgam8/cli/utils"
	ctrl "github.com/amalgam8/amalgam8/controller/client"
	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/urfave/cli"
)

// RuleGetCommand is used for the rule-get command.
type RuleGetCommand struct {
	ctx        *cli.Context
	controller *ctrl.Client
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
		Name:        "rule-get",
		Description: T("rule_get_description"),
		Usage:       T("rule_get_usage"),
		// TODO: Complete UsageText
		UsageText: T("rule_get_usage"),
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "id, i",
				Usage: T("rule_get_id_usage"),
			},
			cli.StringSliceFlag{
				Name:  "tag, t",
				Usage: T("rule_get_tag_usage"),
			},
			cli.StringSliceFlag{
				Name:  "destination, d",
				Usage: T("rule_get_destination_usage"),
			},
			cli.StringFlag{
				Name:  "output, o",
				Usage: T("rule_get_output_usage"),
				Value: YAML,
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
	controller, err := NewController(ctx)
	if err != nil {
		// Exit if the controller returned an error
		return nil
	}
	// Update the controller
	cmd.controller = controller
	format := ctx.String("output")

	if ctx.IsSet("all") {
		return cmd.PrettyPrint(&api.RuleFilter{}, format)
	}

	if ctx.IsSet("id") || ctx.IsSet("i") || ctx.IsSet("destination") || ctx.IsSet("d") || ctx.IsSet("tag") || ctx.IsSet("t") {

		filter := &api.RuleFilter{
			IDs:          ctx.StringSlice("id"),
			Destinations: ctx.StringSlice("destination"),
			Tags:         ctx.StringSlice("tag"),
		}

		return cmd.PrettyPrint(filter, format)
	}

	return cmd.DefaultAction(ctx)
}

// DefaultAction runs the default action.
func (cmd *RuleGetCommand) DefaultAction(ctx *cli.Context) error {
	return cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
}

// PrettyPrint prints the rules returned by the controller in given format.
func (cmd *RuleGetCommand) PrettyPrint(filter *api.RuleFilter, format string) error {
	rules, err := cmd.controller.ListRules(filter)
	if err != nil {
		return err
	}

	return utils.MarshallReader(cmd.ctx.App.Writer, rules, format)
}
