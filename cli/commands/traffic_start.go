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
	reg "github.com/amalgam8/amalgam8/registry/client"
	"github.com/urfave/cli"
)

// TrafficStartCommand is used for the route-list command.
type TrafficStartCommand struct {
	ctx        *cli.Context
	registry   *reg.Client
	controller *ctrl.Client
	term       terminal.UI
}

// NewTrafficStartCommand constructs a new TrafficStart.
func NewTrafficStartCommand(term terminal.UI) (cmd *TrafficStartCommand) {
	return &TrafficStartCommand{
		term: term,
	}
}

// GetMetadata returns the metadata.
func (cmd *TrafficStartCommand) GetMetadata() cli.Command {
	T := utils.Language(common.DefaultLanguage)
	return cli.Command{
		Name:        "traffic-start",
		Description: T("traffic_start_description"),
		Usage:       T("traffic_start_usage"),
		// TODO: Complete UsageText
		UsageText: T("traffic_start_usage"),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "service, s",
				Usage: T("traffic_start_service_usage"),
				Value: "",
			},
			cli.StringFlag{
				Name:  "version, v",
				Usage: T("traffic_start_version_usage"),
				Value: "",
			},
			cli.IntFlag{
				Name:  "amount, a",
				Usage: T("traffic_start_amount_usage"),
			},
		},
		Before:       cmd.Before,
		OnUsageError: cmd.OnUsageError,
		Action:       cmd.Action,
	}
}

// Before runs before the Action
// https://godoc.org/github.com/urfave/cli#BeforeFunc
func (cmd *TrafficStartCommand) Before(ctx *cli.Context) error {
	// Update the context
	cmd.ctx = ctx
	return nil
}

// OnUsageError is executed if an usage error occurs.
func (cmd *TrafficStartCommand) OnUsageError(ctx *cli.Context, err error, isSubcommand bool) error {
	cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
	return nil
}

// Action runs when no subcommands are specified
// https://godoc.org/github.com/urfave/cli#ActionFunc
func (cmd *TrafficStartCommand) Action(ctx *cli.Context) error {
	registry, err := NewRegistry(ctx)
	if err != nil {
		// Exit if the registry returned an error
		return nil
	}
	// Update the registry
	cmd.registry = registry

	controller, err := NewController(ctx)
	if err != nil {
		// Exit if the controller returned an error
		return nil
	}
	// Update the controller
	cmd.controller = controller

	if ctx.IsSet("service") && ctx.IsSet("version") {
		if ctx.Int("amount") < 0 || ctx.Int("amount") > 100 {
			fmt.Fprintf(ctx.App.Writer, "%s\n\n", common.ErrIncorrectAmountRange.Error())
			return nil
		}

		return cmd.StartTraffic(ctx.String("service"), ctx.String("version"), ctx.Int("amount"))

	}

	if ctx.NArg() > 0 {
		cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
		return nil
	}

	return cmd.DefaultAction(ctx)
}

// StartTraffic .
func (cmd *TrafficStartCommand) StartTraffic(serviceName string, version string, amount int) error {

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

	for _, rule := range routingRules {
		if len(rule.Route.Backends) > 1 {
			fmt.Fprintf(cmd.ctx.App.Writer, "%s: service %q traffic is already being split\n\n", common.ErrInvalidStateForTrafficStart.Error(), serviceName)
			return nil
		}
	}

	rule := routingRules[0]
	for _, backend := range rule.Route.Backends {
		if backend.Weight != 0 {
			fmt.Fprintf(cmd.ctx.App.Writer, "%s: service %q traffic is already being split\n\n", common.ErrInvalidStateForTrafficStart.Error(), serviceName)
			return nil
		}
	}

	defaultVersion := strings.Join(rule.Route.Backends[0].Tags, ", ")

	instances, err := cmd.registry.ListServiceInstances(serviceName)
	if err != nil {
		return err
	}

	if !cmd.IsServiceActive(serviceName, defaultVersion, instances) {
		fmt.Fprintf(cmd.ctx.App.Writer, "%s: service %q is not currently receiving traffic\n\n", common.ErrInvalidStateForTrafficStart.Error(), serviceName)
		return nil
	}

	if !cmd.IsServiceActive(serviceName, version, instances) {
		fmt.Fprintf(cmd.ctx.App.Writer, "%s: service %q does not have active instances of version %q\n\n", common.ErrInvalidStateForTrafficStart.Error(), serviceName, version)
		return nil
	}

	if amount == 100 {
		rule.Route.Backends[0].Tags = []string{version}
	} else {
		if amount == 0 {
			amount += 10
		}

		rule.Route.Backends = append([]api.Backend{api.Backend{
			Tags:   []string{version},
			Weight: float64(amount) / 100,
		}}, rule.Route.Backends...)
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
		fmt.Fprintf(cmd.ctx.App.Writer, "Transfer complete for %q: sending %d%% of traffic to %q\n\n", serviceName, amount, version)
		return nil
	}

	fmt.Fprintf(cmd.ctx.App.Writer, "Transfer starting for %q: diverting %d%% of traffic from %q to %q\n\n", serviceName, amount, defaultVersion, version)
	return nil
}

// IsServiceActive .
func (cmd *TrafficStartCommand) IsServiceActive(serviceName, version string, instances []*api.ServiceInstance) bool {
	for _, instance := range instances {
		if version == strings.Join(instance.Tags, ", ") {
			return true
		}
	}
	return false
}

// DefaultAction runs the default action.
func (cmd *TrafficStartCommand) DefaultAction(ctx *cli.Context) error {
	return cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
}
