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

// prettyActionList is used to convert the table data into JSON or YAML
type prettyActionList struct {
	Destination string   `json:"destination" yaml:"destination"`
	ID          string   `json:"id,omitempty" yaml:"id,omitempty"`
	Priority    int      `json:"priority,omitempty" yaml:"priority,omitempty"`
	Match       string   `json:"match,omitempty" yaml:"match,omitempty"`
	Actions     []string `json:"actions,omitempty" yaml:"actions,omitempty"`
	Headers     []string `json:"headers,ommitempty" yaml:"headers,omitempty"`
}

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
func (cmd *ActionListCommand) PrettyPrint(serviceName string, format string) error {
	actionList := []prettyActionList{}

	actions, err := cmd.controller.GetActions()
	if err != nil {
		return err
	}

	// If not serviceName provided, show all Routes
	if serviceName == "" {
		for _, actionRules := range actions.ServiceActions {
			for _, action := range actionRules {
				actionList = append(
					actionList,
					prettyActionList{
						Destination: action.Destination,
						ID:          action.ID,
						Priority:    action.Priority,
						Match:       formatMatch(action.Match),
						Actions:     formatActions(action.Actions),
						Headers:     formatHeaders(action.Match),
					},
				)
			}
		}
	} else {
		// Show routes fot the given serviceName
		for _, actionRules := range actions.ServiceActions {
			for _, action := range actionRules {
				if action.Destination == serviceName {
					actionList = append(
						actionList,
						prettyActionList{
							Destination: action.Destination,
							ID:          action.ID,
							Priority:    action.Priority,
							Match:       formatMatch(action.Match),
							Actions:     formatActions(action.Actions),
							Headers:     formatHeaders(action.Match),
						},
					)
				}
			}
		}
	}

	return utils.MarshallReader(cmd.ctx.App.Writer, actionList, format)
}

// ActionTable prints the list of actions as a table.
// +-------------+----------------+----------------------+----------+-----------------------------+--------------------------------------+
// | Destination | Source         | Headers              | Priority | Actions                     | Rule Id                              |
// +-------------+----------------+----------------------+----------+-----------------------------+--------------------------------------+
// | reviews     | productpage:v1 | Cookie:.*?user=jason | 10       | v2(trace)                   | 2d381a94-1796-45c3-a1d8-3965051b61b1 |
// | ratings     | reviews:v2     | Cookie:.*?user=jason | 10       | v1(trace), v1(1->delay=7.0) | 454a8fb0-d260-4832-8007-5b5344c03c1f |
// | ratings     | reviews:v2     | Cookie:.*?user=jason | 10       | v1(1.0->delay=7.0)          | c2d98e32-8fd0-4e0d-a363-8adff99b0692 |
// | details     | productpage:v1 | Cookie:.*?user=jason | 10       | v1(trace)                   | 9c7198d7-d037-4cb6-8d48-b573608c7de9 |
// | productpage | gateway        | Cookie:.*?user=jason | 10       | v1(trace)                   | 0f12b977-9ab9-4d69-8dfe-3eae07c8f115 |
// +-------------+----------------+----------------------+----------+-----------------------------+--------------------------------------+
func (cmd *ActionListCommand) ActionTable(serviceName string) error {
	table := CommandTable{}
	table.header = []string{
		"Destination",
		"Source",
		"Headers",
		"Priority",
		"Actions",
		"Rule ID",
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
						formatMatch(action.Match),
						strings.Join(formatHeaders(action.Match), ", "),
						fmt.Sprint(action.Priority),
						strings.Join(formatActions(action.Actions), ", "),
						action.ID,
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
							formatMatch(action.Match),
							strings.Join(formatHeaders(action.Match), ", "),
							fmt.Sprint(action.Priority),
							strings.Join(formatActions(action.Actions), ", "),
							action.ID,
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
		fmt.Fprintf(&buf, "%s", match.Source.Name)
		if len(match.Source.Tags) > 0 {
			fmt.Fprintf(&buf, ":%s", strings.Join(match.Source.Tags, ","))
		}
	}
	return buf.String()
}

func formatHeaders(match *api.MatchRules) []string {
	headers := []string{}
	for key, value := range match.Headers {
		headers = append(headers, fmt.Sprintf("%s:%s", key, value))
	}
	return headers
}

func formatActions(actions *api.ActionsRules) []string {
	result := []string{}
	if actions != nil {
		for _, action := range *actions {
			buf := bytes.Buffer{}
			switch action.Action {
			case "delay":
				fmt.Fprintf(&buf, "%s(%g->delay=%g)", strings.Join(action.Tags, ","), action.Probability, action.Duration)
			case "abort":
				fmt.Fprintf(&buf, "%s(%g->abort=%d)", strings.Join(action.Tags, ","), action.Probability, action.ReturnCode)
			case "trace":
				fmt.Fprintf(&buf, "%s(trace)", strings.Join(action.Tags, ","))
			}
			result = append(result, buf.String())
		}
	}
	return result
}
