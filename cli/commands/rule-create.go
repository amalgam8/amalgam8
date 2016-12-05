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

// RuleCreateCommand is used for the rule-create command.
type RuleCreateCommand struct {
	ctx        *cli.Context
	controller api.ControllerClient
	term       terminal.UI
}

// NewRuleCreateCommand constructs a new Rule Create.
func NewRuleCreateCommand(term terminal.UI) (cmd *RuleCreateCommand) {
	return &RuleCreateCommand{
		term: term,
	}
}

// GetMetadata returns the metadata.
func (cmd *RuleCreateCommand) GetMetadata() cli.Command {
	T := utils.Language(common.DefaultLanguage)
	return cli.Command{
		Name:        T("rule_create_name"),
		Description: T("rule_create_description"),
		Usage:       T("rule_create_usage"),
		// TODO: Complete UsageText
		UsageText: T("rule_create_name"),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "file, f",
				Usage: T("rule_create_file_usage"),
			},
		},
		Before:       cmd.Before,
		OnUsageError: cmd.OnUsageError,
		Action:       cmd.Action,
	}
}

// Before runs before the Action
// https://godoc.org/github.com/urfave/cli#BeforeFunc
func (cmd *RuleCreateCommand) Before(ctx *cli.Context) error {
	// Update the context
	cmd.ctx = ctx
	return nil
}

// OnUsageError is executed if an usage error occurs.
func (cmd *RuleCreateCommand) OnUsageError(ctx *cli.Context, err error, isSubcommand bool) error {
	return cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
}

// Action runs when no subcommands are specified
// https://godoc.org/github.com/urfave/cli#ActionFunc
func (cmd *RuleCreateCommand) Action(ctx *cli.Context) error {
	controller, err := api.NewControllerClient(ctx)
	if err != nil {
		// Exit if the controller returned an error
		return nil
	}
	// Update the controller
	cmd.controller = controller

	if ctx.IsSet("file") {
		reader, format, err := utils.ReadInputFile(ctx.String("file"))
		if err != nil {
			return err
		}

		// Convert YAML to JSON
		if format == utils.YAML {
			reader, err = utils.YAMLToJSON(reader, &api.RuleList{})
			if err != nil {
				return err
			}
		}
		// Add errors in client
		result, err := cmd.controller.SetRules(reader)
		if err != nil {
			return err
		}
		return utils.MarshallReader(cmd.ctx.App.Writer, result, utils.JSON)

	}

	return cmd.DefaultAction(ctx)
}

// DefaultAction captures the rules provided in the terminal.
func (cmd *RuleCreateCommand) DefaultAction(ctx *cli.Context) error {
	reader, format, err := utils.ScannerLines(cmd.ctx.App.Writer, "Enter DSL Rules")

	fmt.Fprintf(cmd.ctx.App.Writer, "\n\n")

	if err != nil {
		fmt.Fprintf(cmd.ctx.App.Writer, "%s\n\n", err.Error())
		return nil
	}

	// Convert YAML to JSON
	if format == utils.YAML {
		reader, err = utils.YAMLToJSON(reader, &api.RuleList{})
		if err != nil {
			return err
		}
	}

	// TODO: return user-friendly errors
	result, err := cmd.controller.SetRules(reader)
	if err != nil {
		return err
	}
	return utils.MarshallReader(cmd.ctx.App.Writer, result, utils.JSON)
}
