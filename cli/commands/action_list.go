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
	"bytes"
	"fmt"
	"strings"

	"github.com/amalgam8/amalgam8/cli/api"
	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/terminal"
	"github.com/amalgam8/amalgam8/cli/utils"
	"github.com/urfave/cli"
)

// ActionListCommand is used for the action-list command.
type ActionListCommand struct {
	ctx        *cli.Context
	controller api.ControllerClient
	term       terminal.UI
}

// NewActionListCommand constructs a new Action List.
func NewActionListCommand(term terminal.UI) (cmd *ActionListCommand) {
	return &ActionListCommand{
		term: term,
	}
}

// GetMetadata returns the metadata.
func (cmd *ActionListCommand) GetMetadata() cli.Command {
	T := utils.Language(common.DefaultLanguage)
	return cli.Command{
		Name:        T("action_list_name"),
		Description: T("action_list_description"),
		Usage:       T("action_list_usage"),
		// TODO: Complete UsageText
		UsageText: T("action_list_name") + "[--json]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "output, o",
				Usage: T("action_list_output_usage"),
				Value: TABLE,
			},
			cli.StringFlag{
				Name:  "service, s",
				Usage: T("action_list_service_usage"),
				Value: "",
			},
		},
		Before:       cmd.Before,
		OnUsageError: cmd.OnUsageError,
		Action:       cmd.Action,
	}
}

// Before runs before the Action
// https://godoc.org/github.com/urfave/cli#BeforeFunc
func (cmd *ActionListCommand) Before(ctx *cli.Context) error {
	// Update the context
	cmd.ctx = ctx
	return nil
}

// OnUsageError is executed if an usage error occurs.
func (cmd *ActionListCommand) OnUsageError(ctx *cli.Context, err error, isSubcommand bool) error {
	cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
	return nil
}

// Action runs when no subcommands are specified
// https://godoc.org/github.com/urfave/cli#ActionFunc
func (cmd *ActionListCommand) Action(ctx *cli.Context) error {
	controller, err := api.NewControllerClient(ctx)
	if err != nil {
		// Exit if the controller returned an error
		return nil
	}
	// Update the controller
	cmd.controller = controller

	if ctx.NArg() > 0 {
		cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
		return nil
	}

	format := ctx.String("output")
	switch format {
	case JSON, YAML:
		return cmd.PrettyPrint(ctx.String("service"), format)
	case TABLE:
		return cmd.ActionTable(ctx.String("service"))
	}

	return cmd.DefaultAction(ctx)
}

// DefaultAction runs the default action.
func (cmd *ActionListCommand) DefaultAction(ctx *cli.Context) error {
	return cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
}

// PrettyPrint prints the list of services in the given format (json or yaml).
func (cmd *ActionListCommand) PrettyPrint(service string, format string) error {
	routes, err := cmd.controller.GetActions()
	if err != nil {
		return err
	}

	if service != "" {
		if route, ok := routes.ServiceActions[service]; ok {
			return utils.MarshallReader(cmd.ctx.App.Writer, route, format)
		}
		return common.ErrNotFound
	}

	return utils.MarshallReader(cmd.ctx.App.Writer, routes, format)
}

// ActionTable prints the list of actions as a table.
// +-------------+----------+----------+---------------------------------------------------------+-----------------------------------+
// | Destination | Rule Id  | Priority | Match                                                   | Actions (EXPERIMENTAL)            |
// +-------------+----------+----------+---------------------------------------------------------+-----------------------------------+
// | details     | 9c7198d7 | 10       | source="productpage:v1", header="Cookie:.*?user=jason"  | action=trace, tags=v1, prob=0,... |
// | productpage | 0f12b977 | 10       | source="gateway                                         | action=trace, tags=v1, prob=0,... |
// | ratings     | 454a8fb0 | 10       | source="reviews:v2"                                     | action=trace, tags=v1, prob=0,... |
// | ratings     | dc8b5ffe | 20       | source="reviews:v2"                                     | action=delay, tags=v1, prob=1,... |
// | reviews     | 2d381a94 | 10       | source="productpage:v1", header="Cookie:.*?user=jason"  | action=trace, tags=v2, prob=0,... |
// +-------------+----------+----------+---------------------------------------------------------+-----------------------------------+
func (cmd *ActionListCommand) ActionTable(serviceName string) error {
	table := CommandTable{}
	table.header = []string{
		"Destination",
		"Rule Id",
		"Priority",
		"Match",
		"Actions",
	}

	actions, err := cmd.controller.GetActions()
	if err != nil {
		return err
	}

	// If not serviceName provided, show all Routes
	if serviceName == "" {
		for _, actionRules := range actions.ServiceActions {
			for _, action := range actionRules {
				table.body = append(
					table.body,
					[]string{
						action.Destination,
						action.ID,
						fmt.Sprint(action.Priority),
						formatMatch(action.Match),
						formatActions(action.Actions),
					},
				)
			}
		}
	} else {
		// Show routes fot the given serviceName
		for _, actionRules := range actions.ServiceActions {
			for _, action := range actionRules {
				if action.Destination == serviceName {
					table.body = append(
						table.body,
						[]string{
							action.Destination,
							action.ID,
							fmt.Sprint(action.Priority),
							formatMatch(action.Match),
							formatActions(action.Actions),
						},
					)
				}
			}
		}
	}

	cmd.term.PrintTable(table.header, table.body)
	return nil
}

func formatMatch(match *api.MatchRules) string {
	buf := bytes.Buffer{}
	if match.Source != nil {
		fmt.Fprintf(&buf, "source=\"%s", match.Source.Name)
		if len(match.Source.Tags) > 0 {
			fmt.Fprintf(&buf, ":%s\"", strings.Join(match.Source.Tags, ","))
		}
	}
	buf.WriteString(", ")
	for key, value := range match.Headers {
		buf.WriteString("header=")
		fmt.Fprintf(&buf, "\"%s:%s\" ", key, value)
	}
	return buf.String()
}

func formatActions(actions *api.ActionsRules) string {
	buf := bytes.Buffer{}
	if actions != nil {
		for _, action := range *actions {
			switch action.Action {
			case "delay":
				fmt.Fprintf(&buf, "(action=%s, tags=%s, prob=%g, dur=%g)", action.Action, strings.Join(action.Tags, ", "), action.Probability, action.Duration)
			case "abort":
				fmt.Fprintf(&buf, "(action=%s, tags=%s, code=%d)", action.Action, strings.Join(action.Tags, ", "), action.ReturnCode)
			case "trace":
				fmt.Fprintf(&buf, "(action=%s, tags=%s, key=%s, val=%s)", action.Action, strings.Join(action.Tags, ", "), action.LogKey, action.LogValue)
			}
			buf.WriteString(" ")
		}
	}
	return buf.String()
}
