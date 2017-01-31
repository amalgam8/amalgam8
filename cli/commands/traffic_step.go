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
	"strings"

	"fmt"

	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/terminal"
	"github.com/amalgam8/amalgam8/cli/utils"
	ctrl "github.com/amalgam8/amalgam8/controller/client"
	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/urfave/cli"
)

// TrafficStepCommand is used for the route-list command.
type TrafficStepCommand struct {
	ctx        *cli.Context
	controller *ctrl.Client
	term       terminal.UI
}

// NewTrafficStepCommand constructs a new TrafficStep.
func NewTrafficStepCommand(term terminal.UI) (cmd *TrafficStepCommand) {
	return &TrafficStepCommand{
		term: term,
	}
}

// GetMetadata returns the metadata.
func (cmd *TrafficStepCommand) GetMetadata() cli.Command {
	T := utils.Language(common.DefaultLanguage)
	return cli.Command{
		Name:        "traffic-step",
		Description: T("traffic_step_description"),
		Usage:       T("traffic_step_usage"),
		// TODO: Complete UsageText
		UsageText: T("traffic_step_usage"),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "service, s",
				Usage: T("traffic_step_service_usage"),
				Value: "",
			},
			cli.IntFlag{
				Name:  "amount, a",
				Usage: T("traffic_step_amount_usage"),
				Value: 10,
			},
		},
		Before:       cmd.Before,
		OnUsageError: cmd.OnUsageError,
		Action:       cmd.Action,
	}
}

// Before runs before the Action
// https://godoc.org/github.com/urfave/cli#BeforeFunc
func (cmd *TrafficStepCommand) Before(ctx *cli.Context) error {
	// Update the context
	cmd.ctx = ctx
	return nil
}

// OnUsageError is executed if an usage error occurs.
func (cmd *TrafficStepCommand) OnUsageError(ctx *cli.Context, err error, isSubcommand bool) error {
	cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
	return nil
}

// Action runs when no subcommands are specified
// https://godoc.org/github.com/urfave/cli#ActionFunc
func (cmd *TrafficStepCommand) Action(ctx *cli.Context) error {
	controller, err := NewController(ctx)
	if err != nil {
		// Exit if the controller returned an error
		return nil
	}
	// Update the controller
	cmd.controller = controller

	if ctx.IsSet("service") {
		var amount float64
		if ctx.IsSet("amount") {
			if ctx.Int("amount") < 0 || ctx.Int("amount") > 100 {
				fmt.Fprintf(ctx.App.Writer, "%s\n\n", "--amount must be between 0 and 100")
				return nil
			}
			amount = float64(ctx.Int("amount"))
		}
		return cmd.StepTraffic(ctx.String("service"), amount)
	}

	if ctx.NArg() > 0 {
		cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
		return nil
	}

	return cmd.DefaultAction(ctx)
}

// StepTraffic .
func (cmd *TrafficStepCommand) StepTraffic(serviceName string, amount float64) error {

	filter := &api.RuleFilter{
		Destinations: []string{serviceName},
	}

	routes, err := cmd.controller.ListRoutes(filter)
	if err != nil {
		return err
	}

	routingRules := routes.Services[serviceName]
	if len(routingRules) == 0 {
		fmt.Fprintf(cmd.ctx.App.Writer, "%s: %q\n\n", common.ErrNotRulesFoundForService.Error(), serviceName)
		return nil
	}

	if len(routingRules) > 1 || len(routingRules[0].Route.Backends) != 2 || routingRules[0].Route.Backends[0].Weight == routingRules[0].Route.Backends[1].Weight {
		fmt.Fprintf(cmd.ctx.App.Writer, "%s\n\n", common.ErrInvalidStateForTrafficStep.Error())
		return nil
	}

	rule := routingRules[0]

	// Sort backends by weight, make sure default is last in the slice
	if rule.Route.Backends[0].Weight == 0 {
		rule.Route.Backends[0], rule.Route.Backends[1] = rule.Route.Backends[1], rule.Route.Backends[0]
	}

	currentWeight := rule.Route.Backends[0].Weight
	trafficVersion := strings.Join(rule.Route.Backends[0].Tags, ", ")
	defaultVersion := strings.Join(rule.Route.Backends[1].Tags, ", ")

	if amount == 0 {
		amount = (currentWeight * 100) + 10
	}

	if amount < 100 {
		rule.Route.Backends[0].Weight = amount / 100
	} else {
		amount = 100
		rule.Route.Backends[1].Tags = rule.Route.Backends[0].Tags
		rule.Route.Backends = []api.Backend{rule.Route.Backends[1]}
	}

	rulesSet := &api.RulesSet{
		Rules: []api.Rule{
			rule,
		},
	}

	_, err = cmd.controller.UpdateRules(rulesSet)
	if err != nil {
		return err
	}
	if amount == 100 {
		fmt.Fprintf(cmd.ctx.App.Writer, "Transfer complete for %q: sending %d%% of traffic to %q\n\n", serviceName, int(amount), trafficVersion)
		return nil
	}

	fmt.Fprintf(cmd.ctx.App.Writer, "Transfer starting for %q: diverting %d%% of traffic from %q to %q\n\n", serviceName, int(amount), defaultVersion, trafficVersion)
	return nil
}

// DefaultAction runs the default action.
func (cmd *TrafficStepCommand) DefaultAction(ctx *cli.Context) error {
	return cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
}
