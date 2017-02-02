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
	ctrl "github.com/amalgam8/amalgam8/controller/client"
	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/urfave/cli"
)

// RuleUpdateCommand is used for the rule-update command.
type RuleUpdateCommand struct {
	ctx        *cli.Context
	controller *ctrl.Client
	term       terminal.UI
}

// NewRuleUpdateCommand constructs a new Rule Update.
func NewRuleUpdateCommand(term terminal.UI) (cmd *RuleUpdateCommand) {
	return &RuleUpdateCommand{
		term: term,
	}
}

// GetMetadata returns the metadata.
func (cmd *RuleUpdateCommand) GetMetadata() cli.Command {
	T := utils.Language(common.DefaultLanguage)
	return cli.Command{
		Name:        "rule-update",
		Description: T("rule_update_description"),
		Usage:       T("rule_update_usage"),
		// TODO: Complete UsageText
		UsageText: T("rule_update_usage"),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "file, f",
				Usage: T("rule_update_file_usage"),
			},
			cli.BoolFlag{
				Name:  "redirection, r",
				Usage: T("rule_update_input_redirection_usage"),
			},
		},
		Before:       cmd.Before,
		OnUsageError: cmd.OnUsageError,
		Action:       cmd.Action,
	}
}

// Before runs before the Action
// https://godoc.org/github.com/urfave/cli#BeforeFunc
func (cmd *RuleUpdateCommand) Before(ctx *cli.Context) error {
	// Update the context
	cmd.ctx = ctx
	return nil
}

// OnUsageError is executed if an usage error occurs.
func (cmd *RuleUpdateCommand) OnUsageError(ctx *cli.Context, err error, isSubcommand bool) error {
	return cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
}

// Action runs when no subcommands are specified
// https://godoc.org/github.com/urfave/cli#ActionFunc
func (cmd *RuleUpdateCommand) Action(ctx *cli.Context) error {
	T := utils.Language(common.DefaultLanguage)
	controller, err := NewController(ctx)
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

		// Verify that the rules provided in the file have the right structure
		rulesReader, err := utils.ValidateRulesFormat(reader)
		if err != nil {
			return err
		}

		rules := &api.RulesSet{}

		// Read rules
		err = utils.UnmarshalReader(rulesReader, format, rules)
		if err != nil {
			return err
		}

		// Add errors in client
		_, err = cmd.controller.UpdateRules(rules)
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.ctx.App.Writer, T("request_completed"))
		return nil
	}

	return cmd.DefaultAction(ctx)
}

// DefaultAction captures the rules provided in the terminal.
func (cmd *RuleUpdateCommand) DefaultAction(ctx *cli.Context) error {
	T := utils.Language(common.DefaultLanguage)
	reader, format, err := utils.ScannerLines(cmd.ctx.App.Writer, "Enter DSL Rules", ctx.Bool("redirection"))

	fmt.Fprintf(cmd.ctx.App.Writer, "\n\n")

	if err != nil {
		fmt.Fprintf(cmd.ctx.App.Writer, "%s\n\n", err.Error())
		return nil
	}

	rules := &api.RulesSet{}

	// Read Rules
	err = utils.UnmarshalReader(reader, format, rules)
	if err != nil {
		return err
	}

	// TODO: return user-friendly errors
	_, err = cmd.controller.UpdateRules(rules)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.ctx.App.Writer, T("request_completed"))
	return nil
}
